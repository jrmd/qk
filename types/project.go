package types

import (
	"github.com/charmbracelet/bubbles/spinner"
)

type Project struct {
	Spinner spinner.Model
	Name    string
	Dir     string
	Scripts []*Command
}
