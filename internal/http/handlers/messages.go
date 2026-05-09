package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ebachmann/opencode-webchat/internal/store"
	"github.com/go-chi/chi/v5"
)

type MessagesHandler struct {
	store *store.Store
}

func NewMessagesHandler(s *store.Store) *MessagesHandler {
	return &MessagesHandler{store: s}
}

func (h *MessagesHandler) List(w http.ResponseWriter, r *http.Request) {
	sessionID, _ := strconv.Atoi(chi.URLParam(r, "sessionId"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 100
	}

	messages, err := h.store.GetMessagesBySession(r.Context(), sessionID, limit, offset)
	if err != nil {
		http.Error(w, "failed to get messages", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(messages)
}