package ui

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/meshradio/meshradio/internal/broadcaster"
	"github.com/meshradio/meshradio/internal/listener"
	"github.com/meshradio/meshradio/pkg/audio"
)

// Mode represents the current UI mode
type Mode int

const (
	ModeMain Mode = iota
	ModeBroadcast
	ModeListen
	ModeInput
)

// Model holds the application state
type Model struct {
	mode        Mode
	broadcaster *broadcaster.Broadcaster
	listener    *listener.Listener
	textInput   textinput.Model
	inputPrompt string
	err         error

	// User config
	callsign string
	localIPv6 net.IP

	// Display
	width  int
	height int
	logs   []string
}

// NewModel creates a new UI model
func NewModel(callsign string, ipv6 net.IP) Model {
	ti := textinput.New()
	ti.Placeholder = "Enter IPv6 address..."
	ti.Focus()
	ti.CharLimit = 39 // Max IPv6 length
	ti.Width = 40

	return Model{
		mode:      ModeMain,
		callsign:  callsign,
		localIPv6: ipv6,
		textInput: ti,
		logs:      make([]string, 0, 10),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		tickCmd(),
	)
}

// tickMsg is sent on each tick for UI updates
type tickMsg time.Time

// tickCmd returns a command that ticks every second
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		// Refresh UI every second
		return m, tickCmd()

	case errMsg:
		m.err = msg
		return m, nil
	}

	// Update text input
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeMain:
		return m.handleMainMode(msg)
	case ModeInput:
		return m.handleInputMode(msg)
	case ModeBroadcast:
		return m.handleBroadcastMode(msg)
	case ModeListen:
		return m.handleListenMode(msg)
	}
	return m, nil
}

// handleMainMode handles main menu input
func (m Model) handleMainMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "b":
		// Start broadcast mode
		m.addLog("Starting broadcast mode...")
		return m.startBroadcast()

	case "l":
		// Enter listen mode - prompt for IPv6
		m.mode = ModeInput
		m.inputPrompt = "Enter station IPv6 to listen:"
		m.textInput.Reset()
		m.textInput.Focus()
		return m, nil

	case "i":
		// Show info
		m.addLog(fmt.Sprintf("Callsign: %s", m.callsign))
		m.addLog(fmt.Sprintf("IPv6: %s", m.localIPv6.String()))
		return m, nil
	}

	return m, nil
}

// handleInputMode handles text input
func (m Model) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		// Process input
		ipv6Str := m.textInput.Value()
		ipv6 := net.ParseIP(ipv6Str)
		if ipv6 == nil {
			m.err = fmt.Errorf("invalid IPv6 address: %s", ipv6Str)
			m.mode = ModeMain
			return m, nil
		}

		m.addLog(fmt.Sprintf("Connecting to %s...", ipv6.String()))
		return m.startListener(ipv6)

	case tea.KeyEsc:
		m.mode = ModeMain
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// handleBroadcastMode handles broadcast mode input
func (m Model) handleBroadcastMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return m.stopBroadcast()
	}
	return m, nil
}

// handleListenMode handles listen mode input
func (m Model) handleListenMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return m.stopListener()
	}
	return m, nil
}

// startBroadcast starts broadcasting
func (m Model) startBroadcast() (Model, tea.Cmd) {
	cfg := broadcaster.Config{
		Callsign:    m.callsign,
		IPv6:        m.localIPv6,
		Port:        8799, // 799 ~ Ygg (avoid conflict with Yggdrasil port 9001)
		AudioConfig: audio.DefaultConfig(),
	}

	b, err := broadcaster.New(cfg)
	if err != nil {
		m.err = err
		m.addLog(fmt.Sprintf("Error: %v", err))
		return m, nil
	}

	if err := b.Start(); err != nil {
		m.err = err
		m.addLog(fmt.Sprintf("Error: %v", err))
		return m, nil
	}

	m.broadcaster = b
	m.mode = ModeBroadcast
	m.addLog(fmt.Sprintf("Broadcasting on %s:8799", m.localIPv6.String()))
	m.addLog("Share this address with listeners!")
	m.addLog("Press 'q' or ESC to stop")

	return m, nil
}

// stopBroadcast stops broadcasting
func (m Model) stopBroadcast() (Model, tea.Cmd) {
	if m.broadcaster != nil {
		m.broadcaster.Stop()
		m.broadcaster = nil
		m.addLog("Broadcast stopped")
	}
	m.mode = ModeMain
	return m, nil
}

