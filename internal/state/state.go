package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ai-manager/internal/process"
)

type StateFile struct {
	Processes map[string]*ProcessState `json:"processes"`
}

type ProcessState struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Command    string            `json:"command"`
	Args       []string          `json:"args,omitempty"`
	WorkingDir string            `json:"working_dir,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	Port       int               `json:"port,omitempty"`
	Host       string            `json:"host,omitempty"`
	LogFile    string            `json:"log_file,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Notes      string            `json:"notes,omitempty"`
	State      string            `json:"state"`
	PID        int               `json:"pid,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	StartedAt  *time.Time        `json:"started_at,omitempty"`
	StoppedAt  *time.Time        `json:"stopped_at,omitempty"`
}

type Manager struct {
	stateFile string
	state     *StateFile
}

func NewManager(stateFilePath string) *Manager {
	return &Manager{
		stateFile: stateFilePath,
		state:     &StateFile{Processes: make(map[string]*ProcessState)},
	}
}

func (m *Manager) Load() error {
	data, err := os.ReadFile(m.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read state file: %w", err)
	}

	if err := json.Unmarshal(data, m.state); err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}

	return nil
}

func (m *Manager) Save() error {
	dir := filepath.Dir(m.stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := json.MarshalIndent(m.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(m.stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

func (m *Manager) SaveProcess(p *process.Process) error {
	ps := processToState(p)
	m.state.Processes[p.ID] = ps
	return m.Save()
}

func (m *Manager) SaveAll(procs map[string]*process.Process) error {
	for _, p := range procs {
		ps := processToState(p)
		m.state.Processes[p.ID] = ps
	}
	return m.Save()
}

func (m *Manager) GetState(id string) (*ProcessState, bool) {
	ps, ok := m.state.Processes[id]
	return ps, ok
}

func (m *Manager) GetAllStates() map[string]*ProcessState {
	return m.state.Processes
}

func (m *Manager) Remove(id string) bool {
	_, ok := m.state.Processes[id]
	if !ok {
		return false
	}
	delete(m.state.Processes, id)
	return true
}

func processToState(p *process.Process) *ProcessState {
	ps := &ProcessState{
		ID:         p.ID,
		Name:       p.Name,
		Type:       string(p.Type),
		Command:    p.Command,
		Args:       p.Args,
		WorkingDir: p.WorkingDir,
		Env:        p.Env,
		Port:       p.Port,
		Host:       p.Host,
		LogFile:    p.LogFile,
		Labels:     p.Labels,
		Notes:      p.Notes,
		State:      string(p.State),
		PID:        p.PID,
		CreatedAt:  p.CreatedAt,
		StartedAt:  p.StartedAt,
		StoppedAt:  p.StoppedAt,
	}
	return ps
}

func DefaultStatePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".ai-manager-state.json")
	}
	return filepath.Join(home, ".ai-manager-state.json")
}
