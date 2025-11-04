/*
Copyright © 2025 Jerome Duncan <jerome@jrmd.dev>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"jrmd.dev/qk/utils"
	"jrmd.dev/qk/views"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:     "build",
	Aliases: []string{"b"},
	Short:   "Runs yarn build:prod across all projects",
	Run: func(cmd *cobra.Command, args []string) {
		depth, _ := cmd.Flags().GetInt("depth");
		joined, _ := cmd.Flags().GetBool("joined");
		m := views.CreateCommandRunner(depth, joined)
		m.
			AddOptionalCommand(utils.HasYarn, RenderCommand("yarn"), "yarn", "build:prod").
			AddOptionalCommand(utils.Not(utils.HasYarn), RenderCommand("npm"), "npm", "run", "build:prod").
			Run()
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolP("joined", "j", false, "Joined output")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// buildCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
