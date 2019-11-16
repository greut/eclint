package main

import (
	"bufio"
	"io"
)

// lineFunc is the callback for a line.
//
// It returns the line number starting from zero.
type lineFunc func(int, []byte) error

// splitLines works like bufio.ScanLines while keeping the line endings.
func splitLines(data []byte, atEOF bool) (int, []byte, error) {
	i := 0
	for i < len(data) {
		if data[i] == cr {
			i++
			if i < len(data) && data[i] == lf {
				i++
			}
			return i, data[0:i], nil
		} else if data[i] == lf {
			i++
			return i, data[0:i], nil
		}
		i++
	}

	if !atEOF {
		return 0, nil, nil
	}

	if atEOF && i != 0 {
		return 0, data, bufio.ErrFinalToken
	}

	return 0, nil, io.EOF
}

// readLines consumes the reader and emit each line via the lineFunc
//
// Line numbering starts at 1. Scanner is pretty smart an will reuse
// its memory structure. This is somehing we explicitly avoid by copying
// the content to a new slice.
func readLines(r io.Reader, fn lineFunc) []error {
	errs := make([]error, 0)
	sc := bufio.NewScanner(r)
	sc.Split(splitLines)

	i := 1
	for sc.Scan() {
		l := sc.Bytes()
		line := make([]byte, len(l))
		copy(line, l)
		if err := fn(i, line); err != nil {
			errs = append(errs, err)
		}
		i++
	}

	return errs
}
