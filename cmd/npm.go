/*
Copyright © 2025 Jerome Duncan <jerome@jrmd.dev>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"jrmd.dev/qk/views"
	"os"
)

// cmdCmd represents the cmd command
var npmCmd = &cobra.Command{
	Use:     "npm",
	Aliases: []string{"n"},
	Short:   "run a npm command across all projects",
	Long:    `This command runs your npm command in all project folders`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Provide a command...")
			os.Exit(1)
		}

		m := views.CreateCommandRunner()
		m.
			AddCommand(RenderCommand("npm"), "npm", args...).
			Run()
	},
}

func init() {
	rootCmd.AddCommand(npmCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cmdCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cmdCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
