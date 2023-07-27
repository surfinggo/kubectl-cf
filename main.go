package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spongeprojects/magicconch"
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
	kubeDir           = filepath.Join(homeDir, ".kube")
	defaultConfigPath = filepath.Join(kubeDir, "config")
	configPath        = filepath.Join(kubeDir, "config") // same as defaultConfigPath for now, maybe allow user to specify
	cfDir             = filepath.Join(kubeDir, "kubectl-cf")
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
	magicconch.Must(os.WriteFile(filepath.Join(cfDir, PreviousKubeconfigFullPath), []byte(m.currentConfigPath), 0644))

	err := Symlink(name, configPath)
	if err != nil {
		return Warning(t("createSymlinkError", err))
	}
	s := t("symlinkNowPointTo", Info(configPath), Info(name))
	kubeconfigEnv := os.Getenv("KUBECONFIG")
	if !(kubeconfigEnv == configPath || (configPath == defaultConfigPath && kubeconfigEnv == "")) {
		s += "\n" + Warning(t("kubeconfigEnvWarning", configPath))
	}
	return s
}

func (m *model) Init() tea.Cmd {
	if len(flag.Args()) > 1 {
		return m.quit(t("wrongNumberOfArgumentExpect1"))
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
			return m.quit(Warning(t("kubeconfigNotSymlink", configPath)))
		}
	}
	addDebugMessage("Current using kubeconfig: %s", initialModel.currentConfigPath)

	if debug {
		f, err := os.Open(filepath.Join(cfDir, PreviousKubeconfigFullPath))
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
			f, err := os.Open(filepath.Join(cfDir, PreviousKubeconfigFullPath))
			if err != nil {
				if !os.IsNotExist(err) {
					panic(err)
				}
				return m.quit(Warning(t("noPreviousKubeconfig")))
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
			return m.quit(Warning(t("noMatchFound", search)))
		}

		if len(guess) == 1 {
			return m.quit(m.symlinkConfigPathTo(guess[0].FullPath))
		}

		var s []string
		for _, g := range guess {
			s = append(s, g.Name)
		}
		return m.quit(Warning(t("moreThanOneMatchesFound", search, strings.Join(s, ", "))))
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

	s += t("whatKubeconfig") + "\n\n"

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
	s += Subtle("\n" + t("helpActions") + "\n")
	return s
}

func init() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), t("cfUsage"))
		flag.PrintDefaults()
	}

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
	flag.Parse()

	p := tea.NewProgram(initialModel)
	magicconch.Must(p.Start())
}
