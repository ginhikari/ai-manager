package tabs

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ai-manager/internal/ollama"
	"ai-manager/internal/tui/styles"
	"ai-manager/internal/tui/types"
	"ai-manager/internal/tui/util"
)

type ModelsTab struct {
	ctx         *types.AppContext
	models      []ollama.ModelInfo
	selectedIdx int
	loading     bool
	errMsg      string
	connected   bool
}

func NewModelsTab(ctx *types.AppContext) ModelsTab {
	return ModelsTab{
		ctx:     ctx,
		loading: true,
	}
}

func (t *ModelsTab) Init() tea.Cmd {
	return t.refreshModels()
}

func (t *ModelsTab) Update(msg tea.Msg) (TabInterface, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if t.selectedIdx > 0 {
				t.selectedIdx--
			}
		case "down", "j":
			if t.selectedIdx < len(t.models)-1 {
				t.selectedIdx++
			}
		case "l":
			if len(t.models) > 0 {
				return t, t.loadModel()
			}
		case "u":
			if len(t.models) > 0 {
				return t, t.unloadModel()
			}
		case "r":
			t.loading = true
			t.errMsg = ""
			return t, t.refreshModels()
		}
	case refreshModelsMsg:
		t.models = msg.models
		t.loading = false
		t.connected = msg.connected
	}
	return t, nil
}

func (t *ModelsTab) View() string {
	if t.loading {
		return styles.HelpText.Render("Loading models...")
	}

	if t.errMsg != "" {
		return lipgloss.NewStyle().Foreground(styles.ErrorColor).Render(t.errMsg)
	}

	var lines []string
	lines = append(lines, "╭─ Ollama Models ────────────────────────────────────────╮")

	if !t.connected {
		lines = append(lines, "│  Ollama server not connected                            │")
		lines = append(lines, "│  Start ollama server and refresh                        │")
	} else if len(t.models) == 0 {
		lines = append(lines, "│  No models found                                        │")
	} else {
		for i, m := range t.models {
			selected := "  "
			if i == t.selectedIdx {
				selected = "▸ "
			}
			size := util.FormatSize(m.Size)
			lines = append(lines, fmt.Sprintf("│  %s%-50s %s│", selected, m.Name, size))
		}
	}

	lines = append(lines, "╰────────────────────────────────────────────────────────╯")
	lines = append(lines, "")
	lines = append(lines, "  ↑/↓ Navigate  l: Load  u: Unload  r: Refresh")

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (t *ModelsTab) refreshModels() tea.Cmd {
	return func() tea.Msg {
		connected := t.ctx.Ollama.IsServerRunning()
		var models []ollama.ModelInfo
		var err error

		if connected {
			models, err = t.ctx.Ollama.ListModels()
		}

		if err != nil {
			return refreshModelsMsg{models: models, connected: connected, err: err}
		}
		return refreshModelsMsg{models: models, connected: connected}
	}
}

func (t *ModelsTab) loadModel() tea.Cmd {
	return func() tea.Msg {
		if t.selectedIdx >= len(t.models) {
			return nil
		}
		model := t.models[t.selectedIdx]
		err := t.ctx.Ollama.LoadModel(model.Name)
		return modelActionMsg{model: model, action: "load", err: err}
	}
}

func (t *ModelsTab) unloadModel() tea.Cmd {
	return func() tea.Msg {
		if t.selectedIdx >= len(t.models) {
			return nil
		}
		model := t.models[t.selectedIdx]
		err := t.ctx.Ollama.UnloadModel(model.Name)
		return modelActionMsg{model: model, action: "unload", err: err}
	}
}

type refreshModelsMsg struct {
	models    []ollama.ModelInfo
	connected bool
	err       error
}

type modelActionMsg struct {
	model  ollama.ModelInfo
	action string
	err    error
}
