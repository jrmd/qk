/*
Copyright © 2025 Jerome Duncan <jerome@jrmd.dev>
*/
package views

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"jrmd.dev/qk/types"
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
	Debug   key.Binding
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
		{k.Debug, k.Scripts, k.Timer}, // first column
		{k.Help, k.Quit},              // second column
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
	Debug: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "toggle debug"),
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

type commandOutputMessage struct {
	index       int
	scriptIndex int
	output      string
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

func runCommand(ctx context.Context, wg *sync.WaitGroup, program *tea.Program, projIndex int, project types.Project, scriptIndex int, command *types.Command) tea.Cmd {
	return func() tea.Msg {
		defer wg.Done()

		c := exec.CommandContext(ctx, command.Script, command.Args...)
		c.Dir = project.Dir
		c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		stdout, err := c.StdoutPipe()
		if err != nil {
			return commandFinishedMessage{projIndex, scriptIndex, err}
		}

		stderr, err := c.StderrPipe()
		if err != nil {
			return commandFinishedMessage{projIndex, scriptIndex, err}
		}

		if err := c.Start(); err != nil {
			return commandFinishedMessage{projIndex, scriptIndex, err}
		}

		pid := c.Process.Pid

		// Start goroutines to stream output
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				select {
				case <-ctx.Done():
					return
				default:
					line := scanner.Text()
					command.Output.WriteString(line + "\n")
					// Send the message to the program
					program.Send(commandOutputMessage{projIndex, scriptIndex, line})
				}
			}
		}()

		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				select {
				case <-ctx.Done():
					return
				default:
					line := scanner.Text()
					command.Output.WriteString(line + "\n")
					// Send the message to the program
					program.Send(commandOutputMessage{projIndex, scriptIndex, line})
				}
			}
		}()

		// Handle process termination
		waitChan := make(chan error, 1)
		go func() {
			select {
			case <-ctx.Done():
				_ = syscall.Kill(-pid, syscall.SIGTERM)
				time.Sleep(100 * time.Millisecond)
				_ = syscall.Kill(-pid, syscall.SIGKILL)
				waitChan <- ctx.Err()
			case errWait := <-waitChan:
				waitChan <- errWait
				return
			}
		}()

		errWait := c.Wait()
		waitChan <- errWait
		finalErr := <-waitChan

		return commandFinishedMessage{projIndex, scriptIndex, finalErr}
	}
}

// Function to check if an error indicates a signal kill
func wasKilledBySignal(err error) (bool, syscall.Signal) {
	if err == nil {
		return false, 0 // No error, wasn't killed
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		// Error is an ExitError, now check the process state
		status, ok := exitErr.Sys().(syscall.WaitStatus)
		if !ok {
			// This should not happen on Unix-like systems if it's an ExitError
			// Might happen on Windows or other OSes where Sys() has a different type
			// Fallback: Check if ExitCode is -1, often indicates signal on Unix
			// or abnormal termination elsewhere. Less reliable than WaitStatus.
			if exitErr.ProcessState != nil && exitErr.ProcessState.ExitCode() == -1 {
				return true, 0 // Indicate killed, but signal unknown
			}
			return false, 0
		}

		// Check if the process was signaled
		if status.Signaled() {
			return true, status.Signal() // Return true and the specific signal
		}
	}

	// Error is not an ExitError or process exited normally (even if non-zero)
	return false, 0
}

func done(success bool) tea.Cmd {
	return func() tea.Msg {
		return programDoneMessage{success, nil}
	}
}

type model struct {
	program       *tea.Program
	projects      []types.Project
	liveOutput    map[string][]string // key: "projIndex-scriptIndex"
	start         time.Time
	finish        time.Time
	done          bool
	keys          keyMap
	help          help.Model
	stopwatch     stopwatch.Model
	showStopwatch bool
	showScripts   bool
	showStdout    bool
	ctx           context.Context
	cancel        context.CancelFunc
	cmdWg         sync.WaitGroup // Add WaitGroup to track running commands
}

func CreateCommandRunner() model {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	projects := utils.GetAllProjects(wd, 0)

	if len(projects) == 0 {
		fmt.Println(lipgloss.NewStyle().Foreground(errColor).Render("Error: no projects found!"))
		os.Exit(1)
	}

	projs := []types.Project{}

	for _, project := range projects {
		s := spinner.New()
		s.Spinner = spinner.Dot
		s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		projs = append(projs, types.Project{
			Spinner: s,
			Name:    project.Name,
			Dir:     project.Dir,
			Scripts: []*types.Command{},
		})
	}

	conf := utils.GetConfig()
	ctx, cancel := context.WithCancel(context.Background())
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
		showStdout:    conf.ShowStdout,
		ctx:           ctx,
		cancel:        cancel,
		liveOutput:    make(map[string][]string),
	}
}

func (m *model) SetProgram(p *tea.Program) *model {
	m.program = p
	return m
}

func (m *model) Run() {
	p := tea.NewProgram(m)
	m.SetProgram(p)

	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}

	fmt.Print(m.Output(0))
}

func (m *model) AddCommand(render func(*types.Command) string, script string, args ...string) *model {
	for i := range m.projects {
		ctx, cancel := context.WithCancel(context.Background())
		cmd := &types.Command{Script: script, Args: args, Status: "running", Ctx: ctx, Cancel: cancel, Output: bytes.NewBuffer([]byte{}), Render: render, Reader: nil}
		m.projects[i].Scripts = append(m.projects[i].Scripts, cmd)
	}
	return m
}

