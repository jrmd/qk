/*
Copyright © 2025 Jerome Duncan <jerome@jrmd.dev>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"jrmd.dev/qk/views"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#969B86", Dark: "#696969"}
	highlight = lipgloss.AdaptiveColor{Light: "#dc8a78", Dark: "#dc8a78"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	errColor  = lipgloss.AdaptiveColor{Light: "#FF5555", Dark: "#FF5555"}

	subtleText    = lipgloss.NewStyle().Foreground(subtle)
	highlightText = lipgloss.NewStyle().Foreground(highlight)
	successText   = lipgloss.NewStyle().Foreground(special)
	errorText     = lipgloss.NewStyle().Foreground(errColor)
)

func RenderCommand(name string) func(views.Command) string {
	return func(c views.Command) string {
		stat := c.Status()
		status := stat
		switch stat {
		case "finished":
			status = successText.Render(stat)
		case "failed":
			status = errorText.Render(stat)
		}

		return fmt.Sprintf("%s %s", highlightText.Render(name), status)
	}
}

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:     "install",
	Aliases: []string{"i"},
	Short:   "runs yarn and composer install across all projects",
	Run: func(cmd *cobra.Command, args []string) {
		m := views.CreateCommandRunner()
		m.AddCommand(RenderCommand("yarn"), "yarn").
			AddCommand(RenderCommand("composer"), "composer", "install")

		if _, err := tea.NewProgram(m).Run(); err != nil {
			fmt.Println("could not run program:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
