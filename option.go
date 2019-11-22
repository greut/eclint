package eclint

import (
	"io"

	"github.com/go-logr/logr"
)

// Option contains the environment of the program.
type Option struct {
	IsTerminal        bool
	NoColors          bool
	ShowAllErrors     bool
	Summary           bool
	ShowErrorQuantity int
	Exclude           string
	Log               logr.Logger
	Stdout            io.Writer
}
