package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ebachmann/opencode-webchat/internal/store"
)

type ProjectsHandler struct {
	store *store.Store
}

func NewProjectsHandler(s *store.Store) *ProjectsHandler {
	return &ProjectsHandler{store: s}
}

type Project struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	RootPath   string `json:"root_path"`
	LastUsedAt string `json:"last_used_at"`
}

func (h *ProjectsHandler) List(w http.ResponseWriter, r *http.Request) {
	projects := []Project{
		{ID: 1, Name: "Default Project", RootPath: "/home/pi/projects"},
	}
	json.NewEncoder(w).Encode(projects)
}