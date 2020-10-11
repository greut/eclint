package eclint_test

import (
	"bytes"
	"fmt"
	"testing"

	"gitlab.com/greut/eclint"
)

func TestReadLines(t *testing.T) { // nolint: funlen
	tests := []struct {
		Name     string
		File     []byte
		LineFunc eclint.LineFunc
	}{
		{
			Name: "Empty file",
			File: []byte(""),
			LineFunc: func(i int, line []byte, isEOF bool) error {
				if i != 0 || len(line) > 0 {
					return fmt.Errorf("more than one line found (%d), or non epmty line %q", i, line)
				}

				return nil
			},
		}, {
			Name: "crlf",
			File: []byte("\r\n\r\n"),
			LineFunc: func(i int, line []byte, isEOF bool) error {
				if i > 1 || len(line) > 2 {
					return fmt.Errorf("more than two lines found (%d), or non empty line %q", i, line)
				}

				return nil
			},
		}, {
			Name: "cr",
			File: []byte("\r\r"),
			LineFunc: func(i int, line []byte, isEOF bool) error {
				if i > 1 || len(line) > 2 {
					return fmt.Errorf("more than two lines found (%d), or non empty line %q", i, line)
				}

				return nil
			},
		}, {
			Name: "lf",
			File: []byte("\n\n"),
			LineFunc: func(i int, line []byte, isEOF bool) error {
				if i > 1 || len(line) > 2 {
					return fmt.Errorf("more than two lines found (%d), or non empty line %q", i, line)
				}

				return nil
			},
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			r := bytes.NewReader(tc.File)
			errs := eclint.ReadLines(r, -1, tc.LineFunc)
			if len(errs) > 0 {
				t.Errorf("no errors were expected, got some. %s", errs[0])

				return
			}
		})
	}
}
