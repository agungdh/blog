package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"blog/internal/model"
	"blog/internal/service"
)

type contextKey string

const userContextKey contextKey = "user"

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
	Nama  string `json:"nama"`
}

type meResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nama     string `json:"nama"`
}

// Login handles user authentication and returns a bearer token.
//
// @Summary Login
// @Description Authenticate with username and password to get a bearer token
// @Tags admin
// @Accept json
// @Produce json
// @Param request body loginRequest true "Credentials"
// @Success 200 {object} loginResponse "Login successful"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Router /api/admin/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Username == "" || req.Password == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "username and password are required"})
		return
	}

	token, nama, err := h.auth.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid username or password"})
		return
	}

	respondJSON(w, http.StatusOK, loginResponse{Token: token, Nama: nama})
}

// Me returns the authenticated user's profile.
//
// @Summary Get current user
// @Description Returns the currently authenticated user's information
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} meResponse "User profile"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/admin/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	respondJSON(w, http.StatusOK, meResponse{
		ID:       user.ID,
		Username: user.Username,
		Nama:     user.Nama,
	})
}

// Logout invalidates the current session token.
//
// @Summary Logout
// @Description Invalidates the current bearer token
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "Logged out"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/admin/logout [delete]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if err := h.auth.Logout(r.Context(), token); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to logout"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}

		user, err := h.auth.ValidateToken(r.Context(), token)
		if err != nil {
			respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func GetUserFromContext(ctx context.Context) *model.User {
	user, ok := ctx.Value(userContextKey).(*model.User)
	if !ok {
		return nil
	}
	return user
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
