package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"github.com/pkg/errors"
	"github.com/spongeprojects/magicconch"
	"log"
	"os"
	"path"
)

type model struct {
	// meta is extra information displayed on top of the window
	meta []string

	candidates []Candidate

	// cursor indicates which choice our cursor is pointing at
	cursor int

	confirmed bool

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

func (m *model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	ensureCFDirExists()

	candidates, err := ListKubeconfigCandidatesInDir(kubeDir)
	magicconch.Must(err)
	initialModel.candidates = candidates

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
			m.confirmed = true
			return m, tea.Quit
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	return m, nil
}

func (m *model) View() string {
	s := ""
	if m.confirmed {
		err := Symlink(m.candidates[m.cursor].FullPath, configPath)
		if err != nil {
			log.Fatal(errors.Wrap(err, "Symlink error"))
		}
		s += fmt.Sprintf("\n%s is now symlink to %s\n",
			termenv.String(configPath).Foreground(Info),
			termenv.String(m.candidates[m.cursor].FullPath).Foreground(Info))
		if os.Getenv("KUBECONFIG") != configPath {
			ts := termenv.String(fmt.Sprintf("\nWARNING: You should set KUBECONFIG=%s to make it work.\n", configPath))
			ts = ts.Foreground(Warning)
			s += ts.String()
		}
		return s
	}
	// The header
	for _, meta := range m.meta {
		s += meta + "\n"
	}
	s += "What kubeconfig you want to use?\n\n"

	// Iterate over our candidates
	for key, candidate := range m.candidates {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == key {
			ts := termenv.String(">") // cursor!
			ts.Blink()
			cursor = ts.String()
		}
		s += cursor

		suffix := ""
		if candidate.FullPath == m.currentConfigPath {
			suffix = "*"
		}
		ts := termenv.String(fmt.Sprintf(" %s%s (%s)\n", candidate.Name, suffix, candidate.FullPath))
		if candidate.FullPath == m.currentConfigPath {
			ts = ts.Foreground(Info)
		}

		s += ts.String()
	}

	// The footer
	s += "\nPress enter to confirm, press q to quit.\n"

	return s
}

func main() {
	p := tea.NewProgram(initialModel)
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
