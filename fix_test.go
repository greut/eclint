package eclint

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/google/go-cmp/cmp"
)

func TestFixEndOfLine(t *testing.T) { //nolint:gocognit
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

	ctx := context.TODO()

	for _, tc := range tests {
		tc := tc

		file := bytes.Join(tc.Lines, []byte("\n"))
		fileSize := int64(len(file))

		// Test the nominal case
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			def, err := newDefinition(&editorconfig.Definition{
				EndOfLine: editorconfig.EndOfLineLf,
			})
			if err != nil {
				t.Fatal(err)
			}

			r := bytes.NewReader(file)
			out, fixed, err := fix(ctx, r, fileSize, "utf-8", def)
			if err != nil {
				t.Fatalf("no errors where expected, got %s", err)
			}

			if fixed {
				t.Errorf("file should not have been fixed")
			}

			result, err := io.ReadAll(out)
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
				EndOfLine: editorconfig.EndOfLineCrLf,
			})
			if err != nil {
				t.Fatal(err)
			}

			r := bytes.NewReader(file)
			out, fixed, err := fix(ctx, r, fileSize, "utf-8", def)
			if err != nil {
				t.Fatalf("no errors where expected, got %s", err)
			}

			if !fixed {
				t.Errorf("file should have been fixed")
			}

			result, err := io.ReadAll(out)
			if err != nil {
				t.Fatalf("cannot read result %s", err)
			}

			if cmp.Equal(file, result) {
				t.Errorf("no differences, the file was not fixed")
			}
		})
	}
}

func TestFixIndentStyle(t *testing.T) {
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

	ctx := context.TODO()

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
			out, _, err := fix(ctx, r, fileSize, "utf-8", def)
			if err != nil {
				t.Fatalf("no errors where expected, got %s", err)
			}

			result, err := io.ReadAll(out)
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
				m, _ := fixTrailingWhitespace(l)

				err := checkTrimTrailingWhitespace(m)
				if err != nil {
					t.Errorf("no errors were expected. %s", err)
				}
			}
		})
	}
}

func TestFixInsertFinalNewline(t *testing.T) {
	eolVariants := [][]byte{
		{cr},
		{lf},
		{cr, lf},
	}

	insertFinalNewlineVariants := []bool{true, false}
	newlinesAtEOLVariants := []int{0, 1, 3}

	type Test struct {
		InsertFinalNewline bool
		File               []byte
		EolVariant         []byte
		NewlinesAtEOL      int
	}

	tests := make([]Test, 0, 54)

	// single line tests
	singleLineFile := []byte(`A single line file.`)

	for _, eolVariant := range eolVariants {
		for _, insertFinalNewlineVariant := range insertFinalNewlineVariants {
			for newlinesAtEOL := range newlinesAtEOLVariants {
				file := singleLineFile
				for i := 0; i < newlinesAtEOL; i++ {
					file = append(file, eolVariant...)
				}

				tests = append(tests,
					Test{
						InsertFinalNewline: insertFinalNewlineVariant,
						File:               file,
						EolVariant:         eolVariant,
						NewlinesAtEOL:      newlinesAtEOL,
					},
				)
			}
		}
	}

	// multiline tests
	multilineComponents := [][]byte{[]byte(`A`), []byte(`multiline`), []byte(`file.`)}

	for _, eolVariant := range eolVariants {
		multilineFile := bytes.Join(multilineComponents, eolVariant)

		for _, insertFinalNewlineVariant := range insertFinalNewlineVariants {
			for newlinesAtEOL := range newlinesAtEOLVariants {
				file := multilineFile
				for i := 0; i < newlinesAtEOL; i++ {
					file = append(file, eolVariant...)
				}

				tests = append(tests,
					Test{
						InsertFinalNewline: insertFinalNewlineVariant,
						File:               file,
						EolVariant:         eolVariant,
						NewlinesAtEOL:      newlinesAtEOL,
					},
				)
			}
		}
	}

	// empty file tests
	emptyFile := []byte("")

	for _, eolVariant := range eolVariants {
		for _, insertFinalNewlineVariant := range insertFinalNewlineVariants {
			tests = append(tests,
				Test{
					InsertFinalNewline: insertFinalNewlineVariant,
					File:               emptyFile,
					EolVariant:         eolVariant,
				},
			)
		}
	}

	for _, tc := range tests {
		tc := tc

		t.Run("TestFixInsertFinalNewline", func(t *testing.T) {
			t.Parallel()

			buf := bytes.Buffer{}
			buf.Write(tc.File)
			before := buf.Bytes()
			fixInsertFinalNewline(&buf, tc.InsertFinalNewline, tc.EolVariant)
			after := buf.Bytes()
			err := checkInsertFinalNewline(buf.Bytes(), tc.InsertFinalNewline)
			if err != nil {
				t.Logf("before: %q", string(before))
				t.Logf("after: %q", string(after))
				t.Errorf("encountered error %s with test configuration %+v", err, tc)
			}
		})
	}
}
