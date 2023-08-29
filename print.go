package eclint

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/logrusorgru/aurora"
)

// PrintErrors is the rich output of the program.
func PrintErrors(ctx context.Context, opt *Option, filename string, errs []error) error {
	counter := 0

	log := logr.FromContextOrDiscard(ctx)
	stdout := opt.Stdout

	au := aurora.NewAurora(opt.IsTerminal && !opt.NoColors)

	for _, err := range errs {
		if err != nil { //nolint:nestif
			if counter == 0 && !opt.Summary {
				fmt.Fprintf(stdout, "%s:\n", au.Magenta(filename).Bold())
			}

			var ve ValidationError
			if ok := errors.As(err, &ve); ok {
				log.V(4).Info("lint error", "error", ve)

				if !opt.Summary {
					vi := au.Green(strconv.Itoa(ve.Index + 1)).Bold()
					vp := au.Green(strconv.Itoa(ve.Position + 1)).Bold()
					fmt.Fprintf(stdout, "%s:%s: %s\n", vi, vp, ve.Message)

					l, err := errorAt(au, ve.Line, ve.Position)
					if err != nil {
						log.Error(err, "line formatting failure", "error", ve)

						return err
					}

					fmt.Fprintln(stdout, l)
				}
			} else {
				log.V(2).Info("lint error", "filename", filename, "error", err.Error())
				fmt.Fprintln(stdout, err)
			}

			counter++

			if opt.ShowErrorQuantity > 0 && counter >= opt.ShowErrorQuantity && len(errs) > counter {
				fmt.Fprintf(
					stdout,
					" ... skipping at most %s errors\n",
					au.BrightRed(strconv.Itoa(len(errs)-counter)),
				)

				break
			}
		}
	}

	if counter > 0 {
		if !opt.Summary {
			fmt.Fprintln(stdout, "")
		} else {
			fmt.Fprintf(stdout, "%s: %d errors\n", au.Magenta(filename), counter)
		}
	}

	return nil
}

// errorAt highlights the ValidationError position within the line.
func errorAt(au aurora.Aurora, line []byte, position int) (string, error) {
	b := bytes.NewBuffer(make([]byte, len(line)))

	if position > len(line)-1 {
		position = len(line) - 1
	}

	for i := 0; i < position; i++ {
		if line[i] != cr && line[i] != lf {
			if err := b.WriteByte(line[i]); err != nil {
				return "", fmt.Errorf("error writing byte: %w", err)
			}
		}
	}

	// Rewind the 0x10xxxxxx that are UTF-8 continuation markers
	for i := position; i > 0; i-- {
		if (line[i] >> 6) != 0b10 {
			break
		}
		position--
	}

	// XXX this will break every non latin1 line.
	s := " "
	if position < len(line)-1 {
		s = string(line[position : position+1])
	}

	if _, err := b.WriteString(au.White(s).BgRed().String()); err != nil {
		return "", fmt.Errorf("error writing string: %w", err)
	}

	for i := position + 1; i < len(line); i++ {
		if line[i] != cr && line[i] != lf {
			if err := b.WriteByte(line[i]); err != nil {
				return "", fmt.Errorf("error writing byte: %w", err)
			}

			if (line[i] >> 6) == 0b10 {
				i++
			}
		}
	}

	return b.String(), nil
}
