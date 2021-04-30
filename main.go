package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spongeprojects/magicconch"
	"os"
	"path"
	"strings"
)

type model struct {
	// meta is extra information displayed on top of the window
	meta []string

	candidates []Candidate

	// cursor indicates which candidate our cursor is pointing at
	cursor int

	quitting bool

	// farewell is the message which will be printed before quitting
	farewell string

	// currentConfigPath is the full path of current kubeconfig
	currentConfigPath string
}

var initialModel = &model{}

var (
	homeDir           = HomeDir()
	kubeDir           = path.Join(homeDir, ".kube")
	defaultConfigPath = path.Join(kubeDir, "config")
	cfDir             = path.Join(kubeDir, "kubectl-cf")
	configPath        = path.Join(cfDir, "config")
)

func ensureCFDirExists() {
	if _, err := os.Lstat(cfDir); err != nil {
		if os.IsNotExist(err) {
			debug("Default config dir %s not exist, creating", cfDir)
			magicconch.Must(os.Mkdir(cfDir, 0755))
		} else {
			magicconch.Must(err)
		}
	}
}

func symlinkConfigPathTo(path string) string {
	err := Symlink(path, configPath)
	if err != nil {
		return Warning(fmt.Sprintf("Symlink error: %s", err))
	}
	s := fmt.Sprintf("\n%s is now symlink to %s\n",
		Info(configPath), Info(path))
	if os.Getenv("KUBECONFIG") != configPath {
		s += Warning(fmt.Sprintf("\nWARNING: You should set KUBECONFIG=%s to make it work.\n", configPath))
	}
	return s
}

func (m *model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	ensureCFDirExists()

	candidates, err := ListKubeconfigCandidatesInDir(kubeDir)
	magicconch.Must(err)
	initialModel.candidates = candidates

	if len(os.Args) > 1 && os.Args[1] != "" {
		var guess []Candidate
		for _, candidate := range candidates {
			if strings.HasPrefix(candidate.Name, os.Args[1]) {
				guess = append(guess, candidate)
			}
		}
		m.quitting = true
		if guess == nil {
			m.farewell = Warning(fmt.Sprintf("No match found: %s\n", os.Args[1]))
		} else if len(guess) == 1 {
			m.farewell = symlinkConfigPathTo(guess[0].FullPath)
		} else {
			var s []string
			for _, g := range guess {
				s = append(s, g.Name)
			}
			m.farewell = Warning(fmt.Sprintf("More than 1 matches found: %s, can not determine: %s\n",
				os.Args[1], strings.Join(s, ", ")))
		}
		// when tea.Quit is returned in Init, view cannot be rendered properly,
		// so we need to print the farewell message ourselves
		return tea.Quit
	}

	info, err := os.Lstat(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		debug("Config %s not exist, using the default config: %s", configPath, defaultConfigPath)
		initialModel.currentConfigPath = defaultConfigPath
	} else {
		if info.Mode()&os.ModeSymlink == 0 {
			// is not a symlink
			debug("Config %s is not a symlink", configPath)
			initialModel.currentConfigPath = configPath
		} else {
			// is a symlink
			target, err := os.Readlink(configPath)
			magicconch.Must(err)
			debug("Config %s is a symlink to: %s", configPath, target)
			initialModel.currentConfigPath = target
		}
	}
	debug("Current config: %s", initialModel.currentConfigPath)
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			} else {
				m.cursor = len(m.candidates) - 1
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.candidates)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}

		case "enter":
			m.quitting = true
			m.farewell = symlinkConfigPathTo(m.candidates[m.cursor].FullPath)
			return m, tea.Quit
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	return m, nil
}

func (m *model) View() string {
	if m.quitting {
		return m.farewell
	}

	// The header
	s := ""
	for _, meta := range m.meta {
		s += meta + "\n"
	}
	s += "What kubeconfig you want to use?\n\n"

	// Iterate over our candidates
	longestName := 0
	for _, candidate := range m.candidates {
		if len(candidate.Name) > longestName {
			longestName = len(candidate.Name)
		}
	}
	for key, candidate := range m.candidates {
		cursor := " " // no cursor
		if m.cursor == key {
			cursor = ">"
		}
		s += cursor

		suffix := ""
		if candidate.FullPath == m.currentConfigPath {
			suffix = "*"
		}
		tmpl := fmt.Sprintf(" %%-%ds %%s%%s\n", longestName)
		ts := fmt.Sprintf(tmpl, candidate.Name, candidate.FullPath, suffix)
		if candidate.FullPath == m.currentConfigPath {
			ts = Info(ts)
		}
		s += ts
	}

	// The footer
	s += Subtle("\nj/k, up/down: select • enter: choose • q, esc: quit\n")
	return s
}

func main() {
	p := tea.NewProgram(initialModel)
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
