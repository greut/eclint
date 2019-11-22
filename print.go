package main

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"

	"github.com/logrusorgru/aurora"
	"github.com/mattn/go-colorable"
)

// lintAndPrint is the rich output of the program.
func lintAndPrint(opt option, filename string) int {
	c := 0
	d := 0

	stdout := opt.stdout
	if runtime.GOOS == "windows" {
		stdout = colorable.NewColorableStdout()
	}

	au := aurora.NewAurora(opt.isTerminal && !opt.noColors)
	errs := lint(filename, opt.log)
	for _, err := range errs {
		if err != nil {
			if d == 0 && !opt.summary {
				fmt.Fprintf(stdout, "%s:\n", au.Magenta(filename))
			}

			if ve, ok := err.(validationError); ok {
				opt.log.V(4).Info("lint error", "error", ve)
				if !opt.summary {
					vi := au.Green(strconv.Itoa(ve.index))
					vp := au.Green(strconv.Itoa(ve.position))
					fmt.Fprintf(stdout, "%s:%s: %s\n", vi, vp, ve.error)
					l, err := errorAt(au, ve.line, ve.position-1)
					if err != nil {
						opt.log.Error(err, "line formating failure", "error", ve)
						continue
					}
					fmt.Fprintln(stdout, l)
				}
			} else {
				opt.log.V(4).Info("lint error", "filename", filename, "error", err)
				fmt.Fprintln(stdout, err)
			}

			if d >= opt.showErrorQuantity && len(errs) > d {
				fmt.Fprintln(
					stdout,
					fmt.Sprintf(" ... skipping at most %s errors", au.BrightRed(strconv.Itoa(len(errs)-d))),
				)
				break
			}

			d++
			c++
		}
	}
	if d > 0 {
		if !opt.summary {
			fmt.Fprintln(stdout, "")
		} else {
			fmt.Fprintf(stdout, "%s: %d errors\n", au.Magenta(filename), d)
		}
	}
	return c
}

// errorAt highlights the validationError position within the line.
func errorAt(au aurora.Aurora, line []byte, position int) (string, error) {
	b := bytes.NewBuffer(make([]byte, len(line)))

	if position > len(line) {
		position = len(line)
	}

	for i := 0; i < position; i++ {
		if line[i] != cr && line[i] != lf {
			if err := b.WriteByte(line[i]); err != nil {
				return "", err
			}

			// skip 0x10xxxxxx that are UTF-8 continuation markers
			if (line[i] >> 6) == 0b10 {
				position++
			}
		}
	}

	// XXX this will break every non latin1 line.
	s := " "
	if position < len(line)-1 {
		s = string(line[position : position+1])
	}
	if _, err := b.WriteString(au.White(s).BgRed().String()); err != nil {
		return "", err
	}

	for i := position + 1; i < len(line); i++ {
		if line[i] != cr && line[i] != lf {
			if err := b.WriteByte(line[i]); err != nil {
				return "", err
			}

			if (line[i] >> 6) == 0b10 {
				position++
			}
		}
	}

	return b.String(), nil
}
