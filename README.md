# ai-manager

CLI tool for managing AI model inference, API services, and training jobs on local machines.

## Features

- **TUI Dashboard** — Bubble Tea-based interactive terminal interface
- Process lifecycle management (start, stop, pause, restart)
- Resource monitoring (CPU, memory, GPU via nvidia-smi)
- Log management (tail, search, list)
- Ollama model discovery and management
- YAML-based configuration with auto-start support
- Runtime state persistence across restarts
- Multi-instance process support
- Windows Registry autostart
- Windows process management with OS-level integration

## Requirements

- Go 1.22+
- Windows (tested on Windows 10/11)
- NVIDIA GPU with nvidia-smi for GPU monitoring (optional)

## Installation

```bash
git clone https://github.com/ginhikari/ai-manager.git
cd ai-manager
go build -o ai-manager.exe ./cmd
```

The binary will be created at `./ai-manager.exe`.

## Quick Start

```bash
# View all processes
ai-manager.exe process list

# Add a new process
ai-manager.exe process add --id ollama --name "Ollama Server" --type api --command ollama --port 11434

# Start a process
ai-manager.exe process start ollama

# Check system status
ai-manager.exe status

# View logs
ai-manager.exe log tail ollama
```

## Configuration

Config is stored at `~/.ai-manager.yaml`:

```yaml
general:
  log_dir: "~/.ai-manager/logs"
  data_dir: "~/.ai-manager/data"
  auto_save: true

processes:
  - id: ollama
    name: "Ollama Server"
    type: api
    command: ollama
    port: 11434
    host: localhost
    auto_start: true

monitor:
  interval: "10s"
```

### Process Configuration

| Field | Description | Required |
|-------|-------------|----------|
| `id` | Unique process identifier | Yes |
| `name` | Display name | Yes |
| `type` | Instance type: `model`, `api`, `training` | Yes |
| `command` | Executable command | Yes |
| `args` | Command arguments (list) | No |
| `working_dir` | Working directory | No |
| `env` | Environment variables (key-value map) | No |
| `port` | Service port | No |
| `host` | Service host (default: `localhost`) | No |
| `log_file` | Custom log file path | No |
| `labels` | Key-value labels | No |
| `notes` | Notes | No |
| `auto_start` | Auto-start on dashboard launch | No |

### Monitor Configuration

| Field | Description | Default |
|-------|-------------|---------|
| `interval` | Monitoring interval (Go duration) | `10s` |

## Commands Reference

### Process Management

```bash
ai-manager process list                          # List all processes
ai-manager process add --id <id> --name <name> --type <type> --command <cmd> [--port <port>] [--auto-start]
ai-manager process start <id>                    # Start a process
ai-manager process stop <id>                     # Stop a process
ai-manager process pause <id>                    # Pause a process
ai-manager process restart <id>                  # Restart a process
ai-manager process remove <id>                   # Remove a process
ai-manager process edit <id> [--name <name>] [--command <cmd>] [--args <args>] [--notes <notes>]
ai-manager process auto-start <id> [enable|disable]  # Toggle auto-start
ai-manager process health <id>                   # Check process health
```

### Monitoring

```bash
ai-manager monitor status                        # Show current resource usage
ai-manager monitor stream                        # Stream resource usage (Ctrl+C to stop)
```

### Logging

```bash
ai-manager log tail <id> [--lines <n>]           # Tail log file
ai-manager log search <id> --pattern <regex>     # Search log file
ai-manager log list                              # List all log files
ai-manager log clear <id>                        # Clear log file
```

### Configuration

```bash
ai-manager config show                           # Show current configuration
ai-manager config reset                          # Reset to defaults
```

### System

```bash
ai-manager status                                # Show overall system status
ai-manager version                               # Show version
```

## TUI Dashboard

Launch the interactive dashboard:

```bash
ai-manager --tui
```

### Navigation

| Key | Action |
|-----|--------|
| `1-5` | Switch tabs |
| `r` | Refresh current view |
| `ctrl+c` | Quit |

### Tabs

- **Dashboard** — System overview with CPU, memory, GPU, running services, and Ollama models
- **Models** — Discover and manage Ollama models (load, unload, refresh)
- **Utilities** — Log viewer and health checks for running services
- **Resources** — Real-time resource monitoring with history
- **Settings** — Service configuration and paths

### Keyboard Shortcuts (per tab)

**Models tab:**
- `up/down` or `k/j` — Navigate model list
- `l` — Load selected model
- `u` — Unload selected model
- `r` — Refresh model list

**Utilities tab:**
- `left/right` — Switch between logs and health views
- `up/down` or `k/j` — Navigate logs
- `g/G` — Jump to top/bottom of logs
- `r` — Refresh

**Resources tab:**
- `up/down` or `k/j` — Navigate history
- `r` — Refresh

**Settings tab:**
- `up/down` or `k/j` — Navigate services
- `enter` — Stop selected service

## Project Structure

```
ai-manager/
├── cmd/
│   └── main.go              # Entry point with --tui flag
├── internal/
│   ├── cli/
│   │   └── commands.go      # CLI commands (cobra)
│   ├── config/
│   │   └── config.go        # YAML config management
│   ├── process/
│   │   └── manager.go       # Process lifecycle management
│   ├── monitor/
│   │   └── monitor.go       # Resource monitoring (CPU, memory, GPU)
│   ├── logger/
│   │   └── logger.go        # Log file management
│   ├── state/
│   │   └── state.go         # Runtime state persistence
│   ├── ollama/
│   │   └── client.go        # Ollama API client
│   ├── registry/
│   │   └── autostart.go     # Windows Registry autostart
│   └── tui/
│       ├── app.go           # Bubble Tea main app
│       ├── model.go         # AppContext shared state
│       ├── types/
│       │   └── types.go     # Shared types
│       ├── styles/
│       │   └── theme.go     # Consistent styling
│       ├── util/
│       │   └── format.go    # Formatting utilities
│       ├── components/      # Reusable UI components
│       │   ├── gauge.go
│       │   ├── sparkline.go
│       │   ├── table.go
│       │   ├── confirm.go
│       │   ├── tabbar.go
│       │   ├── statusbar.go
│       │   └── tabbar.go
│       └── tabs/            # Tab implementations
│           ├── dashboard.go
│           ├── models.go
│           ├── utilities.go
│           ├── resources.go
│           └── settings.go
├── go.mod
├── go.sum
└── README.md
```

## Runtime Files

- `~/.ai-manager.yaml` — User configuration
- `~/.ai-manager-state.json` — Runtime state (PIDs, timestamps)
- `~/.ai-manager/logs/` — Log files per process

## License

MIT
