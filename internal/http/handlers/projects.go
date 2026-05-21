package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/ebachmann/opencode-webchat/internal/auth"
	"github.com/ebachmann/opencode-webchat/internal/store"
	"github.com/go-chi/chi/v5"
)

type ProjectsHandler struct {
	store *store.Store
}

func NewProjectsHandler(s *store.Store) *ProjectsHandler {
	return &ProjectsHandler{store: s}
}

func (h *ProjectsHandler) List(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	projects, err := h.store.ListProjects(r.Context(), user.ID)
	if err != nil {
		log.Printf("failed to list projects: %v", err)
		projects = []store.Project{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

type createProjectRequest struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Repo   string `json:"repo"`
	Branch string `json:"branch"`
}

func (h *ProjectsHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Path == "" {
		http.Error(w, "name and path required", http.StatusBadRequest)
		return
	}

	branch := req.Branch
	if branch == "" {
		branch = "main"
	}

	userID := user.ID
	project, err := h.store.CreateProject(r.Context(), &userID, req.Name, req.Path, req.Repo, branch)
	if err != nil {
		log.Printf("failed to create project: %v", err)
		http.Error(w, "failed to create project", http.StatusInternalServerError)
		return
	}

	if req.Repo != "" {
		go cloneRepo(req.Repo, project.RootPath, branch, user.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func (h *ProjectsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteProject(r.Context(), id); err != nil {
		http.Error(w, "failed to delete project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func cloneRepo(repoURL, targetPath, branch string, userID int) {
	log.Printf("cloning %s (branch=%s) into %s", repoURL, branch, targetPath)

	parentDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		log.Printf("failed to create directory %s: %v", parentDir, err)
		return
	}

	args := []string{"clone", "--branch", branch, repoURL, targetPath}
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("clone failed for user %d: %v\n%s", userID, err, string(output))
		return
	}
	log.Printf("clone succeeded for user %d: %s", userID, targetPath)
	fmt.Sprintf("%s", output)
}