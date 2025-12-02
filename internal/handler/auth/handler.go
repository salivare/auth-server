package auth

import (
	"encoding/json"
	"errors"
	"github.com/salivare/auth-server/internal/service"
	"net/http"
)

type Handler struct {
	svc *service.AuthService
}

func NewHandler(svc *service.AuthService) *Handler {
	return &Handler{svc: svc}
}

func jsonError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func jsonResponse(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	req, ok := GetValidated[RegisterRequest](r)
	if !ok {
		jsonError(w, http.StatusBadRequest, "Invalid body")
		return
	}

	accessToken, refreshToken, err := h.svc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			jsonError(w, http.StatusConflict, err.Error())
			return
		}
		jsonError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	resp := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	jsonResponse(w, http.StatusCreated, resp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	req, ok := GetValidated[LoginRequest](r)
	if !ok || req == nil {
		jsonError(w, http.StatusBadRequest, "Invalid body")
		return
	}
	accessToken, refreshToken, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrBadCreds) || errors.Is(err, service.ErrNotFound) {
			jsonError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		jsonError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	resp := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	req, ok := GetValidated[RefreshRequest](r)

	if !ok || req == nil {
		jsonError(w, http.StatusBadRequest, "Invalid body")
		return
	}

	access, newRefresh, err := h.svc.RefreshUserToken(r.Context(), req.RefreshToken)

	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	resp := LoginResponse{
		AccessToken:  access,
		RefreshToken: newRefresh,
	}

	jsonResponse(w, http.StatusOK, resp)
}
