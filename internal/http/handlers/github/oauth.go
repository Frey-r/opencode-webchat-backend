package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ebachmann/opencode-webchat/internal/auth"
	"github.com/ebachmann/opencode-webchat/internal/config"
	"github.com/ebachmann/opencode-webchat/internal/store"
)

type Handler struct {
	store  *store.Store
	config config.GitHubConfig
}

func NewHandler(s *store.Store, cfg config.GitHubConfig) *Handler {
	return &Handler{store: s, config: cfg}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	u, _ := url.Parse("https://github.com/login/oauth/authorize")
	q := u.Query()
	q.Set("client_id", h.config.ClientID)
	q.Set("redirect_uri", h.config.RedirectURI)
	q.Set("scope", "repo")
	q.Set("state", "github")
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "no code", http.StatusBadRequest)
		return
	}

	token, err := h.exchangeCode(r.Context(), code)
	if err != nil {
		http.Error(w, "failed to exchange code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	settings, _ := h.store.GetUserSettings(r.Context(), user.ID)
	if settings == nil {
		settings = make(map[string]string)
	}
	settings["GITHUB_ACCESS_TOKEN"] = token

	if err := h.store.SaveUserSettings(r.Context(), user.ID, settings); err != nil {
		http.Error(w, "failed to save token", http.StatusInternalServerError)
		return
	}

	redirectURL := h.config.RedirectURI
	if idx := strings.LastIndex(redirectURL, "/api"); idx >= 0 {
		redirectURL = redirectURL[:idx]
	}
	redirectURL += "/projects"
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	settings, _ := h.store.GetUserSettings(r.Context(), user.ID)
	connected := settings["GITHUB_ACCESS_TOKEN"] != ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"connected": connected,
		"login":     settings["GITHUB_LOGIN"],
	})
}

func (h *Handler) Disconnect(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	settings, _ := h.store.GetUserSettings(r.Context(), user.ID)
	delete(settings, "GITHUB_ACCESS_TOKEN")
	delete(settings, "GITHUB_LOGIN")
	h.store.SaveUserSettings(r.Context(), user.ID, settings)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "disconnected",
	})
}

func (h *Handler) exchangeCode(ctx context.Context, code string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	form := url.Values{}
	form.Set("client_id", h.config.ClientID)
	form.Set("client_secret", h.config.ClientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", h.config.RedirectURI)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	json.Unmarshal(body, &result)

	if result.Error != "" {
		return "", fmt.Errorf("github oauth error: %s", result.Error)
	}

	return result.AccessToken, nil
}
