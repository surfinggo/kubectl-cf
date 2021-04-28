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
	choices []Candidate
	cursor  int // indicates which choice our cursor is pointing at

	// current is the full path of current kubeconfig,
	// current only works when mode == ModeDefault
	current string
}

var initialModel = model{
	choices: make([]Candidate, 0),
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

		case "enter":
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
			ts := termenv.String(">") // cursor!
			ts.Blink()
			cursor = ts.String()
		}

		prefix := " "
		if candidate.FullPath == m.current {
			prefix = "*"
		}

		s += cursor
		es := termenv.String(fmt.Sprintf(" %s%s (%s)\n", prefix, candidate.Name, candidate.FullPath))
		if candidate.FullPath == m.current {
			es = es.Foreground(ColorProfile.Color("28"))
		}
		s += es.String()
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
