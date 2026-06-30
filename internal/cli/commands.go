package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"ai-manager/internal/config"
	"ai-manager/internal/logger"
	"ai-manager/internal/monitor"
	"ai-manager/internal/process"
	"ai-manager/internal/state"
	"ai-manager/internal/tui"
	"github.com/spf13/cobra"
)

var (
	cfgPath    string
	cfg        *config.AppConfig
	configMgr  *config.Manager
	stateMgr   *state.Manager
	pManager   *process.ProcessManager
	lManager   *logger.LogManager
	mon        *monitor.Monitor
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "ai-manager",
		Short: "Manage all AI-related processes on your machine",
		Long: `ai-manager is a CLI tool for managing AI inference, API services,
and training jobs. It provides process lifecycle management, resource
monitoring, logging, and configuration all in one place.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initApp()
		},
	}

	root.PersistentFlags().StringVar(&cfgPath, "config", config.DefaultConfigPath(), "path to config file")

	root.AddCommand(
		newProcessCmd(),
		newMonitorCmd(),
		newLogCmd(),
		newConfigCmd(),
		newStatusCmd(),
		newVersionCmd(),
		newTuiCmd(),
	)

	return root
}

func initApp() error {
	configMgr = config.NewManager(cfgPath)
	var err error
	cfg, err = configMgr.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logDir := cfg.General.LogDir
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	dataDir := cfg.General.DataDir
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	stateMgr = state.NewManager(state.DefaultStatePath())
	if err := stateMgr.Load(); err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	pManager = process.NewProcessManager(logDir, stateMgr)
	lManager = logger.NewLogManager(logDir)

	interval := 10 * time.Second
	if cfg.Monitor.Interval != "" {
		d, err := time.ParseDuration(cfg.Monitor.Interval)
		if err == nil {
			interval = d
		}
	}
	mon = monitor.NewMonitor(interval)

	for _, pc := range cfg.Processes {
		p := &process.Process{
			ID:         pc.ID,
			Name:       pc.Name,
			Type:       process.InstanceType(pc.Type),
			Command:    pc.Command,
			Args:       pc.Args,
			WorkingDir: pc.WorkingDir,
			Env:        pc.Env,
			Port:       pc.Port,
			Host:       pc.Host,
			LogFile:    lManager.GetLogPath(pc.ID),
			Labels:     pc.Labels,
			Notes:      pc.Notes,
		}
		pManager.Add(p)
	}

	for id, ps := range stateMgr.GetAllStates() {
		p, ok := pManager.Get(id)
		if !ok {
			continue
		}
		if ps.PID > 0 && ps.State == "running" {
			p.PID = ps.PID
			p.State = process.ProcessState(ps.State)
			p.StartedAt = ps.StartedAt
		}
	}

	if err := pManager.RestoreState(); err != nil {
		return fmt.Errorf("failed to restore state: %w", err)
	}

	for _, pc := range cfg.Processes {
		if !pc.AutoStart {
			continue
		}
		p, ok := pManager.Get(pc.ID)
		if !ok {
			continue
		}
		if p.State == process.StateRunning {
			continue
		}
		if err := pManager.Start(pc.ID); err != nil {
			fmt.Printf("Warning: failed to auto-start %s: %v\n", pc.ID, err)
		} else {
			fmt.Printf("Auto-started process %s\n", pc.ID)
		}
	}

	return nil
}

func newProcessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "process",
		Aliases: []string{"proc", "p"},
		Short:   "Manage AI process instances",
	}

	cmd.AddCommand(
		newProcessListCmd(),
		newProcessStartCmd(),
		newProcessStopCmd(),
		newProcessPauseCmd(),
		newProcessRestartCmd(),
		newProcessAddCmd(),
		newProcessRemoveCmd(),
		newProcessEditCmd(),
		newProcessAutoStartCmd(),
		newProcessHealthCmd(),
	)

	return cmd
}

func newProcessListCmd() *cobra.Command {
	var (
		filterType string
		filterState string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Aliases: []string{"ls"},
		Short: "List all process instances",
		RunE: func(cmd *cobra.Command, args []string) error {
			var procs []*process.Process

			if filterType != "" {
				procs = pManager.ListByType(process.InstanceType(filterType))
			} else if filterState != "" {
				procs = pManager.ListByState(process.ProcessState(filterState))
			} else {
				procs = pManager.List()
			}

			if len(procs) == 0 {
				fmt.Println("No processes found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tTYPE\tSTATE\tPID\tPORT")
			fmt.Fprintln(w, strings.Repeat("-", 60))

			for _, p := range procs {
				pid := ""
				if p.PID > 0 {
					pid = fmt.Sprintf("%d", p.PID)
				}
				port := ""
				if p.Port > 0 {
					port = fmt.Sprintf("%d", p.Port)
				}
				fmt.Fprintf(w, "%-15s %-20s %-10s %-10s %-8s %-8s\n",
					p.ID, p.Name, p.Type, p.State, pid, port)
			}
			w.Flush()
			return nil
		},
	}

	cmd.Flags().StringVar(&filterType, "type", "", "filter by type (model, api, training)")
	cmd.Flags().StringVar(&filterState, "state", "", "filter by state (running, stopped, paused, error)")

	return cmd
}

func newProcessStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start <id>",
		Short: "Start a process instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			if err := pManager.Start(id); err != nil {
				return err
			}
			fmt.Printf("Started process %s\n", id)
			return nil
		},
	}
}

func newProcessStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop <id>",
		Short: "Stop a running process",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			if err := pManager.Stop(id); err != nil {
				return err
			}
			fmt.Printf("Stopped process %s\n", id)
			return nil
		},
	}
}

func newProcessPauseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pause <id>",
		Short: "Pause a running process",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			if err := pManager.Pause(id); err != nil {
				return err
			}
			fmt.Printf("Paused process %s\n", id)
			return nil
		},
	}
}

func newProcessRestartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart <id>",
		Short: "Restart a process",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			if err := pManager.Restart(id); err != nil {
				return err
			}
			fmt.Printf("Restarted process %s\n", id)
			return nil
		},
	}
}

func newProcessAddCmd() *cobra.Command {
	var (
		pType      string
		command    string
		argsStr    string
		workingDir string
		port       int
		host       string
		labelsStr  string
		logFile    string
		notes      string
		autoStart  bool
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new process configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, _ := cmd.Flags().GetString("id")
			name, _ := cmd.Flags().GetString("name")
			pType, _ = cmd.Flags().GetString("type")
			command, _ = cmd.Flags().GetString("command")
			workingDir, _ = cmd.Flags().GetString("working-dir")
			port, _ = cmd.Flags().GetInt("port")
			host, _ = cmd.Flags().GetString("host")
			logFile, _ = cmd.Flags().GetString("log")
			labelsStr, _ = cmd.Flags().GetString("labels")
			notes, _ = cmd.Flags().GetString("notes")
			autoStart, _ = cmd.Flags().GetBool("auto-start")

			var procArgs []string
			if argsStr != "" {
				procArgs = strings.Split(argsStr, ",")
			}

			var procLabels map[string]string
			if labelsStr != "" {
				procLabels = parseKeyValue(labelsStr)
			}

			pc := config.ProcessConfig{
				ID:         id,
				Name:       name,
				Type:       pType,
				Command:    command,
				Args:       procArgs,
				WorkingDir: workingDir,
				Port:       port,
				Host:       host,
				LogFile:    logFile,
				Labels:     procLabels,
				Notes:      notes,
				AutoStart:  autoStart,
			}

			if err := getConfigMgr().AddProcess(pc); err != nil {
				return err
			}

			p := &process.Process{
				ID:         pc.ID,
				Name:       pc.Name,
				Type:       process.InstanceType(pc.Type),
				Command:    pc.Command,
				Args:       pc.Args,
				WorkingDir: pc.WorkingDir,
				Port:       pc.Port,
				Host:       pc.Host,
				LogFile:    lManager.GetLogPath(pc.ID),
				Labels:     pc.Labels,
				Notes:      pc.Notes,
			}
			pManager.Add(p)

			fmt.Printf("Added process %s (%s)\n", id, name)
			return nil
		},
	}

	cmd.Flags().String("id", "", "process ID (required)")
	cmd.Flags().String("name", "", "process name (required)")
	cmd.Flags().StringP("type", "t", "model", "instance type (model, api, training)")
	cmd.Flags().String("command", "", "executable command (required)")
	cmd.Flags().String("args", "", "comma-separated arguments")
	cmd.Flags().String("working-dir", "", "working directory")
	cmd.Flags().Int("port", 0, "service port")
	cmd.Flags().String("host", "localhost", "service host")
	cmd.Flags().String("log", "", "log file path")
	cmd.Flags().String("labels", "", "key=value pairs separated by commas")
	cmd.Flags().String("notes", "", "notes")
	cmd.Flags().Bool("auto-start", false, "auto-start on dashboard launch")
	cmd.MarkFlagRequired("id")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("command")

	return cmd
}

func newProcessRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <id>",
		Aliases: []string{"rm"},
		Short: "Remove a process configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			if err := getConfigMgr().RemoveProcess(id); err != nil {
				return err
			}
			if !pManager.Delete(id) {
				lManager.Clear(id)
			}
			stateMgr.Remove(id)
			stateMgr.Save()
			fmt.Printf("Removed process %s\n", id)
			return nil
		},
	}
}

func newProcessEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a process configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			pc, err := getConfigMgr().GetProcess(id)
			if err != nil {
				return err
			}

			name, _ := cmd.Flags().GetString("name")
			if name != "" {
				pc.Name = name
			}
			command, _ := cmd.Flags().GetString("command")
			if command != "" {
				pc.Command = command
			}
			argsStr, _ := cmd.Flags().GetString("args")
			if argsStr != "" {
				pc.Args = strings.Split(argsStr, ",")
			}
			notes, _ := cmd.Flags().GetString("notes")
			if notes != "" {
				pc.Notes = notes
			}

			if err := getConfigMgr().UpdateProcess(*pc); err != nil {
				return err
			}

			p, ok := pManager.Get(id)
			if ok {
				p.Name = pc.Name
				p.Command = pc.Command
				p.Args = pc.Args
				p.Notes = pc.Notes
				pManager.Update(p)
			}

			fmt.Printf("Updated process %s\n", id)
			return nil
		},
	}

	cmd := &cobra.Command{}
	cmd.Flags().String("name", "", "new name")
	cmd.Flags().String("command", "", "new command")
	cmd.Flags().String("args", "", "new arguments (comma-separated)")
	cmd.Flags().String("notes", "", "new notes")

	return cmd
}

func newProcessAutoStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "auto-start <id> [enable|disable]",
		Short: "Enable or disable auto-start for a process",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			action := "enable"
			if len(args) > 1 {
				action = args[1]
			}

			pc, err := getConfigMgr().GetProcess(id)
			if err != nil {
				return err
			}

			enabled := false
			if action == "enable" {
				enabled = true
			}

			pc.AutoStart = enabled
			if err := getConfigMgr().UpdateProcess(*pc); err != nil {
				return err
			}

			if enabled {
				fmt.Printf("Auto-start enabled for %s\n", id)
			} else {
				fmt.Printf("Auto-start disabled for %s\n", id)
			}
			return nil
		},
	}
}

func newProcessHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health <id>",
		Short: "Check health of a running process",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			healthy, err := pManager.HealthCheck(id)
			if err != nil {
				return err
			}
			if healthy {
				fmt.Printf("Process %s is healthy\n", id)
			} else {
				fmt.Printf("Process %s is unhealthy or not responding\n", id)
			}
			return nil
		},
	}
}

func newMonitorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Aliases: []string{"mon"},
		Short: "Monitor system resources",
	}

	cmd.AddCommand(
		newMonitorStatusCmd(),
		newMonitorStreamCmd(),
	)

	return cmd
}

func newMonitorStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current resource usage",
		RunE: func(cmd *cobra.Command, args []string) error {
			usage, err := mon.GetUsage()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "CPU Usage:\t%.1f%%\n", usage.CPU)
			fmt.Fprintf(w, "Memory Used:\t%d MB\n", usage.MemoryMB)
			if usage.GPU != nil {
				fmt.Fprintf(w, "GPU (%s):\t%d%% util, %d/%d MB VRAM, %d C, %.1f W\n",
					usage.GPU.Name, usage.GPU.Utilization,
					usage.GPU.MemoryUsedMB, usage.GPU.MemoryTotalMB,
					usage.GPU.Temperature, usage.GPU.PowerWatts)
			} else {
				fmt.Fprintln(w, "GPU:\tNot detected")
			}
			w.Flush()
			return nil
		},
	}
}

func newMonitorStreamCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stream",
		Short: "Stream resource usage in real-time",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Streaming resource usage (Ctrl+C to stop)...")
			mon.StartStreaming(func(u *monitor.ResourceUsage) {
				fmt.Printf("\r[%s] CPU: %.1f%% | MEM: %d MB",
					u.Timestamp.Format("15:04:05"), u.CPU, u.MemoryMB)
				if u.GPU != nil {
					fmt.Printf(" | GPU: %.0f%%", u.GPU.Utilization)
				}
			})
			<-cmd.Context().Done()
			fmt.Println()
			return nil
		},
	}
}

func newLogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Aliases: []string{"logs"},
		Short: "Manage process logs",
	}

	cmd.AddCommand(
		newLogTailCmd(),
		newLogSearchCmd(),
		newLogListCmd(),
		newLogClearCmd(),
	)

	return cmd
}

func newLogTailCmd() *cobra.Command {
	var lines int
	cmd := &cobra.Command{
		Use:   "tail <id>",
		Short: "Show recent log entries",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			entries, err := lManager.Tail(id, lines)
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				fmt.Println("No log entries found.")
				return nil
			}
			for _, line := range entries {
				fmt.Println(line)
			}
			return nil
		},
	}

	cmd.Flags().IntVarP(&lines, "lines", "n", 50, "number of lines to show")
	return cmd
}

func newLogSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <id> <query>",
		Short: "Search log entries",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			query := args[1]
			matches, err := lManager.Search(id, query)
			if err != nil {
				return err
			}
			if len(matches) == 0 {
				fmt.Println("No matching entries found.")
				return nil
			}
			for _, line := range matches {
				fmt.Println(line)
			}
			return nil
		},
	}
}

func newLogListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Aliases: []string{"ls"},
		Short: "List all log files",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := lManager.ListLogs()
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				fmt.Println("No log files found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "PROCESS ID\tSIZE\tLAST MODIFIED")
			fmt.Fprintln(w, strings.Repeat("-", 50))
			for _, e := range entries {
				fmt.Fprintf(w, "%s\t%s\t%s\n",
					e.ProcessID, formatSize(e.Size), e.Modified.Format("2006-01-02 15:04:05"))
			}
			w.Flush()
			return nil
		},
	}
}

func newLogClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear <id>",
		Short: "Clear a process log",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			if err := lManager.Clear(id); err != nil {
				return err
			}
			fmt.Printf("Cleared log for process %s\n", id)
			return nil
		},
	}
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Aliases: []string{"cfg"},
		Short: "Manage configuration",
	}

	cmd.AddCommand(
		newConfigShowCmd(),
		newConfigResetCmd(),
	)

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(cfgPath)
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		},
	}
}

func newConfigResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset to default configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			defaultCfg := config.DefaultConfig()
			if err := getConfigMgr().Save(defaultCfg); err != nil {
				return err
			}
			fmt.Println("Configuration reset to defaults.")
			return nil
		},
	}
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show overall system status",
		RunE: func(cmd *cobra.Command, args []string) error {
			usage, err := mon.GetUsage()
			if err != nil {
				return err
			}

			fmt.Println("=== AI Manager Status ===")
			fmt.Printf("Config:    %s\n", cfgPath)
			fmt.Printf("Log Dir:   %s\n", cfg.General.LogDir)
			fmt.Printf("CPU:       %.1f%%\n", usage.CPU)
			fmt.Printf("Memory:    %d MB\n", usage.MemoryMB)

			if usage.GPU != nil {
				fmt.Printf("GPU:       %s (%.0f%% util, %d/%d MB VRAM)\n",
					usage.GPU.Name, usage.GPU.Utilization,
					usage.GPU.MemoryUsedMB, usage.GPU.MemoryTotalMB)
			}

			fmt.Println()
			fmt.Println("Processes:")
			procs := pManager.List()
			if len(procs) == 0 {
				fmt.Println("  No processes configured.")
			} else {
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "  ID\tNAME\tTYPE\tSTATE\tAUTO-START")
				fmt.Fprintln(w, "  "+strings.Repeat("-", 60))
				for _, p := range procs {
					autoStart := ""
					for _, pc := range cfg.Processes {
						if pc.ID == p.ID && pc.AutoStart {
							autoStart = "yes"
							break
						}
					}
					fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\n", p.ID, p.Name, p.Type, p.State, autoStart)
				}
				w.Flush()
			}
			return nil
		},
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("ai-manager v0.1.0")
		},
	}
}

func newTuiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Launch TUI mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Launch()
		},
	}
}

func getConfigMgr() *config.Manager {
	return configMgr
}

func parseKeyValue(s string) map[string]string {
	m := make(map[string]string)
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) == 2 {
			m[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return m
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
