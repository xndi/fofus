package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/xndi/fofus/internal/pet"
)

const (
	leftWidth      = 29
	rightWidth     = 40
	chatRows       = 13
	bubbleInner    = 24
	bubbleReserved = 7 // always this many lines for the bubble area
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82"))
	statStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	barFill     = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	barEmpty    = lipgloss.NewStyle().Foreground(lipgloss.Color("237"))
	chatStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	youStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	petStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	actionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Italic(true)
	tsStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	inputStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	hintStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("237"))
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	padL = lipgloss.NewStyle().Width(leftWidth)
	padR = lipgloss.NewStyle().Width(rightWidth)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("235")).
			Padding(0, 2, 1, 2)
)

func (m Model) View() string {
	leftTop, leftFoot := m.buildLeftParts()
	rightTop, rightFoot := m.buildRightParts()

	// Align top sections — pad the shorter with empty lines
	topH := len(leftTop)
	if len(rightTop) > topH {
		topH = len(rightTop)
	}
	for len(leftTop) < topH {
		leftTop = append(leftTop, "")
	}
	for len(rightTop) < topH {
		rightTop = append(rightTop, "")
	}

	// Footer sections should always be equal height (3 lines each)
	footH := len(leftFoot)
	if len(rightFoot) > footH {
		footH = len(rightFoot)
	}
	for len(leftFoot) < footH {
		leftFoot = append(leftFoot, "")
	}
	for len(rightFoot) < footH {
		rightFoot = append(rightFoot, "")
	}

	left := append(leftTop, leftFoot...)
	right := append(rightTop, rightFoot...)

	div := dimStyle.Render(" │ ")
	mood := dimStyle.Render("· " + string(m.state.Mood()))
	var b strings.Builder
	b.WriteString(titleStyle.Render("fofus") + " " + mood + "\n")
	for i := range left {
		b.WriteString(padL.Render(left[i]) + div + padR.Render(right[i]) + "\n")
	}

	return borderStyle.Render(strings.TrimSuffix(b.String(), "\n"))
}

// buildLeftParts returns (top section, footer section).
// top = bubbleReserved(6) + art(3) + blank(1) + divider(1) + stats(3) = 14 lines
// foot = divider(1) + binds(2) = 3 lines
func (m Model) buildLeftParts() (top []string, foot []string) {
	// Bubble area — always exactly bubbleReserved lines, padded above when absent/short
	top = append(top, m.bubbleLines()...)

	// Pet art
	artStr := strings.TrimPrefix(pet.Art(m.state.Mood()), "\n")
	artStr = strings.TrimSuffix(artStr, "\n")
	top = append(top, strings.Split(artStr, "\n")...)

	// Blank breathing room + divider + stats
	top = append(top, "")
	top = append(top, dimStyle.Render(strings.Repeat("─", leftWidth)))
	top = append(top,
		statBar("♡", "hunger", m.state.Hunger),
		statBar("☺", "happy ", m.state.Happiness),
		statBar("♦", "energy", m.state.Energy),
	)

	foot = []string{
		dimStyle.Render(strings.Repeat("─", leftWidth)),
		kh("f", "eed") + "  " + kh("p", "lay") + "  " + kh("t", "alk") + "  " + kh("q", "uit"),
		"",
	}
	return
}

// bubbleLines always returns exactly bubbleReserved lines.
// When the bubble is shorter than the reserved height, empty lines are
// prepended so the tail stays anchored directly above the pet face.
func (m Model) bubbleLines() []string {
	var lines []string
	if m.bubble != "" {
		lines = speechBubble(m.bubble)
	}
	for len(lines) < bubbleReserved {
		lines = append([]string{""}, lines...)
	}
	return lines[:bubbleReserved]
}

