package eclint

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/editorconfig/editorconfig-core-go/v2"
	tlogr "github.com/go-logr/logr/testing"
	"github.com/google/go-cmp/cmp"
)

func TestFixEndOfline(t *testing.T) { // nolint:funlen
	tests := []struct {
		Name  string
		Lines [][]byte
	}{
		{
			Name: "a file with many lines",
			Lines: [][]byte{
				[]byte("A file"),
				[]byte("With many lines"),
			},
		},
		{
			Name: "a file with many lines and a final newline",
			Lines: [][]byte{
				[]byte("A file"),
				[]byte("With many lines"),
				[]byte("and a final newline."),
				[]byte(""),
			},
		},
	}

	l := tlogr.TestLogger{}

	for _, tc := range tests {
		tc := tc

		file := bytes.Join(tc.Lines, []byte("\n"))
		fileSize := int64(len(file))

		// Test the nominal case
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			def, err := newDefinition(&editorconfig.Definition{
				EndOfLine: "lf",
			})
			if err != nil {
				t.Fatal(err)
			}

			r := bytes.NewReader(file)
			out, err := fix(r, fileSize, "utf-8", l, def)
			if err != nil {
				t.Fatalf("no errors where expected, got %s", err)
			}

			result, err := ioutil.ReadAll(out)
			if err != nil {
				t.Fatalf("cannot read result %s", err)
			}

			if !cmp.Equal(file, result) {
				t.Errorf("diff %s", cmp.Diff(file, result))
			}
		})

		// Test the inverse
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			def, err := newDefinition(&editorconfig.Definition{
				EndOfLine: "crlf",
			})
			if err != nil {
				t.Fatal(err)
			}

			r := bytes.NewReader(file)
			out, err := fix(r, fileSize, "utf-8", l, def)
			if err != nil {
				t.Errorf("no errors where expected, got %s", err)
			}

			result, err := ioutil.ReadAll(out)
			if err != nil {
				t.Fatalf("cannot read result %s", err)
			}

			if cmp.Equal(file, result) {
				t.Errorf("no differences, the file was not fixed")
			}
		})
	}
}
