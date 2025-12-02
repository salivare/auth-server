package auth

import (
	"github.com/go-playground/validator/v10"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, h *Handler, v *validator.Validate) {
	registerBind := BindAndValidateMiddleware[RegisterRequest](v)
	loginBind := BindAndValidateMiddleware[LoginRequest](v)
	refreshBind := BindAndValidateMiddleware[RefreshRequest](v)

	mux.Handle("/auth/register", registerBind(http.HandlerFunc(h.Register)))
	mux.Handle("/auth/login", loginBind(http.HandlerFunc(h.Login)))
	mux.Handle("/auth/refresh", refreshBind(http.HandlerFunc(h.Refresh)))
}
