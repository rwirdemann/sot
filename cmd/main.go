package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/rwirdemann/sot"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	focusSidePanel = iota
	focusMainPanel = iota
	modeView
	modeEdit
)

type model struct {
	sidePanel sot.SidePanel
	focus     int
	mode      int
	content   map[string]string
	textarea  textarea.Model
	current   string
}

func initialModel() model {
	content := make(map[string]string)
	content["Journal"] = "Mit dem Hunde spazieren gehen\nUni-Kurse evaluieren"
	content["Dissertation"] = "Uni-Kurse evaluieren"
	content["Softwaredesign"] = "Deep Modules with small interfaces"

	ti := textarea.New()
	return model{
		sidePanel: sot.NewSidePanel(),
		focus:     focusSidePanel,
		content:   content,
		textarea:  ti,
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.focus == focusSidePanel {
				m.focus = focusMainPanel
				m.mode = modeEdit
				m.textarea.SetValue(m.content[m.sidePanel.Selected])
				cmds = append(cmds, m.textarea.Focus())
			}

			if m.focus == focusMainPanel && m.mode == modeView {
				m.mode = modeEdit
				m.textarea.SetValue(m.content[m.sidePanel.Selected])
				cmds = append(cmds, m.textarea.Focus())
			}
		case "esc":
			if m.focus == focusMainPanel {
				m.mode = modeView
				m.content[m.sidePanel.Selected] = m.textarea.Value()
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		}
		if m.focus == focusMainPanel {
			switch msg.String() {
			case "s":
				if m.mode == modeView {
					m.focus = focusSidePanel
				}
			}
		}
		if m.focus == focusSidePanel {
			m.sidePanel, cmd = m.sidePanel.Update(msg)
			if len(m.sidePanel.Selected) > 0 {
				m.current = m.content[m.sidePanel.Selected]
			}
			cmds = append(cmds, cmd)
		}
	}
	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		m.sidePanel.Render(m.focus == focusSidePanel),
		m.renderMainPanel())
}

func mainPanelSize() (int, int) {
	w, h, _ := term.GetSize(os.Stdout.Fd())
	return int(float32(w)*0.75) - 2, h - 2
}

func (m model) renderMainPanel() string {
	w, h := mainPanelSize()
	var style = lipgloss.NewStyle().
		Width(w).
		Height(h).
		BorderStyle(lipgloss.NormalBorder())
	if m.focus == focusMainPanel {
		style = style.BorderForeground(sot.ColorFocus)
		if m.mode == modeEdit {
			return style.Render(m.textarea.View())
		}
	}
	return style.Render(m.content[m.sidePanel.Selected])
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
