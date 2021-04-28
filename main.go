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
	"strings"
)

const (
	ModeDefault = ""    // replace soft link
	ModeEnv     = "env" // print env instead of replace soft link
)

type model struct {
	choices []Candidate
	cursor  int // indicates which choice our cursor is pointing at

	// current is the full path of current kubeconfig,
	// current only works when mode == ModeDefault
	current string

	// selected is a map from full path to a struct, indicates which choices are selected,
	// selected only works when mode == ModeEnv
	selected map[string]struct{}
}

var initialModel = model{
	choices:  make([]Candidate, 0),
	selected: make(map[string]struct{}),
}

var (
	mode = os.Getenv("CF_MODE")

	homeDir           = HomeDir()
	kubeDir           = path.Join(homeDir, ".kube")
	defaultConfigPath = path.Join(kubeDir, "config")
	cfDir             = path.Join(kubeDir, "kubectl-cf")
	configPath        = path.Join(cfDir, "config")
)

func init() {
	_, err := os.Lstat(cfDir)
	if err != nil {
		if os.IsNotExist(err) {
			magicconch.Must(os.Mkdir(cfDir, 0755))
		} else {
			magicconch.Must(err)
		}
	}

	choices, err := ListKubeconfigCandidates()
	magicconch.Must(err)
	initialModel.choices = choices

	debug("Running in mode: %s", mode)

	if mode == ModeEnv {
		selected := make(map[string]struct{})
		for _, c := range strings.Split(os.Getenv("KUBECONFIG"), ":") {
			selected[c] = struct{}{}
		}
		initialModel.selected = selected
	} else if mode == ModeDefault {
		info, err := os.Lstat(configPath)
		if err != nil {
			if !os.IsNotExist(err) {
				panic(err)
			}
			debug("Config %s not exist, using the default config: %s", configPath, defaultConfigPath)
			initialModel.current = defaultConfigPath
		} else {
			if info.Mode()&os.ModeSymlink == 0 {
				// is not a symlink
				debug("Config %s is not a symlink", configPath)
				initialModel.current = configPath
			} else {
				// is a symlink
				target, err := os.Readlink(configPath)
				magicconch.Must(err)
				debug("Config %s is a symlink to: %s", configPath, target)
				initialModel.current = target
			}
		}
		debug("Current config: %s", initialModel.current)
	}
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
			if mode == ModeEnv {
				choice := m.choices[m.cursor]
				fullPath := choice.FullPath
				if _, ok := m.selected[fullPath]; ok {
					delete(m.selected, fullPath)
				} else {
					m.selected[fullPath] = struct{}{}
				}
			}

		case "enter":
			if mode == ModeEnv {
				var ss []string
				for s := range m.selected {
					ss = append(ss, s)
				}
				kubeconfig := strings.Join(ss, ":")
				fmt.Printf("\nexport KUBECONFIG=%s\n", kubeconfig)
				return m, tea.Quit
			} else {
				err := Symlink(m.choices[m.cursor].FullPath, configPath)
				if err != nil {
					log.Fatal(errors.Wrap(err, "Symlink error"))
				}
				fmt.Printf("\n%s is now symlink to %s\n",
					termenv.String(configPath).Foreground(ColorProfile.Color("32")),
					termenv.String(m.choices[m.cursor].FullPath).Foreground(ColorProfile.Color("32")))
				if os.Getenv("KUBECONFIG") != configPath {
					s := termenv.String(fmt.Sprintf("\nWARNING: You should set KUBECONFIG=%s to make it work.\n", configPath))
					s = s.Foreground(ColorProfile.Color("1"))
					fmt.Println(s)
				}
				return m, tea.Quit
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
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

		prefix := ""
		if mode == ModeEnv {
			if _, ok := m.selected[candidate.FullPath]; ok {
				prefix = "[x] " // selected!
			} else {
				prefix = "[ ] " // not selected
			}
		} else {
			if candidate.FullPath == m.current {
				prefix = "* "
			} else {
				prefix = "  "
			}
		}

		s += fmt.Sprintf("%s %s%s\t%s\n", cursor, prefix, candidate.Name, candidate.FullPath)
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
