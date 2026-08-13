package terminal

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// outputLineMsg es el mensaje interno que Terminal.WriteLine usa para
// inyectar una línea de salida al programa bubbletea desde la goroutine
// de SessionService. Es el único canal de comunicación "hacia adentro".
type outputLineMsg string

var (
	separatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	promptStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	inputBarStyle  = lipgloss.NewStyle().Padding(0, 1)
)

// model implementa tea.Model. Vive enteramente dentro de este paquete;
// el core nunca lo conoce (sigue viendo solo ports.SessionIO).
type model struct {
	viewport viewport.Model
	input    textinput.Model
	history  []string
	ready    bool // true tras el primer WindowSizeMsg (ya conocemos el tamaño real)

	inputCh chan<- string // líneas confirmadas (Enter) salen por acá hacia Terminal.ReadLine
}

func newModel(inputCh chan<- string) model {
	ti := textinput.New()
	ti.Placeholder = "escribí un comando o un prompt..."
	ti.Prompt = "> "
	ti.Focus()

	return model{input: ti, inputCh: inputCh}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *model) refreshViewportContent() {
	if m.viewport.Width == 0 {
		return
	}
	// Re-wrappea TODO el historial al ancho actual del viewport en cada
	// llamada — soluciona el Bug 2 (SetContent no reflowea por sí solo).
	wrapped := lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.history, "\n"))
	m.viewport.SetContent(wrapped)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		vpHeight := msg.Height - 3
		if vpHeight < 1 {
			vpHeight = 1
		}
		if !m.ready {
			m.viewport = viewport.New(msg.Width, vpHeight)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = vpHeight
		}
		m.refreshViewportContent() // <- antes faltaba en el caso "ya ready"
		m.viewport.GotoBottom()

		if w := msg.Width - len(m.input.Prompt) - 3; w > 0 {
			m.input.Width = w
		} else {
			m.input.Width = 1 // clamp: evita Width negativo en terminales muy angostas
		}

		// Mitigación Bug 1: fuerza clear+redraw completo tras cada resize,
		// en vez de confiar en el diffing incremental del renderer.
		return m, tea.ClearScreen

	case outputLineMsg:
		m.history = append(m.history, string(msg))
		m.refreshViewportContent()
		m.viewport.GotoBottom()
		return m, nil

	case tea.MouseMsg:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEnter:
			value := strings.TrimSpace(m.input.Value())
			m.input.Reset()
			if value == "" {
				return m, nil
			}
			m.history = append(m.history, "", promptStyle.Render("> ")+value)
			m.refreshViewportContent()
			m.viewport.GotoBottom()

			select {
			case m.inputCh <- value:
			default:
				m.history = append(m.history, separatorStyle.Render("(procesando el turno anterior, esperá un momento)"))
				m.refreshViewportContent()
				m.viewport.GotoBottom()
			}
			return m, nil

		case tea.KeyPgUp:
			m.viewport.LineUp(m.viewport.Height / 2)
			return m, nil

		case tea.KeyPgDown:
			m.viewport.LineDown(m.viewport.Height / 2)
			return m, nil
		}

		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return "iniciando argos..."
	}
	separator := separatorStyle.Render(strings.Repeat("─", m.viewport.Width))
	return m.viewport.View() + "\n" + separator + "\n" + inputBarStyle.Render(m.input.View())
}
