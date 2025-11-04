/*
Copyright © 2025 Jerome Duncan <jerome@jrmd.dev>
*/
package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"jrmd.dev/qk/types"
	"jrmd.dev/qk/utils"
	"jrmd.dev/qk/views"
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

func RenderCommand(name string) func(*types.Command, bool) string {
	return func(c *types.Command, showStatus bool) string {
		if !showStatus {
			return highlightText.Render(name)
		}

		stat := c.Status
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
		depth, _ := cmd.Flags().GetInt("depth");
		joined, _ := cmd.Flags().GetBool("joined");

		m := views.CreateCommandRunner(depth, joined)
		m.
			AddOptionalCommand(utils.HasYarn, RenderCommand("yarn"), "yarn").
			AddOptionalCommand(utils.Not(utils.HasYarn), RenderCommand("npm"), "npm", "install").
			AddCommand(RenderCommand("composer"), "composer", "install").
			Run()
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().BoolP("joined", "j", false, "Joined output")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
