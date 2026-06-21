package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	figure "github.com/common-nighthawk/go-figure"
)

// ---- Theme (Figma-like dark) ----
var (
	bg     = lipgloss.Color("#0B0B0B") // near-black canvas
	fg     = lipgloss.Color("#FFFFFF") // pure white text
	muted  = lipgloss.Color("#9B9B9B") // dim grey
	accent = lipgloss.Color("#FF5C39") // Figma orange accent
)

const figFont = "big" // FIGlet font for the big titles

// ---- Slides ----
type kind int

const (
	cover kind = iota
	content
)

type slide struct {
	kind  kind
	title string
	body  string
}

var slides = []slide{
	{kind: cover, title: "Sprint Retrospective"},
	{kind: content, title: "Retrospective", body: "Hopper mock"},
	{kind: content, title: "Sprint 13", body: "Frontend monitor"},
	{kind: content, title: "six-month plan", body: "Research and Development.\ncost optimization\nknowhow"},
}

type model struct {
	idx int
	w   int
	h   int
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "right", "l", "n", " ", "enter", "pgdown", "down":
			if m.idx < len(slides)-1 {
				m.idx++
			}
		case "left", "h", "p", "pgup", "up":
			if m.idx > 0 {
				m.idx--
			}
		case "home", "r":
			m.idx = 0
		case "end":
			m.idx = len(slides) - 1
		}
	}
	return m, nil
}

// visibleWidth returns the widest line (ignoring trailing spaces).
func visibleWidth(s string) int {
	max := 0
	for _, line := range strings.Split(s, "\n") {
		if w := lipgloss.Width(strings.TrimRight(line, " ")); w > max {
			max = w
		}
	}
	return max
}

func fig(s string) string {
	return strings.TrimRight(figure.NewFigure(s, figFont, true).String(), "\n")
}

// bigText renders s as FIGlet art, wrapping word-by-word so it never
// exceeds maxWidth. Each wrapped group becomes its own stacked block.
func bigText(s string, maxWidth int) string {
	if maxWidth < 10 {
		return s // terminal too narrow for art; fall back to plain
	}
	whole := fig(s)
	if visibleWidth(whole) <= maxWidth {
		return whole
	}
	words := strings.Fields(s)
	var groups []string
	cur := ""
	for _, wd := range words {
		if cur == "" {
			cur = wd
			continue
		}
		try := cur + " " + wd
		if visibleWidth(fig(try)) <= maxWidth {
			cur = try
		} else {
			groups = append(groups, cur)
			cur = wd
		}
	}
	if cur != "" {
		groups = append(groups, cur)
	}
	blocks := make([]string, len(groups))
	for i, g := range groups {
		blocks[i] = fig(g)
	}
	return strings.Join(blocks, "\n")
}

func (m model) View() string {
	if m.w == 0 {
		return "" // wait for first WindowSizeMsg
	}

	s := slides[m.idx]
	maxW := m.w - 8

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(fg)
	bodyStyle := lipgloss.NewStyle().Foreground(fg)
	barStyle := lipgloss.NewStyle().Foreground(accent)

	bigTitle := titleStyle.Render(bigText(s.title, maxW))
	barWidth := visibleWidth(bigTitle)
	if barWidth > maxW {
		barWidth = maxW
	}
	underline := barStyle.Render(strings.Repeat("━", barWidth))

	var block string
	switch s.kind {
	case cover:
		block = lipgloss.JoinVertical(lipgloss.Left, bigTitle, underline)
	default:
		block = lipgloss.JoinVertical(
			lipgloss.Left,
			bigTitle,
			underline,
			"",
			bodyStyle.Render(s.body),
		)
	}

	canvas := lipgloss.NewStyle().
		Width(m.w).
		Height(m.h-1).
		Background(bg).
		Render(lipgloss.Place(m.w, m.h-1, lipgloss.Center, lipgloss.Center, block))

	return canvas + "\n" + m.footer()
}

func (m model) footer() string {
	counter := fmt.Sprintf("%d / %d", m.idx+1, len(slides))
	left := lipgloss.NewStyle().Foreground(muted).
		Render("←/→ navigate   ·   q quit   ·   r restart")
	right := lipgloss.NewStyle().Foreground(accent).Bold(true).Render(counter)

	gap := m.w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	bar := left + strings.Repeat(" ", gap) + right
	return lipgloss.NewStyle().Width(m.w).Background(bg).Render(bar)
}

func main() {
	opts := []tea.ProgramOption{tea.WithAltScreen()}
	if os.Getenv("NO_ALTSCREEN") != "" {
		opts = []tea.ProgramOption{}
	}
	p := tea.NewProgram(model{}, opts...)
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}