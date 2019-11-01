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
	for i := 0; i < len(data); {
		if data[i] == '\r' {
			i++
			if data[i] == '\n' {
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
	return len(data) + 1, data, nil
}

func readLines(r io.Reader, fn lineFunc) error {
	sc := bufio.NewScanner(r)
	sc.Split(splitLines)

	i := 0
	for sc.Scan() {
		line := sc.Bytes()
		if err := fn(i, line); err != nil {
			return err
		}
		i++
	}

	return nil
}
