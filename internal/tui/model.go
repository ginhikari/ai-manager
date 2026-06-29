package tui

import (
	"time"

	"ai-manager/internal/config"
	"ai-manager/internal/logger"
	"ai-manager/internal/monitor"
	"ai-manager/internal/ollama"
	"ai-manager/internal/process"
	"ai-manager/internal/state"
)

type AppContext struct {
	ConfigMgr       *config.Manager
	ProcMgr         *process.ProcessManager
	Monitor         *monitor.Monitor
	LogMgr          *logger.LogManager
	StateMgr        *state.Manager
	Ollama          *ollama.Client
	ModelDir        string
	StartTime       time.Time
	LastRefresh     time.Time
	Connection      string
	RefreshInterval time.Duration
}
