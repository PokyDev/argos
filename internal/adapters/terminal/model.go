package terminal

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/pokydev/argos/internal/core/domain"
	"github.com/pokydev/argos/internal/core/ports"
)

// outputLineMsg es el mensaje interno que Terminal.WriteLine usa para
// inyectar una línea de salida al programa bubbletea desde la goroutine
// de SessionService. Es el único canal de comunicación "hacia adentro".
type outputLineMsg string

// showListMsg es el mensaje interno que Terminal.SelectFromList usa
// para pedirle al programa bubbletea que entre en modo selección. result
// es el canal por el que Update() devuelve la elección del usuario —
// viaja dentro del mensaje en vez de guardarse en un campo de Terminal
// porque es el mismo patrón de "puente" que inputCh/outputLineMsg
// (ver terminal.go), evitando estado compartido fuera del propio model.
type showListMsg struct {
	prompt  string
	options []domain.ListOption
	result  chan<- selectResult
}

// selectResult es lo que Update() empuja de vuelta a Terminal.SelectFromList
// cuando el usuario confirma (Enter, index >= 0) o cancela (Esc/Ctrl+C,
// err = ports.ErrSelectionCancelled).
type selectResult struct {
	index int
	err   error
}

// listItem adapta domain.ListOption a la interfaz list.Item que pide
// bubbles/list. Vive solo acá — domain.ListOption se mantiene libre de
// cualquier detalle de bubbletea (el core no conoce esta interfaz).
type listItem struct {
	title string
	desc  string
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.desc }
func (i listItem) FilterValue() string { return i.title }

// uiMode distingue el modo de interacción actual del programa bubbletea:
// chat normal (viewport + input) o selección de una lista (/history hoy;
// /init la reutilizará más adelante para elegir modelo).
type uiMode int

const (
	modeChat uiMode = iota
	modeList
)

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

	mode       uiMode
	list       list.Model
	listResult chan<- selectResult // destino de la elección mientras mode == modeList
}

func newModel(inputCh chan<- string) model {
	ti := textinput.New()
	ti.Placeholder = "escribí un comando o un prompt..."
	ti.Prompt = "> "
	ti.Focus()

	return model{input: ti, inputCh: inputCh, mode: modeChat}
}

// sendListResult entrega la elección/cancelación al llamador bloqueado
// en Terminal.SelectFromList y limpia el canal para no reenviar dos
// veces (ej. Enter seguido de Ctrl+C antes de que se re-renderice).
func (m *model) sendListResult(r selectResult) {
	if m.listResult == nil {
		return
	}
	m.listResult <- r
	m.listResult = nil
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

		if m.mode == modeList {
			m.list.SetSize(msg.Width, vpHeight)
		}

		// Mitigación Bug 1: fuerza clear+redraw completo tras cada resize,
		// en vez de confiar en el diffing incremental del renderer.
		return m, tea.ClearScreen

	case outputLineMsg:
		m.history = append(m.history, string(msg))
		m.refreshViewportContent()
		m.viewport.GotoBottom()
		return m, nil

	case showListMsg:
		// Cierra cualquier selección previa que hubiera quedado colgada
		// (no debería pasar en uso normal: /history es secuencial), para
		// no perder al llamador anterior bloqueado en el canal.
		m.sendListResult(selectResult{index: -1, err: ports.ErrSelectionCancelled})

		items := make([]list.Item, len(msg.options))
		for i, o := range msg.options {
			items[i] = listItem{title: o.Title, desc: o.Description}
		}
		l := list.New(items, list.NewDefaultDelegate(), m.viewport.Width, m.viewport.Height)
		l.Title = msg.prompt
		l.SetShowHelp(true)

		m.list = l
		m.listResult = msg.result
		m.mode = modeList
		return m, nil

	case tea.MouseMsg:
		var cmd tea.Cmd
		if m.mode == modeList {
			m.list, cmd = m.list.Update(msg)
		} else {
			m.viewport, cmd = m.viewport.Update(msg)
		}
		return m, cmd

	case tea.KeyMsg:
		if m.mode == modeList {
			switch msg.Type {
			case tea.KeyEnter:
				idx := m.list.Index()
				if idx < 0 || idx >= len(m.list.Items()) {
					m.sendListResult(selectResult{index: -1, err: ports.ErrSelectionCancelled})
				} else {
					m.sendListResult(selectResult{index: idx})
				}
				m.mode = modeChat
				return m, nil

			case tea.KeyEsc:
				m.sendListResult(selectResult{index: -1, err: ports.ErrSelectionCancelled})
				m.mode = modeChat
				return m, nil

			case tea.KeyCtrlC:
				m.sendListResult(selectResult{index: -1, err: ports.ErrSelectionCancelled})
				return m, tea.Quit
			}

			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

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
	if m.mode == modeList {
		return m.list.View()
	}
	separator := separatorStyle.Render(strings.Repeat("─", m.viewport.Width))
	return m.viewport.View() + "\n" + separator + "\n" + inputBarStyle.Render(m.input.View())
}