// startListener starts listening
func (m Model) startListener(targetIPv6 net.IP) (Model, tea.Cmd) {
	cfg := listener.Config{
		Callsign:    m.callsign,
		LocalIPv6:   m.localIPv6,
		LocalPort:   9799, // 799 ~ Ygg (listener port, pairs with broadcaster 8799)
		TargetIPv6:  targetIPv6,
		TargetPort:  8799, // 799 ~ Ygg (broadcaster port)
		AudioConfig: audio.DefaultConfig(),
	}

	l, err := listener.New(cfg)
	if err != nil {
		m.err = err
		m.addLog(fmt.Sprintf("Error: %v", err))
		m.mode = ModeMain
		return m, nil
	}

	if err := l.Start(); err != nil {
		m.err = err
		m.addLog(fmt.Sprintf("Error: %v", err))
		m.mode = ModeMain
		return m, nil
	}

	m.listener = l
	m.mode = ModeListen
	m.addLog(fmt.Sprintf("Listening to %s:9001", targetIPv6.String()))
	m.addLog("Press 'q' or ESC to stop")

	return m, nil
}

// stopListener stops listening
func (m Model) stopListener() (Model, tea.Cmd) {
	if m.listener != nil {
		m.listener.Stop()
		m.listener = nil
		m.addLog("Listener stopped")
	}
	m.mode = ModeMain
	return m, nil
}

// addLog adds a log message
func (m *Model) addLog(msg string) {
	m.logs = append(m.logs, msg)
	if len(m.logs) > 10 {
		m.logs = m.logs[len(m.logs)-10:]
	}
}

// View renders the UI
func (m Model) View() string {
	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Background(lipgloss.Color("235")).
		Padding(0, 1)

	b.WriteString(headerStyle.Render("MeshRadio v0.1-alpha"))
	b.WriteString("\n\n")

	// Status
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	b.WriteString(statusStyle.Render(fmt.Sprintf("Callsign: %s | IPv6: %s", m.callsign, m.localIPv6.String())))
	b.WriteString("\n\n")

	// Mode-specific view
	switch m.mode {
	case ModeMain:
		b.WriteString(m.renderMainMenu())
	case ModeInput:
		b.WriteString(m.renderInput())
	case ModeBroadcast:
		b.WriteString(m.renderBroadcast())
	case ModeListen:
		b.WriteString(m.renderListen())
	}

	// Logs
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Recent Activity:"))
	b.WriteString("\n")
	for _, log := range m.logs {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("248")).Render("  " + log))
		b.WriteString("\n")
	}

	// Error
	if m.err != nil {
		b.WriteString("\n")
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		b.WriteString(errStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	return b.String()
}

// renderMainMenu renders the main menu
func (m Model) renderMainMenu() string {
	menu := `
Main Menu:
  [b] Broadcast - Start broadcasting audio
  [l] Listen    - Tune to a station
  [i] Info      - Show station info
  [q] Quit      - Exit MeshRadio
`
	return lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(menu)
}

// renderInput renders the input prompt
func (m Model) renderInput() string {
	return fmt.Sprintf("%s\n\n%s\n\n(Press ESC to cancel)",
		m.inputPrompt,
		m.textInput.View())
}

// renderBroadcast renders the broadcast view
func (m Model) renderBroadcast() string {
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true)

	status := statusStyle.Render("● BROADCASTING")

	// Audio level bar (simulated)
	audioLevel := "▓▓▓▓▓░░░░░"

	// Running indicator (animated)
	dots := strings.Repeat(".", (int(time.Now().Unix()) % 4))

	info := fmt.Sprintf(`
%s %s

Station:  %s
Address:  %s:9001
Codec:    Opus (simulated)
Quality:  48kHz, Mono, 64kbps
Multicast: ff02::1 (all local nodes)

Audio Level: %s

📡 Transmitting audio frames...
   Network: UDP multicast
   Status: Active

Press 'q' or ESC to stop broadcasting
`, status, dots, m.callsign, m.localIPv6.String(), audioLevel)

	return info
}

// renderListen renders the listen view
func (m Model) renderListen() string {
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)

	status := statusStyle.Render("● LISTENING")

	// Running indicator (animated)
	dots := strings.Repeat(".", (int(time.Now().Unix()) % 4))

	var stationInfo string
	var signalBar string
	if m.listener != nil {
		packets, seq, station := m.listener.GetStats()
		if station != "" {
			stationInfo = fmt.Sprintf("Station: %s", station)
			// Signal strength bar based on packets received
			strength := packets % 10
			signalBar = strings.Repeat("▓", int(strength)) + strings.Repeat("░", 10-int(strength))
		} else {
			stationInfo = "Station: Waiting for signal..."
			signalBar = "░░░░░░░░░░"
		}
		stationInfo += fmt.Sprintf("\nPackets: %d | Sequence: %d", packets, seq)
	}

	info := fmt.Sprintf(`
%s %s

%s

Signal:  %s
Network: Receiving on port 9002

🎧 Listening for audio...
   Codec: Opus
   Buffer: Good

Press 'q' or ESC to stop listening
`, status, dots, stationInfo, signalBar)

	return info
}

// errMsg is a custom error message type
type errMsg error
