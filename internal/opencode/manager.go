package opencode

import (
	"context"
	"errors"
	"log"
	"sync"
)

type Manager struct {
	binaryPath string
	processes  map[int]*Process
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewManager(binaryPath string) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		binaryPath: binaryPath,
		processes:  make(map[int]*Process),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (m *Manager) StartProcess(sessionID int) (*Process, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.processes[sessionID]; exists {
		return nil, errors.New("session already exists")
	}

	proc, err := StartProcess(m.ctx, m.binaryPath)
	if err != nil {
		return nil, err
	}

	m.processes[sessionID] = proc
	log.Printf("opencode process started for session %d", sessionID)
	return proc, nil
}

func (m *Manager) GetProcess(sessionID int) *Process {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.processes[sessionID]
}

func (m *Manager) StopProcess(sessionID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	proc, exists := m.processes[sessionID]
	if !exists {
		return nil
	}

	proc.Kill()
	delete(m.processes, sessionID)
	log.Printf("opencode process stopped for session %d", sessionID)
	return nil
}

func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for sessionID, proc := range m.processes {
		proc.Kill()
		delete(m.processes, sessionID)
	}
	m.cancel()
}