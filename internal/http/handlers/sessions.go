package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ebachmann/opencode-webchat/internal/store"
	"github.com/go-chi/chi/v5"
)

type SessionsHandler struct {
	store *store.Store
}

func NewSessionsHandler(s *store.Store) *SessionsHandler {
	return &SessionsHandler{store: s}
}

type CreateSessionRequest struct {
	ProjectID int    `json:"project_id"`
	Title     string `json:"title"`
}

func (h *SessionsHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, _ := strconv.Atoi(r.URL.Query().Get("project_id"))
	sessions, err := h.store.GetSessionsByProject(r.Context(), projectID)
	if err != nil {
		http.Error(w, "failed to get sessions", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(sessions)
}

func (h *SessionsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		req.Title = "New Session"
	}

	session, err := h.store.CreateSession(r.Context(), req.ProjectID, req.Title)
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(session)
}

func (h *SessionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	session, err := h.store.GetSessionByID(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to get session", http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(session)
}

func (h *SessionsHandler) UpdateTitle(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.store.UpdateSessionTitle(r.Context(), id, req.Title); err != nil {
		http.Error(w, "failed to update session", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SessionsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.store.DeleteSession(r.Context(), id); err != nil {
		http.Error(w, "failed to delete session", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}