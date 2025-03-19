/*
Copyright © 2025 Jerome Duncan <jerome@jrmd.dev>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
	"jrmd.dev/qk/utils"
)

var (
    purple    = lipgloss.Color("99")
    gray      = lipgloss.Color("245")
    lightGray = lipgloss.Color("241")

    headerStyle  = lipgloss.NewStyle().Foreground(purple).Bold(true).Align(lipgloss.Center)
    cellStyle    = lipgloss.NewStyle().Padding(0, 1)
    oddRowStyle  = cellStyle.Foreground(gray)
    evenRowStyle = cellStyle.Foreground(lightGray)
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
  Aliases: []string{"l"},
	Short: "List all projects that would be targetted",
	Run: func(cmd *cobra.Command, args []string) {
    wd, err := os.Getwd()
    if err != nil {
      panic(err)
    }

    projects := utils.GetAllProjects(wd, 0)
    rows := [][]string{}
    for _, project := range projects {
      rows = append(rows, []string{ project.Name })
    }
    t := table.New().
    Border(lipgloss.NormalBorder()).
    BorderStyle(lipgloss.NewStyle().Foreground(purple)).
    StyleFunc(func(row, col int) lipgloss.Style {
        switch {
        case row == table.HeaderRow:
            return headerStyle
        case row%2 == 0:
            return evenRowStyle
        default:
            return oddRowStyle
        }
    }).
    Headers("Targets").
    Rows(rows...)

    fmt.Println(t)
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// lsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// lsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
