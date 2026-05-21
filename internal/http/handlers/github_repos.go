package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ebachmann/opencode-webchat/internal/auth"
)

type GitHubRepo struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	CloneURL    string `json:"clone_url"`
	HTMLURL     string `json:"html_url"`
	Private     bool   `json:"private"`
	UpdatedAt   string `json:"updated_at"`
}

func (h *ProjectsHandler) ListGitHubRepos(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	settings, _ := h.store.GetUserSettings(r.Context(), user.ID)
	token := settings["GITHUB_ACCESS_TOKEN"]
	if token == "" {
		http.Error(w, "github not connected", http.StatusUnauthorized)
		return
	}

	repos, err := fetchGitHubRepos(r.Context(), token)
	if err != nil {
		http.Error(w, "failed to fetch repos: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repos)
}

func fetchGitHubRepos(ctx context.Context, token string) ([]GitHubRepo, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/repos?sort=updated&per_page=100&affiliation=owner,collaborator", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api error %d: %s", resp.StatusCode, string(body))
	}

	var repos []GitHubRepo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	return repos, nil
}
