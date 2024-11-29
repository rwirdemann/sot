package sot

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"os"
)

var ColorFocus = lipgloss.Color("12")

type SidePanel struct {
	cursor   int
	entries  []string
	Selected string
}

func NewSidePanel() SidePanel {
	return SidePanel{cursor: 0, entries: []string{"Journal", "Dissertation", "Softwaredesign"}, Selected: "Journal"}
}

func (p SidePanel) Render(focus bool) string {
	w, h := size()
	var style = lipgloss.NewStyle().
		Width(w).
		Height(h).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	if focus {
		style = style.BorderForeground(ColorFocus)
	}

	s := ""
	for i, e := range p.entries {
		if i == p.cursor {
			s = fmt.Sprintf("%s> %s\n", s, e)
		} else {
			s = fmt.Sprintf("%s  %s\n", s, e)
		}
	}
	return style.Render(s)
}

func (p SidePanel) Update(msg tea.KeyMsg) (SidePanel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if p.cursor > 0 {
			p.cursor--
			p.Selected = p.entries[p.cursor]
		}
	case "down", "j":
		if p.cursor < len(p.entries)-1 {
			p.cursor++
			p.Selected = p.entries[p.cursor]
		}
	}
	return p, nil
}

func size() (int, int) {
	w, h, _ := term.GetSize(os.Stdout.Fd())
	return int(float32(w)*0.25) - 2, h - 2
}
