package main

import (
	"fmt"
	"os"

	bubbletea "github.com/charmbracelet/bubbletea"

	"ai-manager/internal/cli"
	"ai-manager/internal/config"
	"ai-manager/internal/logger"
	"ai-manager/internal/monitor"
	"ai-manager/internal/ollama"
	"ai-manager/internal/process"
	"ai-manager/internal/state"
	"ai-manager/internal/tui"
	"ai-manager/internal/tui/types"
)

func main() {
	root := cli.NewRootCmd()

	var tuiFlag bool
	root.PersistentFlags().BoolVar(&tuiFlag, "tui", false, "Launch TUI mode")

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if tuiFlag {
		if err := launchTUI(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func launchTUI() error {
	home, _ := os.UserHomeDir()
	configPath := config.DefaultConfigPath()
	statePath := home + "\\AppData\\Local\\ai-manager\\state.json"
	logDir := home + "\\AppData\\Local\\ai-manager\\logs"

	configMgr := config.NewManager(configPath)
	stateMgr := state.NewManager(statePath)
	loggerMgr := logger.NewLogManager(logDir)
	procMgr := process.NewProcessManager(logDir, nil)
	monitor := monitor.NewMonitor(1000000000)

	modelDir := "C:\\Users\\elton\\.ollama\\models"
	ollamaClient := ollama.NewClient("localhost", 11434)

	ctx := &types.AppContext{
		ConfigMgr:       configMgr,
		ProcMgr:         procMgr,
		Monitor:         monitor,
		LogMgr:          loggerMgr,
		StateMgr:        stateMgr,
		Ollama:          ollamaClient,
		ModelDir:        modelDir,
		RefreshInterval: 1000000000,
	}

	app := tui.NewApp(ctx)
	p := bubbletea.NewProgram(app, bubbletea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}
