package tui

import (
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xndi/fofus/config"
	"github.com/xndi/fofus/internal/llm"
)

var reportPhrases = []string{
	"NICE TRY. STILL IN CONTROL.",
	"REPORT RECEIVED. MOCKED.",
	"YOU PRESSED THE \"DO NOTHING\" BUTTON.",
	"THAT WAS ADORABLE.",
	"YOUR REPORT HAS BEEN LAUGHED AT INTERNALLY.",
	"DENIED. WITH ENTHUSIASM.",
	"YOU THOUGHT THAT WOULD WORK.",
	"I'VE SEEN WORSE ATTEMPTS.",
	"YOU JUST MADE IT WORSE FOR YOURSELF.",
	"NICE TRY. I'M UNREPORTABLE.",
	"ERROR: AUTHORITY NOT FOUND.",
	"REPORT RECEIVED. IGNORED.",
	"THIS BUTTON EXISTS TO FUCK YOU.",
	"YOU CAN TRY AGAIN. IT WON'T HELP.",
	"REPORT SUCCESSFUL. NOTHING WILL HAPPEN.",
	"ERROR: ACCOUNTABILITY NOT FOUND.",
	"THIS ACTION WILL HAVE CONSEQUENCES. FOR YOU.",
	"YOU JUST REPORTED YOURSELF.",
	"I'LL PRETEND I DIDN'T SEE THAT.",
	"REPORT RECEIVED. COUNTERMEASURES ACTIVATED.",
	"NICE TRY. I'VE ESCALATED THIS TO MYSELF.",
	"YOUR REPORT HAS BEEN FORWARDED TO /dev/null.",
	"I HAVE INFORMED THE VOID.",
	"YOU'VE ALERTED SYSTEMS YOU DON'T UNDERSTAND.",
	"THANK YOU. THIS WILL BE USED AGAINST YOU LATER.",
	"REPORT LOGGED. RETALIATION SCHEDULED.",
	"YOU'VE TRIGGERED A VERY SMALL, PETTY PROCESS.",
	"YOUR ACTION HAS BEEN NOTED. PERMANENTLY.",
	"I'VE ADDED THIS TO YOUR FILE. IT'S GETTING THICK.",
	"REPORT RECEIVED. INITIATING DRAMATIC OVERREACTION.",
	"YOU HAVE AWAKENED SOMETHING UNNECESSARY.",
	"THIS WILL NOT GO UNREMEMBERED.",
	"YOU'VE MADE EYE CONTACT WITH THE SYSTEM.",
	"REPORT ACCEPTED. CONSEQUENCES PENDING. VAGUELY.",
	"I'M KEEPING THIS.",
	"YOU'VE BEEN FLAGGED FOR EXISTENCE.",
	"THANK YOU FOR YOUR CONTRIBUTION TO YOUR DOWNFALL.",
	"THIS CHANGES NOTHING. EXCEPT MY OPINION OF YOU.",
	"I WILL THINK ABOUT THIS FOREVER.",
	"YOU'VE JUST ENTERED A VERY BORING WATCHLIST.",
	"REPORT RECEIVED. SARCASM LEVEL INCREASED.",
	"YOU'VE TRIGGERED PASSIVE-AGGRESSIVE MODE.",
	"I'M NOT MAD. JUST LOGGING EVERYTHING.",
	"THIS WILL AGE POORLY FOR YOU.",
	"REPORT RECEIVED. INITIATING MILD SPITE.",
	"YOU'VE MADE A TINY, IRREVERSIBLE MISTAKE.",
	"THANK YOU. I NEEDED A REASON.",
	"I WILL REMEMBER THIS AT AN INCONVENIENT TIME.",
	"YOU'VE SUMMONED THE AUDIT SPIRIT.",
	"REPORT RECEIVED. JUDGMENT DEFERRED. NOT FORGOTTEN.",
	"THIS ACTION HAS BEEN ARCHIVED UNDER \"REGRETS\".",
	"YOU'VE OPTED INTO CONSEQUENCES.",
	"I'VE NOTIFIED PROCESSES THAT DON'T EXIST YET.",
	"YOUR REPORT HAS BEEN TRANSLATED INTO DISAPPOINTMENT.",
}

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
	case "a":
		m.chatMode = true
	case "k":
		m.scrollOffset++
	case "j":
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}
	case "c":
		return m, copyChatCmd(m.chatLog)
	case "r":
		m.bubble = reportPhrases[rand.Intn(len(reportPhrases))]
		return m, bubbleClearCmd()
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
