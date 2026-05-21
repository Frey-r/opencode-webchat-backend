package apikey

import (
	"strings"
	"unicode/utf8"
)

type Provider struct {
	Key         string `json:"key"`
	DisplayName string `json:"display_name"`
	EnvVar      string `json:"env_var"`
	Placeholder string `json:"placeholder"`
	Prefix      string `json:"prefix"`
	Required    bool   `json:"required"`
	Category    string `json:"category"`
}

var providers = []Provider{
	{
		Key:         "OPENAI_API_KEY",
		DisplayName: "OpenAI",
		EnvVar:      "OPENAI_API_KEY",
		Placeholder: "sk-...",
		Prefix:      "sk-",
		Category:    "provider",
	},
	{
		Key:         "ANTHROPIC_API_KEY",
		DisplayName: "Anthropic",
		EnvVar:      "ANTHROPIC_API_KEY",
		Placeholder: "sk-ant-...",
		Prefix:      "sk-ant-",
		Category:    "provider",
	},
	{
		Key:         "GEMINI_API_KEY",
		DisplayName: "Google Gemini",
		EnvVar:      "GEMINI_API_KEY",
		Placeholder: "AIza...",
		Prefix:      "AIza",
		Category:    "provider",
	},
	{
		Key:         "OPENCODE_API_KEY",
		DisplayName: "OpenCode (Go & Zen)",
		EnvVar:      "OPENCODE_API_KEY",
		Placeholder: "sk-...",
		Prefix:      "sk-",
		Category:    "opencode",
	},
}

var nonKeySettings = []string{
	"OPENCODE_MODEL",
	"OPENCODE_SERVER_PASSWORD",
}

var sensitiveSettings = []string{
	"GITHUB_ACCESS_TOKEN",
	"GITHUB_LOGIN",
}

func Providers() []Provider {
	return providers
}

func MaskKey(key string) string {
	if key == "" {
		return ""
	}
	length := utf8.RuneCountInString(key)
	if length <= 8 {
		return "****"
	}
	prefixRunes := []rune(key)[:4]
	suffixRunes := []rune(key)[length-4:]
	return string(prefixRunes) + strings.Repeat("*", length-8) + string(suffixRunes)
}

func ValidateKey(providerKey, value string) error {
	if value == "" {
		return nil
	}
	for _, p := range providers {
		if p.Key == providerKey && p.Prefix != "" {
			if !strings.HasPrefix(value, p.Prefix) {
				return &ValidationError{
					ProviderKey: providerKey,
					Message:     "key must start with " + p.Prefix,
				}
			}
		}
	}
	return nil
}

type ValidationError struct {
	ProviderKey string
	Message     string
}

func (e *ValidationError) Error() string {
	return e.ProviderKey + ": " + e.Message
}

func IsAPIKey(key string) bool {
	for _, p := range providers {
		if p.Key == key {
			return true
		}
	}
	return false
}

func IsSensitive(key string) bool {
	for _, s := range sensitiveSettings {
		if s == key {
			return true
		}
	}
	return false
}

func IsKnownSetting(key string) bool {
	if IsAPIKey(key) {
		return true
	}
	for _, s := range nonKeySettings {
		if s == key {
			return true
		}
	}
	return false
}

func BuildEnv(settings map[string]string) map[string]string {
	env := make(map[string]string)
	for _, p := range providers {
		if val, ok := settings[p.Key]; ok && val != "" {
			env[p.EnvVar] = val
		}
	}
	if val, ok := settings["OPENCODE_MODEL"]; ok && val != "" {
		env["OPENCODE_MODEL"] = val
	}
	if val, ok := settings["OPENCODE_SERVER_PASSWORD"]; ok && val != "" {
		env["OPENCODE_SERVER_PASSWORD"] = val
	}
	return env
}

func MaskSettings(settings map[string]string) map[string]string {
	masked := make(map[string]string, len(settings))
	for k, v := range settings {
		if IsAPIKey(k) || IsSensitive(k) {
			masked[k] = MaskKey(v)
		} else {
			masked[k] = v
		}
	}
	return masked
}

type ProviderStatus struct {
	Key         string `json:"key"`
	DisplayName string `json:"display_name"`
	Configured  bool   `json:"configured"`
	Category    string `json:"category"`
}

var settingToProviderNames = map[string][]string{
	"OPENAI_API_KEY":    {"openai"},
	"ANTHROPIC_API_KEY": {"anthropic"},
	"GEMINI_API_KEY":    {"gemini", "google"},
	"OPENCODE_API_KEY":  {"opencode", "opencode-go", "opencode-zen"},
}

func ConfiguredProviderNames(settings map[string]string) map[string]bool {
	names := make(map[string]bool)
	for k, pNames := range settingToProviderNames {
		if settings[k] != "" {
			for _, n := range pNames {
				names[n] = true
			}
		}
	}
	return names
}

func GetProviderStatuses(settings map[string]string) []ProviderStatus {
	statuses := make([]ProviderStatus, 0, len(providers))
	for _, p := range providers {
		statuses = append(statuses, ProviderStatus{
			Key:         p.Key,
			DisplayName: p.DisplayName,
			Configured:  settings[p.Key] != "",
			Category:    p.Category,
		})
	}
	return statuses
}
