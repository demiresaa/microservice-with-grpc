package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/auth/dto"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/auth/usecase"
)

type AuthHandler struct {
	usecase usecase.AuthUseCase
}

func NewAuthHandler(uc usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{usecase: uc}
}

func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/login", h.Login)
	r.Post("/auth/refresh", h.RefreshToken)
}

// Login godoc
// @Summary      Giris yap
// @Description  Kullanici girisi yapar ve JWT token alir
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Giris bilgileri"
// @Success      200 {object} dto.LoginResponse
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "username and password are required",
		})
		return
	}

	resp, err := h.usecase.Login(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// RefreshToken godoc
// @Summary      Token yenile
// @Description  Refresh token ile yeni access token alir
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RefreshRequest true "Refresh token"
// @Success      200 {object} dto.RefreshResponse
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	if req.RefreshToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "refresh_token is required",
		})
		return
	}

	resp, err := h.usecase.RefreshToken(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
