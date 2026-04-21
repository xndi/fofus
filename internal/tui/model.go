package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xndi/fofus/config"
	"github.com/xndi/fofus/internal/pet"
)

type tickMsg time.Time
type bubbleTickMsg time.Time
type bubbleClearMsg struct{}
type thinkingTickMsg struct{}

type llmResponseMsg struct {
	text    string
	isSmart bool
	err     error
}

type bubbleResponseMsg struct {
	text string
}

type Model struct {
	cfg           config.Config
	state         pet.State
	chatLog       []chatLine
	input         string
	chatMode      bool
	thinking      bool
	thinkingFrame int
	bubble        string
	scrollOffset  int // 0 = pinned to bottom; positive = scrolled up N rendered lines
}

type chatLine struct {
	from string
	text string
	at   time.Time
}

func New(cfg config.Config) Model {
	return Model{
		cfg:   cfg,
		state: pet.Load(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), bubbleTickCmd())
}

func tickCmd() tea.Cmd {
	return tea.Tick(15*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func bubbleTickCmd() tea.Cmd {
	return tea.Tick(15*time.Second, func(t time.Time) tea.Msg {
		return bubbleTickMsg(t)
	})
}

func bubbleClearCmd() tea.Cmd {
	return tea.Tick(8*time.Second, func(t time.Time) tea.Msg {
		return bubbleClearMsg{}
	})
}

func thinkingTickCmd() tea.Cmd {
	return tea.Tick(400*time.Millisecond, func(t time.Time) tea.Msg {
		return thinkingTickMsg{}
	})
}
