package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ebachmann/opencode-webchat/internal/auth"
	"github.com/ebachmann/opencode-webchat/internal/store"
)

type FileEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

type FilesHandler struct {
	store *store.Store
}

func NewFilesHandler(s *store.Store) *FilesHandler {
	return &FilesHandler{store: s}
}

func (h *FilesHandler) List(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	projectIDStr := r.URL.Query().Get("project_id")
	if projectIDStr == "" {
		http.Error(w, "project_id required", http.StatusBadRequest)
		return
	}

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		http.Error(w, "invalid project_id", http.StatusBadRequest)
		return
	}

	project, err := h.store.GetProject(r.Context(), projectID)
	if err != nil || project == nil {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	subPath := r.URL.Query().Get("path")
	fullPath := project.RootPath
	if subPath != "" {
		fullPath = filepath.Join(project.RootPath, subPath)
	}

	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(project.RootPath)) {
		http.Error(w, "invalid path", http.StatusForbidden)
		return
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		log.Printf("failed to read dir %s: %v", fullPath, err)
		http.Error(w, "failed to read directory", http.StatusInternalServerError)
		return
	}

	var files []FileEntry
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		relPath := filepath.Join(subPath, name)
		files = append(files, FileEntry{
			Name:  name,
			Path:  relPath,
			IsDir: e.IsDir(),
			Size:  info.Size(),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}
