package eclint

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
	"unicode/utf16"

	"github.com/editorconfig/editorconfig-core-go/v2"
	tlogr "github.com/go-logr/logr/testing"
)

func utf16le(s string) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, []uint16{0xfeff})        // nolint: errcheck
	binary.Write(buf, binary.LittleEndian, utf16.Encode([]rune(s))) // nolint: errcheck

	return buf.Bytes()
}

func utf16be(s string) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, []uint16{0xfeff})        // nolint: errcheck
	binary.Write(buf, binary.BigEndian, utf16.Encode([]rune(s))) // nolint: errcheck

	return buf.Bytes()
}

func TestCharset(t *testing.T) {
	tests := []struct {
		Name    string
		Charset string
		File    []byte
	}{
		{
			Name:    "utf-8",
			Charset: "utf-8",
			File:    []byte("Hello world."),
		}, {
			Name:    "utf-8 bom",
			Charset: "utf-8 bom",
			File:    []byte{0xef, 0xbb, 0xbf, 'h', 'e', 'l', 'l', 'o', '.'},
		}, {
			Name:    "latin1",
			Charset: "latin1",
			File:    []byte("Hello world."),
		}, {
			Name:    "utf-16le",
			Charset: "utf-16le",
			File:    utf16le("Hello world."),
		}, {
			Name:    "utf-16be",
			Charset: "utf-16be",
			File:    utf16be("Hello world."),
		}, {
			Name:    "utf-32le",
			Charset: "utf-32le",
			File:    []byte{0xff, 0xfe, 0, 0, 'h', 0, 0, 0},
		}, {
			Name:    "utf-32be",
			Charset: "utf-32be",
			File:    []byte{0, 0, 0xfe, 0xff, 0, 0, 0, 'h'},
		},
	}

	l := tlogr.TestLogger{}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			def := &editorconfig.Definition{
				Charset: tc.Charset,
			}

			r := bytes.NewReader(tc.File)
			for _, err := range validate(r, "utf-8", l, def) {
				if err != nil {
					t.Errorf("no errors were expected, got %s", err)
				}
			}
		})
	}
}

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

			def := &editorconfig.Definition{
				InsertFinalNewline: &tc.InsertFinalNewline,
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
			def := &editorconfig.Definition{
				InsertFinalNewline: &insertFinalNewline,
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

			r := bytes.NewReader(tc.File)
			for _, err := range validate(r, "utf-8", l, def) {
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

	l := tlogr.TestLogger{}

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

			r := bytes.NewReader(tc.File)
			errs := validate(r, "utf-8", l, def)
			if len(errs) == 0 {
				t.Fatal("one error was expected, got none")
			}
			if errs[0] == nil {
				t.Errorf("no errors where expected, got %s", errs[0])
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

func TestOverridingUsingPrefix(t *testing.T) {
	def := &editorconfig.Definition{
		Charset:     "utf-8 bom",
		IndentStyle: "tab",
		IndentSize:  "3",
		TabWidth:    3,
	}
	raw := make(map[string]string)
	raw["@_charset"] = "unset"
	raw["@_indent_style"] = "space"
	raw["@_indent_size"] = "4"
	raw["@_tab_width"] = "4"
	def.Raw = raw

	err := overrideUsingPrefix(def, "@_")
	if err != nil {
		t.Fatal(err)
	}

	if def.Charset != "unset" {
		t.Errorf("charset not changed, got %q", def.Charset)
	}

	if def.IndentStyle != "space" {
		t.Errorf("indent_style not changed, got %q", def.IndentStyle)
	}

	if def.IndentSize != "4" {
		t.Errorf("indent_size not changed, got %q", def.IndentSize)
	}

	if def.TabWidth != 4 {
		t.Errorf("tab_width not changed, got %d", def.TabWidth)
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
