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
	fg     = lipgloss.Color("#FFFFFF") // pure white text
	muted  = lipgloss.Color("#9B9B9B") // dim grey
	accent = lipgloss.Color("#FF5C39") // Figma orange accent
)

const titleFont = "big" // FIGlet font for slide titles

// ---- Slides ----
type kind int

const (
	cover kind = iota
	content
)

type slide struct {
	kind  kind
	title string
	body  string // may contain multiple lines separated by "\n"
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

func fig(s, font string) string {
	return strings.TrimRight(figure.NewFigure(s, font, true).String(), "\n")
}

// bigText renders s as FIGlet art in the given font, wrapping word-by-word
// so it never exceeds maxWidth. Each wrapped group becomes a stacked block.
func bigText(s string, maxWidth int, font string) string {
	if maxWidth < 10 {
		return s
	}
	whole := fig(s, font)
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
		if visibleWidth(fig(try, font)) <= maxWidth {
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
		blocks[i] = fig(g, font)
	}
	return strings.Join(blocks, "\n")
}

func (m model) View() string {
	if m.w == 0 {
		return ""
	}

	s := slides[m.idx]
	maxW := m.w - 8
	avail := m.h - 1 // rows above the footer

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(fg)
	// Body is plain regular text (bold for a little extra weight).
	// A terminal can't enlarge real glyphs from inside the app — zoom the
	// terminal itself (Ctrl + =) if you want the body text physically bigger.
	bodyStyle := lipgloss.NewStyle().Bold(true).Foreground(fg)
	barStyle := lipgloss.NewStyle().Foreground(accent)

	bigTitle := titleStyle.Render(bigText(s.title, maxW, titleFont))
	barWidth := visibleWidth(bigTitle)
	if barWidth > maxW {
		barWidth = maxW
	}
	underline := barStyle.Render(strings.Repeat("━", barWidth))

	var block string
	if s.kind == cover {
		block = lipgloss.JoinVertical(lipgloss.Left, bigTitle, underline)
	} else {
		block = lipgloss.JoinVertical(
			lipgloss.Left,
			bigTitle,
			underline,
			"",
			bodyStyle.Render(s.body),
		)
	}

	// No background fill: empty cells stay transparent so the terminal's
	// own background (image / acrylic / theme) shows through.
	canvas := lipgloss.NewStyle().
		Width(m.w).
		Height(avail).
		Render(lipgloss.Place(m.w, avail, lipgloss.Center, lipgloss.Center, block))

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
	return lipgloss.NewStyle().Width(m.w).Render(bar)
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