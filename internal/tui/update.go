package tui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xndi/fofus/config"
	"github.com/xndi/fofus/internal/llm"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if m.chatMode {
			return m.handleChatInput(msg)
		}
		return m.handleNav(msg)

	case tickMsg:
		m.state.Tick()
		return m, tickCmd()

	case bubbleTickMsg:
		if m.bubble == "" && !m.thinking {
			cfg := m.cfg
			return m, tea.Batch(bubbleTickCmd(), idleQuipCmd(cfg))
		}
		return m, bubbleTickCmd()

	case bubbleResponseMsg:
		if msg.text != "" {
			m.bubble = msg.text
			return m, bubbleClearCmd()
		}
		return m, nil

	case bubbleClearMsg:
		m.bubble = ""
		return m, nil

	case thinkingTickMsg:
		if m.thinking {
			m.thinkingFrame = (m.thinkingFrame + 1) % 3
			return m, thinkingTickCmd()
		}
		return m, nil

	case llmResponseMsg:
		m.thinking = false
		m.thinkingFrame = 0
		m.scrollOffset = 0 // auto-scroll to bottom on new message
		if msg.err != nil {
			m.chatLog = append(m.chatLog, chatLine{"err", msg.err.Error(), time.Now()})
		} else {
			m.chatLog = append(m.chatLog, chatLine{"fofus", msg.text, time.Now()})
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		if m.thinking {
			break
		}
		_ = m.state.Save()
		return m, tea.Quit
	case "f":
		m.state.Feed()
	case "p":
		m.state.Play()
	case "t":
		m.chatMode = true
	case "k":
		m.scrollOffset++
	case "j":
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}
	case "c":
		return m, copyChatCmd(m.chatLog)
	}
	return m, nil
}

func (m Model) handleChatInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.input != "" {
			m.input = ""
		} else {
			m.chatMode = false
		}
	case "enter":
		if m.input == "" || m.thinking {
			break
		}
		userMsg := m.input
		m.chatLog = append(m.chatLog, chatLine{"you", userMsg, time.Now()})
		m.input = ""
		m.chatMode = false
		m.thinking = true
		m.scrollOffset = 0
		cfg := m.cfg
		state := m.state
		return m, tea.Batch(
			thinkingTickCmd(),
			func() tea.Msg {
				reply, isSmart, err := llm.Route(cfg, userMsg, state)
				return llmResponseMsg{reply, isSmart, err}
			},
		)
	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			m.input += string(msg.Runes)
		}
	}
	return m, nil
}

func copyChatCmd(log []chatLine) tea.Cmd {
	return func() tea.Msg {
		var sb strings.Builder
		for _, line := range log {
			sb.WriteString(fmt.Sprintf("[%s] %s: %s\n", line.at.Format("15:04"), line.from, line.text))
		}
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(sb.String())
		_ = cmd.Run()
		return nil
	}
}

func idleQuipCmd(cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		text, err := llm.OllamaIdleQuip(cfg.OllamaURL, cfg.OllamaModel)
		if err != nil {
			return bubbleResponseMsg{}
		}
		return bubbleResponseMsg{text}
	}
}
