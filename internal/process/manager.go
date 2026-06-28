package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type InstanceType string

const (
	InstanceTypeModel    InstanceType = "model"
	InstanceTypeAPI      InstanceType = "api"
	InstanceTypeTraining InstanceType = "training"
)

type ProcessState string

const (
	StateStopped ProcessState = "stopped"
	StateRunning ProcessState = "running"
	StatePaused  ProcessState = "paused"
	StateError   ProcessState = "error"
)

type Process struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Type       InstanceType   `json:"type"`
	Command    string         `json:"command"`
	Args       []string       `json:"args,omitempty"`
	WorkingDir string         `json:"working_dir,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	State      ProcessState   `json:"state"`
	PID        int            `json:"pid,omitempty"`
	Port       int            `json:"port,omitempty"`
	Host       string         `json:"host,omitempty"`
	LogFile    string         `json:"log_file,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	StartedAt  *time.Time     `json:"started_at,omitempty"`
	StoppedAt  *time.Time     `json:"stopped_at,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Notes      string         `json:"notes,omitempty"`
}

type ProcessManager struct {
	mu        sync.RWMutex
	procs     map[string]*Process
	runner    ProcessRunner
	logDir    string
	stateSaver StateSaver
}

type ProcessRunner interface {
	Start(ctx context.Context, p *Process) (*exec.Cmd, error)
	Stop(p *Process) error
	Pause(p *Process) error
	Resume(p *Process) error
	IsRunning(p *Process) bool
}

type StateSaver interface {
	SaveProcess(p *Process) error
	SaveAll(procs map[string]*Process) error
}

type NoopStateSaver struct{}

func (n *NoopStateSaver) SaveProcess(p *Process) error  { return nil }
func (n *NoopStateSaver) SaveAll(procs map[string]*Process) error { return nil }

type OSProcessRunner struct{}

func (r *OSProcessRunner) Start(ctx context.Context, p *Process) (*exec.Cmd, error) {
	var cmd *exec.Cmd
	if len(p.Args) > 0 {
		cmd = exec.CommandContext(ctx, p.Command, p.Args...)
	} else {
		cmd = exec.CommandContext(ctx, p.Command)
	}

	if p.WorkingDir != "" {
		cmd.Dir = p.WorkingDir
	}

	if p.LogFile != "" {
		f, err := os.OpenFile(p.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		cmd.Stdout = f
		cmd.Stderr = f
	}

	if len(p.Env) > 0 {
		cmd.Env = append(os.Environ(), envMapToSlice(p.Env)...)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	now := time.Now()
	p.PID = cmd.Process.Pid
	p.State = StateRunning
	p.StartedAt = &now

	go func() {
		_ = cmd.Wait()
		if p.State == StateRunning {
			p.State = StateStopped
			now := time.Now()
			p.StoppedAt = &now
		}
	}()

	return cmd, nil
}

func (r *OSProcessRunner) Stop(p *Process) error {
	if p.PID == 0 {
		return fmt.Errorf("process %s is not running", p.ID)
	}

	cmd := exec.Command("taskkill", "/F", "/PID", fmt.Sprintf("%d", p.PID))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop process: %w", err)
	}

	now := time.Now()
	p.PID = 0
	p.State = StateStopped
	p.StoppedAt = &now
	return nil
}

func (r *OSProcessRunner) Pause(p *Process) error {
	if p.PID == 0 {
		return fmt.Errorf("process %s is not running", p.ID)
	}

	cmd := exec.Command("taskkill", "/F", "/PID", fmt.Sprintf("%d", p.PID))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pause process: %w", err)
	}

	p.State = StatePaused
	return nil
}

func (r *OSProcessRunner) Resume(p *Process) error {
	if p.State != StatePaused {
		return fmt.Errorf("process %s is not paused", p.ID)
	}

	if p.Command == "" {
		return fmt.Errorf("no command to resume for process %s", p.ID)
	}

	var cmd *exec.Cmd
	if len(p.Args) > 0 {
		cmd = exec.Command(p.Command, p.Args...)
	} else {
		cmd = exec.Command(p.Command)
	}

	if p.WorkingDir != "" {
		cmd.Dir = p.WorkingDir
	}

	if p.LogFile != "" {
		f, err := os.OpenFile(p.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		cmd.Stdout = f
		cmd.Stderr = f
	}

	if len(p.Env) > 0 {
		cmd.Env = append(os.Environ(), envMapToSlice(p.Env)...)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to resume process: %w", err)
	}

	now := time.Now()
	p.PID = cmd.Process.Pid
	p.State = StateRunning
	p.StartedAt = &now

	go func() {
		_ = cmd.Wait()
		if p.State == StateRunning {
			p.State = StateStopped
			now := time.Now()
			p.StoppedAt = &now
		}
	}()

	return nil
}

func (r *OSProcessRunner) IsRunning(p *Process) bool {
	if p.PID == 0 {
		return false
	}

	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", p.PID), "/FO", "CSV", "/NH")
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	lines := strings.Split(string(out), "\r\n")
	return len(lines) > 0 && strings.Contains(lines[0], fmt.Sprintf("\"%d\"", p.PID))
}

func NewProcessManager(logDir string, saver StateSaver) *ProcessManager {
	if saver == nil {
		saver = &NoopStateSaver{}
	}
	return &ProcessManager{
		procs:      make(map[string]*Process),
		runner:     &OSProcessRunner{},
		logDir:     logDir,
		stateSaver: saver,
	}
}

func (m *ProcessManager) Add(p *Process) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p.CreatedAt = time.Now()
	p.State = StateStopped
	m.procs[p.ID] = p
}

func (m *ProcessManager) Get(id string) (*Process, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.procs[id]
	return p, ok
}

func (m *ProcessManager) List() []*Process {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Process, 0, len(m.procs))
	for _, p := range m.procs {
		result = append(result, p)
	}
	return result
}

func (m *ProcessManager) ListByType(t InstanceType) []*Process {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*Process
	for _, p := range m.procs {
		if p.Type == t {
			result = append(result, p)
		}
	}
	return result
}

func (m *ProcessManager) ListByState(s ProcessState) []*Process {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*Process
	for _, p := range m.procs {
		if p.State == s {
			result = append(result, p)
		}
	}
	return result
}

func (m *ProcessManager) ListByName(name string) []*Process {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*Process
	for _, p := range m.procs {
		if strings.EqualFold(p.Name, name) {
			result = append(result, p)
		}
	}
	return result
}

func (m *ProcessManager) Delete(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.procs[id]; !ok {
		return false
	}
	delete(m.procs, id)
	return true
}

func (m *ProcessManager) Update(p *Process) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.procs[p.ID]; !ok {
		return false
	}
	m.procs[p.ID] = p
	return true
}

func (m *ProcessManager) Start(id string) error {
	m.mu.RLock()
	p, ok := m.procs[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("process %s not found", id)
	}
	if p.State == StateRunning {
		return fmt.Errorf("process %s is already running", id)
	}
	if p.State == StatePaused {
		err := m.runner.Resume(p)
		if err != nil {
			return err
		}
		return m.stateSaver.SaveProcess(p)
	}
	_, err := m.runner.Start(context.Background(), p)
	if err != nil {
		return err
	}
	return m.stateSaver.SaveProcess(p)
}

func (m *ProcessManager) Stop(id string) error {
	m.mu.RLock()
	p, ok := m.procs[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("process %s not found", id)
	}
	if p.State != StateRunning {
		return fmt.Errorf("process %s is not running", id)
	}
	err := m.runner.Stop(p)
	if err != nil {
		return err
	}
	return m.stateSaver.SaveProcess(p)
}

func (m *ProcessManager) Pause(id string) error {
	m.mu.RLock()
	p, ok := m.procs[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("process %s not found", id)
	}
	if p.State != StateRunning {
		return fmt.Errorf("process %s is not running", id)
	}
	err := m.runner.Pause(p)
	if err != nil {
		return err
	}
	return m.stateSaver.SaveProcess(p)
}

func (m *ProcessManager) Restart(id string) error {
	if err := m.Stop(id); err != nil {
		return err
	}
	return m.Start(id)
}

func (m *ProcessManager) HealthCheck(id string) (bool, error) {
	m.mu.RLock()
	p, ok := m.procs[id]
	m.mu.RUnlock()
	if !ok {
		return false, fmt.Errorf("process %s not found", id)
	}
	if p.State != StateRunning {
		return false, nil
	}
	return m.runner.IsRunning(p), nil
}

func (m *ProcessManager) RestoreState() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, p := range m.procs {
		if p.State == StateStopped || p.PID == 0 {
			continue
		}

		if m.runner.IsRunning(p) {
			p.State = StateRunning
			continue
		}

		p.State = StateStopped
		p.PID = 0
		now := time.Now()
		p.StoppedAt = &now
	}

	return m.stateSaver.SaveAll(m.procs)
}

func envMapToSlice(m map[string]string) []string {
	result := make([]string, 0, len(m))
	for k, v := range m {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}
