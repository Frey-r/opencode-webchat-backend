package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ebachmann/opencode-webchat/internal/auth"
	"github.com/ebachmann/opencode-webchat/internal/store"
)

type AuthHandler struct {
	jwtManager *auth.JWTManager
	store      *store.Store
}

func NewAuthHandler(jm *auth.JWTManager, s *store.Store) *AuthHandler {
	return &AuthHandler{jwtManager: jm, store: s}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.store.GetUserByUsername(r.Context(), req.Username)
	if err != nil || user == nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.jwtManager.Generate(user.ID, user.Username)
	if err != nil {
		log.Printf("failed to generate token: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.jwtManager.CookieName(),
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	json.NewEncoder(w).Encode(map[string]any{"username": user.Username})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.jwtManager.CookieName(),
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"username": user.Username})
}