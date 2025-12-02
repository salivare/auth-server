package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"io"
	"net/http"
)

type ctxKeyType[T any] struct{}

func WithValidated[T any](ctx context.Context, v *T) context.Context {
	return context.WithValue(ctx, ctxKeyType[T]{}, v)
}

func GetValidated[T any](r *http.Request) (*T, bool) {
	val, ok := r.Context().Value(ctxKeyType[T]{}).(*T)
	return val, ok
}

func safeValidateStruct(v *validator.Validate, s any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("validator panic: %v", r)
		}
	}()
	if err = v.Struct(s); err != nil {
		return err
	}
	return nil
}

func BindAndValidateMiddleware[T any](v *validator.Validate) func(http.Handler) http.Handler {
	const maxBody = 1 << 60

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
					return
				}

				r.Body = http.MaxBytesReader(w, r.Body, maxBody)
				instance := new(T)
				dec := json.NewDecoder(r.Body)
				dec.DisallowUnknownFields()

				if err := dec.Decode(instance); err != nil {
					if err == io.EOF {
						jsonError(w, http.StatusBadRequest, "invalid body")
						return
					}
					jsonError(w, http.StatusBadRequest, "invalid body")
					return
				}

				if dec.More() {
					jsonError(w, http.StatusBadRequest, "invalid body")
					return
				}

				if v != nil {
					if err := safeValidateStruct(v, instance); err != nil {
						var validationErrors validator.ValidationErrors
						if errors.As(err, &validationErrors) {
							jsonError(w, http.StatusBadRequest, "validation failed: "+err.Error())
							return
						}
						// внутренняя ошибка/паника валидатора — 500
						jsonError(w, http.StatusInternalServerError, "internal validation error")
						return
					}
				}

				ctx := WithValidated(r.Context(), instance)
				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}
