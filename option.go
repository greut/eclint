package eclint

import (
	"io"
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
	Stdout            io.Writer
}
