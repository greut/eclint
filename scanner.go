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
		if data[i] == '\r' {
			i++
			if i < len(data) && data[i] == '\n' {
				i++
			}
			return i, data[0:i], nil
		} else if data[i] == '\n' {
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

func readLines(r io.Reader, fn lineFunc) []error {
	errs := make([]error, 0)
	sc := bufio.NewScanner(r)
	sc.Split(splitLines)

	i := 0
	for sc.Scan() {
		line := sc.Bytes()
		if err := fn(i, line); err != nil {
			errs = append(errs, err)
		}
		i++
	}

	return errs
}
