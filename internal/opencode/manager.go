package opencode

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type Manager struct {
	binaryPath string
	sessions   map[int]*SessionInfo
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

type SessionInfo struct {
	OpenCodeSessionID string
	Env               map[string]string
}

func NewManager(binaryPath string) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		binaryPath: binaryPath,
		sessions:   make(map[int]*SessionInfo),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (m *Manager) GetProcessInfo(sessionID int) *SessionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[sessionID]
}

func (m *Manager) RegisterSession(sessionID int, env map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[sessionID] = &SessionInfo{
		Env: env,
	}
}

func (m *Manager) UpdateOpenCodeSessionID(sessionID int, ocSessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if info, ok := m.sessions[sessionID]; ok {
		info.OpenCodeSessionID = ocSessionID
	}
}

func (m *Manager) StopProcess(sessionID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
	return nil
}

func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions = make(map[int]*SessionInfo)
	m.cancel()
}

type ProviderModels struct {
	Provider string   `json:"provider"`
	Models   []string `json:"models"`
}

func (m *Manager) GetAvailableModels(ctx context.Context, env map[string]string, configuredProviders map[string]bool) ([]ProviderModels, error) {
	cmd := exec.CommandContext(ctx, m.binaryPath, "models")
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run opencode models: %w", err)
	}

	return ParseModels(string(output), configuredProviders), nil
}

func ParseModels(output string, configuredProviders map[string]bool) []ProviderModels {
	
	groups := make(map[string][]string)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "/") {
			continue
		}
		parts := strings.SplitN(line, "/", 2)
		if len(parts) != 2 {
			continue
		}
		providerKey := strings.ToLower(parts[0])
		model := parts[1]
		if len(configuredProviders) > 0 && !configuredProviders[providerKey] {
			continue
		}
		displayName := FormatProviderName(parts[0])
		groups[displayName] = append(groups[displayName], model)
	}

	var result []ProviderModels
	for name, models := range groups {
		result = append(result, ProviderModels{
			Provider: name,
			Models:   models,
		})
	}
	return result
}

func FormatProviderName(provider string) string {
	switch strings.ToLower(provider) {
	case "opencode":
		return "OpenCode"
	case "opencode-go":
		return "OpenCode Go"
	case "opencode-zen":
		return "OpenCode Zen"
	case "openai":
		return "OpenAI"
	case "anthropic":
		return "Anthropic"
	case "gemini", "google":
		return "Gemini"
	default:
		return strings.ToUpper(provider[:1]) + provider[1:]
	}
}

type RunResult struct {
	EventType string
	Content   string
	Data      json.RawMessage
}

func (m *Manager) RunPrompt(ctx context.Context, prompt, model, sessionID, workDir string, env map[string]string, onEvent func(RunResult)) error {
	args := []string{"run", "--format", "json"}
	if model != "" {
		args = append(args, "-m", model)
	}
	if sessionID != "" {
		args = append(args, "--session", sessionID)
	}
	args = append(args, prompt)

	cmd := exec.CommandContext(ctx, m.binaryPath, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start opencode run: %w", err)
	}

	stderrBuf := &strings.Builder{}
	go func() {
		scanner := bufio.NewScanner(stderr)
		scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
		for scanner.Scan() {
			text := scanner.Text()
			log.Printf("opencode stderr: %s", text)
			stderrBuf.WriteString(text)
			stderrBuf.WriteString("\n")
		}
	}()

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	stdoutLines := &strings.Builder{}
	for scanner.Scan() {
		line := scanner.Text()
		stdoutLines.WriteString(line)
		stdoutLines.WriteString("\n")
		if line == "" {
			continue
		}

		var raw struct {
			Type    string          `json:"type"`
			Content string          `json:"content,omitempty"`
			Text    string          `json:"text,omitempty"`
			Data    json.RawMessage `json:"data,omitempty"`
			Part    struct {
				Text string `json:"text,omitempty"`
			} `json:"part,omitempty"`
			Error struct {
				Data struct {
					Message string `json:"message,omitempty"`
				} `json:"data,omitempty"`
			} `json:"error,omitempty"`
		}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			log.Printf("opencode stdout: parsed as non-JSON: %s", line)
			onEvent(RunResult{
				EventType: "token",
				Content:   line,
			})
			continue
		}

		content := raw.Content
		if content == "" {
			content = raw.Text
		}
		if content == "" {
			content = raw.Part.Text
		}
		if content == "" && raw.Error.Data.Message != "" {
			content = raw.Error.Data.Message
		}

		log.Printf("opencode stdout: type=%s content=%s", raw.Type, content)
		onEvent(RunResult{
			EventType: raw.Type,
			Content:   content,
			Data:      raw.Data,
		})
	}

	log.Printf("opencode stdout buffer:\n%s", stdoutLines.String())

	if err := cmd.Wait(); err != nil {
		log.Printf("opencode run finished with error: %v\n--- stderr ---\n%s--- end stderr ---", err, stderrBuf.String())
	}

	return nil
}