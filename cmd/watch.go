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
		depth, _ := cmd.Flags().GetInt("depth");
		joined, _ := cmd.Flags().GetBool("joined");
		m := views.CreateCommandRunner(depth, joined)

		m.
			AddOptionalCommand(
			utils.And(
				utils.HasYarn,
					utils.HasScript("start"),
					utils.Not(utils.HasScript("watch:dev")),
					utils.Not(utils.HasScript("dev")),
				),
				RenderCommand("yarn"),
				"yarn",
				"start",
			).
			AddOptionalCommand(
				utils.And(
					utils.Not(utils.HasYarn),
					utils.HasScript("start"),
					utils.Not(utils.HasScript("watch:dev")),
					utils.Not(utils.HasScript("dev")),
				),
				RenderCommand("npm"),
				"npm",
				"run",
				"start",
			).
			AddOptionalCommand(
			utils.And(
				utils.HasYarn,
					utils.HasScript("watch:dev"),
				),
				RenderCommand("yarn"),
				"yarn",
				"watch:dev",
			).
			AddOptionalCommand(
				utils.And(
					utils.Not(utils.HasYarn),
					utils.HasScript("watch:dev"),
				),
				RenderCommand("npm"),
				"npm",
				"run",
				"watch:dev",
			).
			AddOptionalCommand(
			utils.And(
				utils.HasYarn,
					utils.HasScript("dev"),
				),
				RenderCommand("yarn"),
				"yarn",
				"dev",
			).
			AddOptionalCommand(
				utils.And(
					utils.Not(utils.HasYarn),
					utils.HasScript("dev"),
				),
				RenderCommand("npm"),
				"npm",
				"run",
				"dev",
			).
			Run()
	},
}

func init() {
	rootCmd.AddCommand(watchCommand)
	watchCommand.Flags().BoolP("joined", "j", true, "Joined output")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// buildCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
