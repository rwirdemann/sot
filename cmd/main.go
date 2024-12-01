package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"io"
	"log"
	"log/slog"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

var ColorFocus = lipgloss.Color("12")

const (
	focusSidePanel = iota
	focusMainPanel = iota
)

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(2)
)

type file struct {
	name    string
	title   string
	content string
}

type item string

func (i item) FilterValue() string {
	return ""
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	_, err := fmt.Fprint(w, fn(fmt.Sprintf("%s", i)))
	if err != nil {
		slog.Error(err.Error())
	}
}

type model struct {
	focus    int
	mode     int
	cursor   int
	textarea textarea.Model
	list     list.Model
	files    []file
}

func initialModel() model {
	const defaultWidth = 20
	const listHeight = 14

	files := loadJournal()
	items := make([]list.Item, 0, len(files))
	for _, f := range files {
		items = append(items, item(f.title))
	}
	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.Styles.PaginationStyle = paginationStyle

	ti := textarea.New()
	ti.SetWidth(80)
	return model{
		focus:    focusSidePanel,
		textarea: ti,
		list:     l,
		files:    files,
	}
}

func loadJournal() []file {
	const base = "/Users/ralfwirdemann/Library/Mobile Documents/iCloud~com~logseq~logseq/Documents/Zettelkasten/journals"
	files, err := os.ReadDir(base)
	if err != nil {
		log.Fatal(err)
	}

	var ff []file
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
			t, err := time.Parse(time.DateOnly, strings.ReplaceAll(strings.TrimSuffix(f.Name(), ".md"), "_", "-"))
			if err != nil {
				log.Fatal(err)
			}
			content, err := os.ReadFile(path.Join(base, f.Name()))
			if err != nil {
				log.Fatal(err)
			}
			title := t.Format("Mon, 02 Jan 2006")
			ff = append(ff, file{
				name:    f.Name(),
				title:   title,
				content: string(content),
			})
		}
	}

	sort.Slice(ff, func(i, j int) bool {
		return ff[i].name > ff[j].name
	})
	return ff
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
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderFilePanel(),
		m.renderMainPanel())
}

func mainPanelSize() (int, int) {
	w, h, _ := term.GetSize(os.Stdout.Fd())
	return int(float32(w)*0.82) - 2, h - 2
}

func filePanelSize() (int, int) {
	w, h, _ := term.GetSize(os.Stdout.Fd())
	return int(float32(w)*0.18) - 2, h - 2
}

func (m model) renderFilePanel() string {
	w, h := filePanelSize()
	var style = lipgloss.NewStyle().
		Width(w).
		Height(h).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	if m.focus == focusSidePanel {
		style = style.BorderForeground(ColorFocus)
	}
	m.list.SetHeight(h)

	return style.Render(m.list.View())
}

func (m model) renderMainPanel() string {
	w, h := mainPanelSize()
	var style = lipgloss.NewStyle().
		Width(w).
		Height(h).
		BorderStyle(lipgloss.NormalBorder())
	if m.focus == focusMainPanel {
		style = style.BorderForeground(ColorFocus)
	}

	if len(m.files) == 0 {
		return style.Render("this is where the file content is shown")
	}

	s := m.files[m.list.Cursor()].content
	return style.Render(s)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an errorq: %v", err)
		os.Exit(1)
	}
}
