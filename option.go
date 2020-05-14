package eclint

import (
	"io"

	"github.com/go-logr/logr"
)

// Option contains the environment of the program.
//
// When ShowErrorQuantity is 0, it will show all the errors. Use ShowAllErrors false to disable this.
type Option struct {
	IsTerminal        bool
	NoColors          bool
	ShowAllErrors     bool
	Summary           bool
	FixAllErrors      bool
	ShowErrorQuantity int
	Exclude           string
	Log               logr.Logger
	Stdout            io.Writer
}
