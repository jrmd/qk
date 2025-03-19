/*
Copyright © 2025 Jerome Duncan <jerome@jrmd.dev>
*/
package views

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"jrmd.dev/qk/utils"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/stopwatch"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	normal    = lipgloss.Color("#EEEEEE")
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	errColor  = lipgloss.AdaptiveColor{Light: "#FF5555", Dark: "#FF5555"}
	accent    = lipgloss.AdaptiveColor{Light: "#04a5e5", Dark: "#04a5e5"}

	title = lipgloss.NewStyle().
		Padding(0, 1).
		Bold(true).
		Foreground(normal).
		Background(highlight)

	subtitle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(accent)

	divider = lipgloss.NewStyle().
		SetString("•").
		Padding(0, 1).
		Foreground(subtle).
		String()

	checkMark = lipgloss.NewStyle().SetString("✓").
			Foreground(special).
			PaddingRight(1).
			String()
	cross = lipgloss.NewStyle().SetString("x").
		Foreground(errColor).
		PaddingRight(1).
		String()

	projectDone = func(s string) string {
		return lipgloss.NewStyle().
			Strikethrough(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#969B86", Dark: "#696969"}).
			Render(s)
	}

	projectStyle = func(s string) string {
		return lipgloss.NewStyle().
			Foreground(accent).
			Render(s)
	}
)

type keyMap struct {
	Scripts key.Binding
	Timer   key.Binding
	Help    key.Binding
	Quit    key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Scripts, k.Timer}, // first column
		{k.Help, k.Quit},     // second column
	}
}

var keys = keyMap{
	Scripts: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "toggle scripts"),
	),
	Timer: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "toggle timer"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type commandFinishedMessage struct {
	index       int
	scriptIndex int
	err         error
}
type programDoneMessage struct {
	success bool
	err     error
}

func runCommand(projIndex int, project Project, scriptIndex int, command string, args ...string) tea.Cmd {
	c := exec.Command(command, args...) //nolint:gosec
	c.Dir = project.dir
	c.Stdout = nil
	return func() tea.Msg {
		err := c.Run()
		return commandFinishedMessage{projIndex, scriptIndex, err}
	}
}

func done(success bool) tea.Cmd {
	return func() tea.Msg {
		return programDoneMessage{success, nil}
	}
}

type Command struct {
	script string
	args   []string
	status string
	render func(Command) string
}

func (c Command) Status() string {
	return c.status
}

type Project struct {
	spinner spinner.Model
	name    string
	dir     string
	scripts []Command
}

type model struct {
	projects      []Project
	start         time.Time
	finish        time.Time
	done          bool
	keys          keyMap
	help          help.Model
	stopwatch     stopwatch.Model
	showStopwatch bool
	showScripts   bool
}

func CreateCommandRunner() model {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	projects := utils.GetAllProjects(wd, 0)

	projs := []Project{}

	for _, project := range projects {
		s := spinner.New()
		s.Spinner = spinner.Dot
		s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		projs = append(projs, Project{
			s,
			project.Name,
			project.Dir,
			[]Command{},
		})
	}

	conf := utils.GetConfig()

	return model{
		projects:      projs,
		start:         time.Now(),
		finish:        time.Now(),
		done:          false,
		stopwatch:     stopwatch.NewWithInterval(time.Millisecond),
		keys:          keys,
		help:          help.New(),
		showStopwatch: conf.ShowTimer,
		showScripts:   conf.ShowScripts,
	}
}

func (m model) AddCommand(render func(Command) string, script string, args ...string) model {
	cmd := Command{script, args, "running", render}
	for i := range m.projects {
		m.projects[i].scripts = append(m.projects[i].scripts, cmd)
	}
	return m
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.stopwatch.Init(),
	}
	for i, proj := range m.projects {
		cmds = append(cmds, proj.spinner.Tick)
		for j, script := range proj.scripts {
			cmds = append(
				cmds,
				runCommand(
					i,
					proj,
					j,
					script.script,
					script.args...,
				),
			)

		}
	}
	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var stopwatchCmd tea.Cmd
	m.stopwatch, stopwatchCmd = m.stopwatch.Update(msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Scripts):
			m.showScripts = !m.showScripts
		case key.Matches(msg, m.keys.Timer):
			m.showStopwatch = !m.showStopwatch
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
		return m, stopwatchCmd
	case stopwatch.TickMsg:
		m.stopwatch, stopwatchCmd = m.stopwatch.Update(msg)
		return m, stopwatchCmd
	case spinner.TickMsg:
		cmds := []tea.Cmd{stopwatchCmd}
		for i, proj := range m.projects {
			var cmd tea.Cmd
			m.projects[i].spinner, cmd = proj.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	case commandFinishedMessage:
		status := "finished"
		if msg.err != nil {
			status = "failed"
		}

		m.projects[msg.index].scripts[msg.scriptIndex].status = status
		success := true

		if utils.Some(m.projects, func(project Project) bool {
			return utils.Some(project.scripts, func(script Command) bool {
				return script.status == "running"
			})
		}) {
			return m, nil
		}

		if utils.Some(m.projects, func(project Project) bool {
			return utils.Some(project.scripts, func(script Command) bool {
				return script.status == "failed"
			})
		}) {
			success = false
		}

		m.done = true

		return m, tea.Batch(done(success), stopwatchCmd)
	case programDoneMessage:
		return m, tea.Quit

	default:
		return m, stopwatchCmd
	}
}

func (m model) View() (s string) {
	gap := " "

	s += fmt.Sprintf("%s  %s\n\n", title.Render("QK Command Runner"), subtitle.Render("v0.1.0"))

	for _, proj := range m.projects {
		allFinished := utils.All(proj.scripts, func(script Command) bool {
			return script.status == "failed" || script.status == "finished"
		})
		hasError := utils.Some(proj.scripts, func(script Command) bool {
			return script.status == "failed"
		})
		spin := proj.spinner.View()

		if hasError {
			spin = cross
		} else if allFinished {
			spin = checkMark
		}

		name := projectStyle(proj.name)
		if allFinished && !hasError {
			name = projectDone(proj.name)
		}

		s += fmt.Sprintf("%s%s%s\n", spin, gap, name)
		if (!allFinished || hasError) && (m.showScripts || m.done) {
			for i, script := range proj.scripts {
				if i > 0 {
					s += divider
				}
				s += fmt.Sprintf("   %s", script.render(script))
			}
			s += "\n"
		}
	}
	if m.done {
		s += fmt.Sprintf("\nFinished in %s\n", time.Since(m.start))
	} else if m.showStopwatch {
		s += fmt.Sprintf("Elapsed: %s\n", m.stopwatch.View())
	}

	if !m.done {
		s += m.help.View(m.keys)
	}

	return s
}
