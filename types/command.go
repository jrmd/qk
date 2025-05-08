package types

import (
	"bufio"
	"bytes"
	"context"
)

type Command struct {
	Script string
	Args   []string
	Status string
	Ctx    context.Context
	Cancel context.CancelFunc
	Output *bytes.Buffer
	Render func(*Command) string
	Reader *bufio.Scanner
}
