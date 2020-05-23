package eclint

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/editorconfig/editorconfig-core-go/v2"
	tlogr "github.com/go-logr/logr/testing"
	"github.com/google/go-cmp/cmp"
)

func TestFixEndOfLine(t *testing.T) { // nolint:funlen
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
				t.Fatalf("no errors where expected, got %s", err)
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

func TestFixIndentStyle(t *testing.T) { // nolint:funlen
	tests := []struct {
		Name        string
		IndentSize  string
		IndentStyle string
		File        []byte
	}{
		{
			Name:        "space to tab",
			IndentStyle: "tab",
			IndentSize:  "2",
			File:        []byte("\t\t  \tA line\n"),
		},
		{
			Name:        "tab to space",
			IndentStyle: "space",
			IndentSize:  "2",
			File:        []byte("\t\t  \tA line\n"),
		},
	}

	l := tlogr.TestLogger{}

	for _, tc := range tests {
		tc := tc

		fileSize := int64(len(tc.File))

		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			def, err := newDefinition(&editorconfig.Definition{
				EndOfLine:   "lf",
				IndentStyle: tc.IndentStyle,
				IndentSize:  tc.IndentSize,
			})
			if err != nil {
				t.Fatal(err)
			}

			if err := indentStyle(tc.IndentStyle, def.IndentSize, tc.File); err == nil {
				t.Errorf("the initial file should fail")
			}

			r := bytes.NewReader(tc.File)
			out, err := fix(r, fileSize, "utf-8", l, def)
			if err != nil {
				t.Fatalf("no errors where expected, got %s", err)
			}

			result, err := ioutil.ReadAll(out)
			if err != nil {
				t.Fatalf("cannot read result %s", err)
			}

			if cmp.Equal(tc.File, result) {
				t.Errorf("no changes!?")
			}

			if err := indentStyle(tc.IndentStyle, def.IndentSize, result); err != nil {
				t.Errorf("no errors were expected, got %s", err)
			}
		})
	}
}

func TestFixTrimTrailingWhitespace(t *testing.T) {
	tests := []struct {
		Name  string
		Lines [][]byte
	}{
		{
			Name: "space",
			Lines: [][]byte{
				[]byte("A file"),
				[]byte(" with spaces "),
				[]byte(" at the end  "),
				[]byte(" "),
			},
		},
		{
			Name: "tabs",
			Lines: [][]byte{
				[]byte("A file"),
				[]byte(" with tabs\t"),
				[]byte(" at the end\t\t"),
				[]byte("\t"),
			},
		},
		{
			Name: "tabs and spaces",
			Lines: [][]byte{
				[]byte("A file"),
				[]byte(" with tabs\t\t "),
				[]byte(" and spaces\t \t"),
				[]byte(" at the end \t"),
			},
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			for _, l := range tc.Lines {
				m := fixTrailingWhitespace(l)

				err := checkTrimTrailingWhitespace(m)
				if err != nil {
					t.Errorf("no errors were expected. %s", err)
				}
			}
		})
	}
}
