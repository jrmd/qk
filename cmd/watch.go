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
var watchCommand = &cobra.Command{
	Use:     "watch",
	Aliases: []string{"w"},
	Short:   "Runs yarn start across all projects",
	Run: func(cmd *cobra.Command, args []string) {
		m := views.CreateCommandRunner()
		m.
			AddOptionalCommand(utils.HasYarn, RenderCommand("yarn"), "yarn", "start").
			AddOptionalCommand(utils.Not(utils.HasYarn), RenderCommand("npm"), "npm", "run", "start").
			Run()
	},
}

func init() {
	rootCmd.AddCommand(watchCommand)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// buildCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
