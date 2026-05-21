package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/ebachmann/opencode-webchat/internal/apikey"
	"github.com/ebachmann/opencode-webchat/internal/auth"
	opencodepkg "github.com/ebachmann/opencode-webchat/internal/opencode"
	"github.com/ebachmann/opencode-webchat/internal/store"
)

type SettingsHandler struct {
	store       *store.Store
	opencodeMgr *opencodepkg.Manager
}

func NewSettingsHandler(s *store.Store, om *opencodepkg.Manager) *SettingsHandler {
	return &SettingsHandler{store: s, opencodeMgr: om}
}

func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	settings, err := h.store.GetUserSettings(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "failed to get settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apikey.MaskSettings(settings))
}

func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var settings map[string]string
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Check all keys are known settings
	for key := range settings {
		if !apikey.IsKnownSetting(key) {
			http.Error(w, "unknown setting: "+key, http.StatusBadRequest)
			return
		}
	}

	// Restore masked values from DB before validation
	currentSettings, err := h.store.GetUserSettings(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "failed to get current settings", http.StatusInternalServerError)
		return
	}

	for key, value := range settings {
		if apikey.IsAPIKey(key) && value != "" && apikey.MaskKey(currentSettings[key]) == value {
			// Value is the masked version of the current key — restore original
			settings[key] = currentSettings[key]
		}
	}

	// Only validate keys that actually changed (not restored masked values)
	for key, value := range settings {
		if err := apikey.ValidateKey(key, value); err != nil {
			var ve *apikey.ValidationError
			if errors.As(err, &ve) {
				http.Error(w, ve.Error(), http.StatusBadRequest)
				return
			}
		}
	}

	if err := h.store.SaveUserSettings(r.Context(), user.ID, settings); err != nil {
		http.Error(w, "failed to save settings", http.StatusInternalServerError)
		return
	}

	updatedSettings, _ := h.store.GetUserSettings(r.Context(), user.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apikey.MaskSettings(updatedSettings))
}

func (h *SettingsHandler) GetModels(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	settings, _ := h.store.GetUserSettings(r.Context(), user.ID)
	env := apikey.BuildEnv(settings)
	configured := apikey.ConfiguredProviderNames(settings)

	models, err := h.opencodeMgr.GetAvailableModels(r.Context(), env, configured)
	if err != nil {
		log.Printf("failed to get models from opencode: %v", err)
		models = []opencodepkg.ProviderModels{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}

func (h *SettingsHandler) GetProviders(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	settings, _ := h.store.GetUserSettings(r.Context(), user.ID)

	response := struct {
		Providers []apikey.Provider       `json:"providers"`
		Statuses  []apikey.ProviderStatus  `json:"statuses"`
	}{
		Providers: apikey.Providers(),
		Statuses:  apikey.GetProviderStatuses(settings),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}