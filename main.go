package main

import (
	"flag"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spongeprojects/magicconch"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const (
	// PreviousKubeconfigFullPath is the file name which stores the previous kubeconfig file's full path
	PreviousKubeconfigFullPath = "previous"
)

type model struct {
	// meta is extra information displayed on top of the output
	meta []string

	// candidates is a list of (Candidate/kubeconfig)s
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
	configPath        = path.Join(kubeDir, "config") // same as defaultConfigPath for now, maybe allow user to specify
	cfDir             = path.Join(kubeDir, "kubectl-cf")
)

func (m *model) quit(farewell string) tea.Cmd {
	if !strings.HasSuffix(farewell, "\n") {
		farewell += "\n" // there must be a "\n" at the end of message
	}
	m.quitting = true
	m.farewell = farewell
	return tea.Quit
}

func (m *model) symlinkConfigPathTo(name string) string {
	magicconch.Must(os.WriteFile(path.Join(cfDir, PreviousKubeconfigFullPath), []byte(m.currentConfigPath), 0644))

	err := Symlink(name, configPath)
	if err != nil {
		return Warning(fmt.Sprintf("Symlink error: %s", err))
	}
	s := fmt.Sprintf("\n%s is now symlink to %s\n",
		Info(configPath), Info(name))
	kubeconfigEnv := os.Getenv("KUBECONFIG")
	if !(kubeconfigEnv == configPath || (configPath == defaultConfigPath && kubeconfigEnv == "")) {
		s += Warning(fmt.Sprintf("\nWARNING: You should set KUBECONFIG=%s to make it work.\n", configPath))
	}
	return s
}

func (m *model) Init() tea.Cmd {
	if len(flag.Args()) > 1 {
		return m.quit("Wrong number of arguments, expect 1")
	}

	candidates, err := ListKubeconfigCandidatesInDir(kubeDir)
	magicconch.Must(err)
	initialModel.candidates = candidates

	addDebugMessage("Path to config symlink: %s", configPath)

	info, err := os.Lstat(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		addDebugMessage("The symlink not exist, using the default kubeconfig: %s", defaultConfigPath)
		initialModel.currentConfigPath = defaultConfigPath
	} else {
		if IsSymlink(info) {
			target, err := os.Readlink(configPath)
			magicconch.Must(err)
			addDebugMessage("The symlink points to: %s", target)
			initialModel.currentConfigPath = target
		} else {
			addDebugMessage("The symlink is not a symlink")
			return m.quit(Warning(fmt.Sprintf("I'm sorry but %s must be a symlink to use kubectl-cf, "+
				"please move it to other place like ~/.kube/default.kubeconfig", configPath)))
		}
	}
	addDebugMessage("Current using kubeconfig: %s", initialModel.currentConfigPath)

	if debug {
		f, err := os.Open(path.Join(cfDir, PreviousKubeconfigFullPath))
		if err != nil {
			if !os.IsNotExist(err) {
				panic(err)
			}
			addDebugMessage("No previous kubeconfig")
		} else {
			b, err := ioutil.ReadAll(f)
			magicconch.Must(err)
			addDebugMessage("Previous kubeconfig: %s", string(b))
		}
	}

	if search := flag.Arg(0); search != "" {
		if search == "-" {
			f, err := os.Open(path.Join(cfDir, PreviousKubeconfigFullPath))
			if err != nil {
				if !os.IsNotExist(err) {
					panic(err)
				}
				return m.quit(Warning("No previous kubeconfig"))
			}
			b, err := ioutil.ReadAll(f)
			magicconch.Must(err)
			return m.quit(m.symlinkConfigPathTo(string(b)))
		}

		var guess []Candidate
		for _, candidate := range candidates {
			if candidate.Name == search {
				guess = []Candidate{candidate}
				break
			}
			if strings.HasPrefix(candidate.Name, search) {
				guess = append(guess, candidate)
			}
		}

		if guess == nil {
			return m.quit(Warning(fmt.Sprintf("No match found: %s", search)))
		}

		if len(guess) == 1 {
			return m.quit(m.symlinkConfigPathTo(guess[0].FullPath))
		}

		var s []string
		for _, g := range guess {
			s = append(s, g.Name)
		}
		return m.quit(Warning(fmt.Sprintf("More than 1 matches found: %s, can not determine: %s",
			search, strings.Join(s, ", "))))
	}

	// focus on current config path
	for key, candidate := range candidates {
		if candidate.FullPath == m.currentConfigPath {
			m.cursor = key
		}
	}

	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Is it a key press?
	case tea.KeyMsg:
		// The key pressed
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q", "esc":
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
			return m, m.quit(m.symlinkConfigPathTo(m.candidates[m.cursor].FullPath))
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	return m, nil
}

func (m *model) View() string {
	// The header
	s := strings.Join(m.meta, "\n") + "\n"

	if m.quitting {
		return s + m.farewell
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
		cursor := " "
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

func init() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), `Usage of kubectl-cf:

  cf           Select kubeconfig interactively
  cf [config]  Select kubeconfig directly
  cf -         Switch to the previous kubeconfig
`)
		flag.PrintDefaults()
	}

	flag.Parse()

	// ensure config dir exists
	if _, err := os.Lstat(cfDir); err != nil {
		if os.IsNotExist(err) {
			addDebugMessage("Default config dir %s not exist, creating", cfDir)
			magicconch.Must(os.Mkdir(cfDir, 0755))
		} else {
			panic(err)
		}
	}
}

func main() {
	p := tea.NewProgram(initialModel)
	magicconch.Must(p.Start())
}
