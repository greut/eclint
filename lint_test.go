package eclint

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/editorconfig/editorconfig-core-go/v2"
	tlogr "github.com/go-logr/logr/testing"
)

func TestInsertFinalNewline(t *testing.T) { // nolint:funlen
	tests := []struct {
		Name               string
		InsertFinalNewline bool
		File               []byte
	}{
		{
			Name:               "has final newline",
			InsertFinalNewline: true,
			File: []byte(`A file
			with a final newline.
`),
		}, {
			Name:               "has newline",
			InsertFinalNewline: false,
			File: []byte(`A file
without a final newline.`),
		}, {
			Name:               "empty file",
			InsertFinalNewline: true,
			File:               []byte(""),
		},
	}

	l := tlogr.TestLogger{}

	for _, tc := range tests {
		tc := tc

		// Test the nominal case
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			def, err := newDefinition(&editorconfig.Definition{
				InsertFinalNewline: &tc.InsertFinalNewline,
			})
			if err != nil {
				t.Fatal(err)
			}

			r := bytes.NewReader(tc.File)
			for _, err := range validate(r, "utf-8", l, def) {
				if err != nil {
					t.Errorf("no errors where expected, got %s", err)
				}
			}
		})

		// Test the inverse
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			insertFinalNewline := !tc.InsertFinalNewline
			def, err := newDefinition(&editorconfig.Definition{
				InsertFinalNewline: &insertFinalNewline,
			})
			if err != nil {
				t.Fatal(err)
			}

			r := bytes.NewReader(tc.File)

			for _, err := range validate(r, "utf-8", l, def) {
				if err == nil {
					t.Error("an error was expected")
				}
			}
		})
	}
}

func TestLintSimple(t *testing.T) {
	l := tlogr.TestLogger{}

	for _, err := range Lint("testdata/simple/simple.txt", l) {
		if err != nil {
			t.Errorf("no errors where expected, got %s", err)
		}
	}
}

func TestLintMissing(t *testing.T) {
	l := tlogr.TestLogger{}

	errs := Lint("testdata/missing/file", l)
	if len(errs) == 0 {
		t.Error("an error was expected, got none")
	}

	for _, err := range errs {
		if err == nil {
			t.Error("an error was expected")
		}
	}
}

func TestLintInvalid(t *testing.T) {
	l := tlogr.TestLogger{}

	errs := Lint("testdata/invalid/.editorconfig", l)
	if len(errs) == 0 {
		t.Error("an error was expected, got none")
	}

	for _, err := range errs {
		if err == nil {
			t.Error("an error was expected")
		}
	}
}

func TestBlockComment(t *testing.T) {
	tests := []struct {
		Name              string
		BlockCommentStart string
		BlockComment      string
		BlockCommentEnd   string
		File              []byte
	}{
		{
			Name:              "Java",
			BlockCommentStart: "/*",
			BlockComment:      "*",
			BlockCommentEnd:   "*/",
			File: []byte(`
	/**
	 *
	 */
	public class ... {}
`),
		},
	}

	l := tlogr.TestLogger{}

	for _, tc := range tests {
		tc := tc

		// Test the nominal case
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			def := &editorconfig.Definition{
				EndOfLine:   "lf",
				Charset:     "utf-8",
				IndentStyle: "tab",
			}
			def.Raw = make(map[string]string)
			def.Raw["block_comment_start"] = tc.BlockCommentStart
			def.Raw["block_comment"] = tc.BlockComment
			def.Raw["block_comment_end"] = tc.BlockCommentEnd
			d, err := newDefinition(def)
			if err != nil {
				t.Fatal(err)
			}

			r := bytes.NewReader(tc.File)
			for _, err := range validate(r, "utf-8", l, d) {
				if err != nil {
					t.Errorf("no errors where expected, got %s", err)
				}
			}
		})
	}
}

func TestBlockCommentFailure(t *testing.T) {
	tests := []struct {
		Name              string
		BlockCommentStart string
		BlockComment      string
		BlockCommentEnd   string
		File              []byte
	}{
		{
			Name:              "Java no block_comment_end",
			BlockCommentStart: "/*",
			BlockComment:      "*",
			BlockCommentEnd:   "",
			File:              []byte(`Hello!`),
		},
	}

	for _, tc := range tests {
		tc := tc

		// Test the nominal case
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			def := &editorconfig.Definition{
				IndentStyle: "tab",
			}
			def.Raw = make(map[string]string)
			def.Raw["block_comment_start"] = tc.BlockCommentStart
			def.Raw["block_comment"] = tc.BlockComment
			def.Raw["block_comment_end"] = tc.BlockCommentEnd

			_, err := newDefinition(def)
			if err == nil {
				t.Fatal("one error was expected, got none")
			}
		})
	}
}

func TestBlockCommentValidSpec(t *testing.T) {
	l := tlogr.TestLogger{}

	for _, f := range []string{"a", "b"} {
		for _, err := range Lint(fmt.Sprintf("./testdata/block_comments/%s", f), l) {
			if err != nil {
				t.Fatalf("no errors where expected, got %s", err)
			}
		}
	}
}

func TestBlockCommentInvalidSpec(t *testing.T) {
	l := tlogr.TestLogger{}

	for _, f := range []string{"c"} {
		errs := Lint(fmt.Sprintf("./testdata/block_comments/%s", f), l)
		if len(errs) == 0 {
			t.Errorf("one error was expected, got none")
		}
	}
}

func TestLintCharset(t *testing.T) {
	l := tlogr.TestLogger{}

	for _, f := range []string{"latin1", "utf8"} {
		for _, err := range Lint(fmt.Sprintf("./testdata/charset/%s.txt", f), l) {
			if err != nil {
				t.Errorf("no errors where expected, got %s", err)
			}
		}
	}
}

func TestLintImages(t *testing.T) {
	l := tlogr.TestLogger{}

	for _, f := range []string{"edcon_tool.png", "edcon_tool.pdf", "hello.txt.gz"} {
		for _, err := range Lint(fmt.Sprintf("./testdata/images/%s", f), l) {
			if err != nil {
				t.Fatalf("no errors where expected, got %s", err)
			}
		}
	}
}

func TestMaxLineLengthValidSpec(t *testing.T) {
	l := tlogr.TestLogger{}

	for _, f := range []string{"a", "b"} {
		for _, err := range Lint(fmt.Sprintf("./testdata/max_line_length/%s", f), l) {
			if err != nil {
				t.Fatalf("no errors where expected, got %s", err)
			}
		}
	}
}

func TestMaxLineLengthInvalidSpec(t *testing.T) {
	l := tlogr.TestLogger{}

	for _, f := range []string{"c"} {
		errs := Lint(fmt.Sprintf("./testdata/max_line_length/%s", f), l)
		if len(errs) == 0 {
			t.Errorf("one error was expected, got none")
		}
	}
}
