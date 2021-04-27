package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spongeprojects/magicconch"
	"os"
	"strings"
)

type model struct {
	choices  []Candidate
	cursor   int                 // indicates which choice our cursor is pointing at
	selected map[string]struct{} // a map from full path to a struct, indicates which choices are selected
}

var initialModel = model{
	choices:  make([]Candidate, 0),
	selected: make(map[string]struct{}),
}

func init() {
	choices, err := ListKubeconfigCandidates()
	magicconch.Must(err)
	initialModel.choices = choices

	selected := make(map[string]struct{})
	for _, c := range strings.Split(os.Getenv("KUBECONFIG"), ":") {
		selected[c] = struct{}{}
	}
	initialModel.selected = selected
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:
		// The key pressed
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The spacebar (a literal space) toggle the selected state for the item that the cursor is pointing at.
		case " ":
			if m.cursor >= 0 && m.cursor < len(m.choices) {
				choice := m.choices[m.cursor]
				fullPath := choice.FullPath
				if _, ok := m.selected[fullPath]; ok {
					delete(m.selected, fullPath)
				} else {
					m.selected[fullPath] = struct{}{}
				}
			}

		case "enter":
			var ss []string
			for s := range m.selected {
				ss = append(ss, s)
			}
			kubeconfig := strings.Join(ss, ":")
			fmt.Printf("export KUBECONFIG=%s\n", kubeconfig)
			magicconch.Must(os.Setenv("KUBECONFIG", kubeconfig))
			return m, tea.Quit
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	// The header
	s := "What kubeconfig you want to use?\n\n"

	// Iterate over our choices
	for key, candidate := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == key {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[candidate.FullPath]; ok {
			checked = "x" // selected!
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s\t%s\n", cursor, checked, candidate.Name, candidate.FullPath)
	}

	// The footer
	s += "\nPress enter to confirm, press q to quit.\n"

	// Send the UI for rendering
	return s
}

func main() {
	p := tea.NewProgram(initialModel)
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
