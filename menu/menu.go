package menu

import (
	"fmt"
	"os"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	choices []string
	cursor  int
	width   int
	height  int
}

var selectedServer string

var randomColor lipgloss.Color

func (m model) Init() tea.Cmd {

	if os.Getenv("TERM") == "" {
		fmt.Println("This program requires a terminal that supports ANSI escape codes.")
		return nil
	}
	return nil

}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			selectedServer = m.choices[m.cursor]
			return m, tea.Quit
		}
	}

	return m, nil

}

func (m model) View() string {

	s := ""
	randomColor = randomCyberpunkColor()

	for i, choice := range m.choices {
		var style lipgloss.Style
		if m.cursor == i {
			choice = fmt.Sprintf("[ %s ]", choice)
			style = lipgloss.NewStyle().
				Background(randomColor).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)
		} else {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000"))
		}
		s += style.Render(choice) + "\n"
	}

	menuBlock := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(s)

	emptySpace := (m.height - len(m.choices)) / 2

	finalView := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(emptySpace, 0, emptySpace, 0).
		Render(menuBlock)

	return finalView

}

func ShowMenu(sshFiles map[string]string) string {

	var serverNames []string

	for name, _ := range sshFiles {
		serverNames = append(serverNames, name)
	}

	sort.Strings(serverNames)

	exitItem := "x"

	serverNames = append(serverNames, exitItem)

	m := &model{
		choices: serverNames,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}

	if exitItem == selectedServer {
		selectedServer = ""
	}
	return selectedServer

}