func (m *model) AddOptionalCommand(shouldAdd func(types.Project) bool, render func(*types.Command) string, script string, args ...string) *model {
	for i, proj := range m.projects {
		if shouldAdd(proj) {
			ctx, cancel := context.WithCancel(context.Background())
			cmd := &types.Command{Script: script, Args: args, Status: "running", Ctx: ctx, Cancel: cancel, Output: bytes.NewBuffer([]byte{}), Render: render, Reader: nil}

			m.projects[i].Scripts = append(m.projects[i].Scripts, cmd)
		}
	}
	return m
}

func (m *model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.stopwatch.Init(),
	}
	for i, proj := range m.projects {
		cmds = append(cmds, proj.Spinner.Tick)
		for j, script := range proj.Scripts {
			m.cmdWg.Add(1)
			cmds = append(
				cmds,
				runCommand(
					script.Ctx,
					&m.cmdWg,
					m.program,
					i,
					proj,
					j,
					m.projects[i].Scripts[j],
				),
			)

		}
	}
	return tea.Batch(cmds...)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var stopwatchCmd tea.Cmd
	m.stopwatch, stopwatchCmd = m.stopwatch.Update(msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Scripts):
			m.showScripts = !m.showScripts
		case key.Matches(msg, m.keys.Timer):
			m.showStopwatch = !m.showStopwatch
		case key.Matches(msg, m.keys.Debug):
			m.showStdout = !m.showStdout
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Quit):
			m.CancelScripts()
			m.cmdWg.Wait()
			return m, tea.Quit
		}
		return m, stopwatchCmd
	case spinner.TickMsg:
		cmds := []tea.Cmd{stopwatchCmd}
		for i, proj := range m.projects {
			var cmd tea.Cmd
			m.projects[i].Spinner, cmd = proj.Spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	case commandFinishedMessage:
		status := "finished"
		if msg.err != nil {
			status = "failed"

			wasKilled, _ := wasKilledBySignal(msg.err)

			if wasKilled {
				status = "exited"
			}
		}

		m.projects[msg.index].Scripts[msg.scriptIndex].Status = status
		success := true
		m.done = true

		if utils.Some(m.projects, func(project types.Project) bool {
			return utils.Some(project.Scripts, func(script *types.Command) bool {
				return script.Status == "running"
			})
		}) {
			m.done = false
			return m, nil
		}

		if utils.Some(m.projects, func(project types.Project) bool {
			return utils.Some(project.Scripts, func(script *types.Command) bool {
				return script.Status == "failed"
			})
		}) {
			success = false
		}

		if !m.done {
			return m, stopwatchCmd
		}

		return m, tea.Batch(done(success), stopwatchCmd)
	case programDoneMessage:
		m.CancelScripts()
		return m, tea.Quit
	case commandOutputMessage:
		key := fmt.Sprintf("%d-%d", msg.index, msg.scriptIndex)
		if m.liveOutput[key] == nil {
			m.liveOutput[key] = []string{}
		}
		m.liveOutput[key] = append(m.liveOutput[key], msg.output)

		// Keep only last N lines to prevent memory issues
		maxLines := 50
		if len(m.liveOutput[key]) > maxLines {
			m.liveOutput[key] = m.liveOutput[key][len(m.liveOutput[key])-maxLines:]
		}

		return m, stopwatchCmd
	default:
		return m, stopwatchCmd
	}
}

func (m *model) CancelScripts() {
	for _, p := range m.projects {
		for _, c := range p.Scripts {
			c.Cancel()
		}
	}

}

func (m *model) Output(maxLines int) (s string) {
	gap := " "

	s += fmt.Sprintf("%s  %s\n\n", title.Render("QK Command Runner"), subtitle.Render("v0.1.0"))

	for i, proj := range m.projects {
		allFinished := utils.All(proj.Scripts, func(script *types.Command) bool {
			return script.Status == "failed" || script.Status == "finished"
		})

		hasError := utils.Some(proj.Scripts, func(script *types.Command) bool {
			return script.Status == "failed"
		})
		spin := proj.Spinner.View()

		if hasError {
			spin = cross
		} else if allFinished {
			spin = checkMark
		}

		name := projectStyle(proj.Name)
		if allFinished && !hasError {
			name = projectDone(proj.Name)
		}

		s += fmt.Sprintf("%s%s%s\n", spin, gap, name)

		if ((!allFinished || hasError) && (m.showScripts || m.done)) || m.showStdout {
			for j, script := range proj.Scripts {
				if m.done || m.showScripts {
					if j > 0 {
						s += divider
					}
					s += fmt.Sprintf("   %s", script.Render(script))
				}

				// Show live output if debug mode is on
				if m.showStdout {
					key := fmt.Sprintf("%d-%d", i, j)
					stdOut := ""
					if output, exists := m.liveOutput[key]; exists && len(output) > 0 {
						data := output
						if maxLines > 0 && len(data) > maxLines {
							data = output[len(data)-maxLines:]
						}

						for _, line := range data {
							stdOut += fmt.Sprintf("     %s\n",
								lipgloss.NewStyle().
									Foreground(normal).
									Render(line))
						}
					}

					if len(stdOut) > 0 {
						s += "\n"
						s += stdOut
					}
				}
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

func (m *model) View() (s string) {
	if m.done {
		return s
	}

	return m.Output(10)
}