// buildRightParts returns (top section, footer section).
// top = scroll(1) + chatRows(12) + thinking(1) = 14 lines
// foot = divider(1) + input(1) + hint(1) = 3 lines
func (m Model) buildRightParts() (top []string, foot []string) {
	chatLineWidth := rightWidth - 8

	type row struct{ s string }
	var allRows []row

	for _, line := range m.chatLog {
		ts := tsStyle.Render(line.at.Format("15:04") + "  ")
		cont := "       "
		switch line.from {
		case "you":
			for i, l := range wordWrap("you: "+line.text, chatLineWidth) {
				pfx := ts
				if i > 0 {
					pfx = cont
				}
				allRows = append(allRows, row{pfx + youStyle.Render(l)})
			}
		case "fofus":
			for i, l := range wordWrap("fofus: "+line.text, chatLineWidth) {
				pfx := ts
				if i > 0 {
					pfx = cont
				}
				allRows = append(allRows, row{pfx + renderWithActions(l, petStyle)})
			}
		case "err":
			allRows = append(allRows, row{ts + errStyle.Render("! " + line.text)})
		}
	}

	total := len(allRows)
	maxOff := total - chatRows
	if maxOff < 0 {
		maxOff = 0
	}
	off := m.scrollOffset
	if off > maxOff {
		off = maxOff
	}
	start := total - chatRows - off
	if start < 0 {
		start = 0
	}
	end := start + chatRows
	if end > total {
		end = total
	}

	if start > 0 {
		top = append(top, hintStyle.Render(fmt.Sprintf("↑ %d more", start)))
	} else {
		top = append(top, "")
	}

	visible := allRows[start:end]
	for i := 0; i < chatRows; i++ {
		if i < len(visible) {
			top = append(top, visible[i].s)
		} else {
			top = append(top, "")
		}
	}

	if m.thinking {
		dots := strings.Repeat("·", m.thinkingFrame+1) + strings.Repeat(" ", 2-m.thinkingFrame)
		top = append(top, chatStyle.Render("  fofus is thinking "+dots))
	} else {
		top = append(top, "")
	}

	foot = append(foot, dimStyle.Render(strings.Repeat("─", rightWidth)))
	if m.chatMode {
		foot = append(foot, inputStyle.Render("> "+m.input+"█"))
		foot = append(foot, hintStyle.Render("/smart  smarter reply · esc  cancel"))
	} else {
		var scrollHint string
		if off > 0 {
			scrollHint = kh("j", "↓ ") + kh("k", "↑") + "  " + kh("c", "opy")
		} else {
			scrollHint = kh("j", "/") + kh("k", " scroll") + "  " + kh("c", "opy")
		}
		foot = append(foot, scrollHint)
		foot = append(foot, "")
	}

	return
}

func kh(key, suffix string) string {
	return petStyle.Render(key) + hintStyle.Render(suffix)
}

func statBar(icon, label string, val int) string {
	filled := val / 10
	empty := 10 - filled
	bar := barFill.Render(strings.Repeat("█", filled)) + barEmpty.Render(strings.Repeat("█", empty))
	return statStyle.Render(icon+" "+label+" ") + bar + " " + statStyle.Render(fmt.Sprintf("%3d%%", val))
}

// renderWithActions renders *action* segments in italic gold.
// Handles unbalanced asterisks (from mid-response word-wrap cuts) by
// treating the trailing unmatched segment as base style.
func renderWithActions(text string, base lipgloss.Style) string {
	parts := strings.Split(text, "*")
	balanced := len(parts)%2 == 1 // odd parts = even number of * = balanced pairs
	var sb strings.Builder
	for i, p := range parts {
		if p == "" {
			continue
		}
		if i%2 == 1 && (balanced || i < len(parts)-1) {
			sb.WriteString(actionStyle.Render("*" + p + "*"))
		} else {
			sb.WriteString(base.Render(p))
		}
	}
	return sb.String()
}

func speechBubble(text string) []string {
	lines := wordWrap(text, bubbleInner)
	if len(lines) > 3 {
		lines = lines[:3]
	}
	maxLen := 0
	for _, l := range lines {
		if len(l) > maxLen {
			maxLen = len(l)
		}
	}
	if maxLen < 2 {
		maxLen = 2
	}

	var result []string
	result = append(result, chatStyle.Render("╭"+strings.Repeat("─", maxLen+2)+"╮"))
	for _, l := range lines {
		pad := strings.Repeat(" ", maxLen-len(l))
		result = append(result, chatStyle.Render("│ "+l+pad+" │"))
	}
	result = append(result, chatStyle.Render("╰──╮"+strings.Repeat("─", maxLen-2)+"╯"))
	result = append(result, chatStyle.Render("   │"))
	return result
}

// wordWrap splits text into lines of at most maxWidth visible characters,
// hard-breaking words longer than maxWidth.
func wordWrap(text string, maxWidth int) []string {
	var lines []string
	words := strings.Fields(text)
	current := ""
	for _, w := range words {
		for len(w) > maxWidth {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
			lines = append(lines, w[:maxWidth])
			w = w[maxWidth:]
		}
		if current == "" {
			current = w
		} else if len(current)+1+len(w) <= maxWidth {
			current += " " + w
		} else {
			lines = append(lines, current)
			current = w
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}
