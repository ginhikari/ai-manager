package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	General    GeneralConfig    `yaml:"general"`
	Processes  []ProcessConfig  `yaml:"processes"`
	Monitor    MonitorConfig    `yaml:"monitor"`
}

type GeneralConfig struct {
	LogDir   string `yaml:"log_dir"`
	DataDir  string `yaml:"data_dir"`
	AutoSave bool   `yaml:"auto_save"`
}

type ProcessConfig struct {
	ID         string            `yaml:"id"`
	Name       string            `yaml:"name"`
	Type       string            `yaml:"type"`
	Command    string            `yaml:"command"`
	Args       []string          `yaml:"args,omitempty"`
	WorkingDir string            `yaml:"working_dir,omitempty"`
	Env        map[string]string `yaml:"env,omitempty"`
	Port       int               `yaml:"port,omitempty"`
	Host       string            `yaml:"host,omitempty"`
	LogFile    string            `yaml:"log_file,omitempty"`
	Labels     map[string]string `yaml:"labels,omitempty"`
	Notes      string            `yaml:"notes,omitempty"`
	AutoStart  bool              `yaml:"auto_start,omitempty"`
}

type MonitorConfig struct {
	Interval string `yaml:"interval"`
}

type Manager struct {
	configPath string
	appConfig  *AppConfig
}

func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".ai-manager.yaml")
	}
	return filepath.Join(home, ".ai-manager.yaml")
}

func NewManager(configPath string) *Manager {
	return &Manager{
		configPath: configPath,
	}
}

func (m *Manager) Load() (*AppConfig, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			defaultCfg := DefaultConfig()
			if err := m.Save(defaultCfg); err != nil {
				return nil, fmt.Errorf("failed to create default config: %w", err)
			}
			return defaultCfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	m.appConfig = &cfg
	return &cfg, nil
}

func (m *Manager) Save(cfg *AppConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	m.appConfig = cfg
	return nil
}

func (m *Manager) GetConfig() *AppConfig {
	return m.appConfig
}

func (m *Manager) AddProcess(p ProcessConfig) error {
	if m.appConfig == nil {
		return fmt.Errorf("config not loaded")
	}

	for _, existing := range m.appConfig.Processes {
		if existing.ID == p.ID {
			return fmt.Errorf("process with ID %s already exists", p.ID)
		}
	}

	m.appConfig.Processes = append(m.appConfig.Processes, p)
	return m.Save(m.appConfig)
}

func (m *Manager) UpdateProcess(p ProcessConfig) error {
	if m.appConfig == nil {
		return fmt.Errorf("config not loaded")
	}

	for i, existing := range m.appConfig.Processes {
		if existing.ID == p.ID {
			m.appConfig.Processes[i] = p
			return m.Save(m.appConfig)
		}
	}
	return fmt.Errorf("process with ID %s not found", p.ID)
}

func (m *Manager) RemoveProcess(id string) error {
	if m.appConfig == nil {
		return fmt.Errorf("config not loaded")
	}

	for i, p := range m.appConfig.Processes {
		if p.ID == id {
			m.appConfig.Processes = append(m.appConfig.Processes[:i], m.appConfig.Processes[i+1:]...)
			return m.Save(m.appConfig)
		}
	}
	return fmt.Errorf("process with ID %s not found", id)
}

func (m *Manager) GetProcess(id string) (*ProcessConfig, error) {
	if m.appConfig == nil {
		return nil, fmt.Errorf("config not loaded")
	}

	for _, p := range m.appConfig.Processes {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("process with ID %s not found", id)
}

func DefaultConfig() *AppConfig {
	home, _ := os.UserHomeDir()
	return &AppConfig{
		General: GeneralConfig{
			LogDir:  filepath.Join(home, ".ai-manager", "logs"),
			DataDir: filepath.Join(home, ".ai-manager", "data"),
			AutoSave: true,
		},
		Processes: []ProcessConfig{},
		Monitor: MonitorConfig{
			Interval: "10s",
		},
	}
}
