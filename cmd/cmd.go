/*
Copyright © 2025 Jerome Duncan <jerome@jrmd.dev>
*/
package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"jrmd.dev/qk/views"
	"os"
)

// cmdCmd represents the cmd command
var cmdCmd = &cobra.Command{
	Use:     "cmd",
	Aliases: []string{"c"},
	Short:   "run a custom command across all projects",
	Long:    `This command runs your custom command in all project folders`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Provide a command...")
			os.Exit(1)
		}

		c := args[0]
		arg := args[1:]

		m := views.CreateCommandRunner()
		m.AddCommand(RenderCommand(c), c, arg...)

		if _, err := tea.NewProgram(&m).Run(); err != nil {
			fmt.Println("could not run program:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(cmdCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cmdCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cmdCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
