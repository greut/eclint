package eclint

import (
	"bufio"
	"io"
)

// LineFunc is the callback for a line.
//
// It returns the line number starting from zero.
type LineFunc func(int, []byte) error

// SplitLines works like bufio.ScanLines while keeping the line endings.
func SplitLines(data []byte, atEOF bool) (int, []byte, error) {
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

// ReadLines consumes the reader and emit each line via the LineFunc
//
// Line numbering starts at 0. Scanner is pretty smart an will reuse
// its memory structure. This is somehing we explicitly avoid by copying
// the content to a new slice.
func ReadLines(r io.Reader, fn LineFunc) []error {
	errs := make([]error, 0)
	sc := bufio.NewScanner(r)
	sc.Split(SplitLines)

	i := 0
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
