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
var yarnCmd = &cobra.Command{
	Use:     "yarn",
	Aliases: []string{"y"},
	Short:   "run a yarn command across all projects",
	Long:    `This command runs your yarn command in all project folders`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Provide a command...")
			os.Exit(1)
		}

		depth, _ := cmd.Flags().GetInt("depth");
		joined, _ := cmd.Flags().GetBool("joined");

		m := views.CreateCommandRunner(depth, joined)
		m.
			AddCommand(RenderCommand("yarn"), "yarn", args...).
			Run()
	},
}

func init() {
	rootCmd.AddCommand(yarnCmd)
	yarnCmd.Flags().BoolP("joined", "j", false, "Joined output")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cmdCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cmdCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
