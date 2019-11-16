package main

import (
	"bytes"
	"testing"

	"github.com/editorconfig/editorconfig-core-go/v2"
	tlogr "github.com/go-logr/logr/testing"
)

func TestCharsetUsingBOMFailure(t *testing.T) {
	tests := []struct {
		Charset string
	}{
		{
			Charset: "utf-8 bom",
		}, {
			Charset: "utf-16le",
		}, {
			Charset: "utf-16be",
		}, {
			Charset: "utf-32le",
		}, {
			Charset: "utf-32be",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Charset, func(t *testing.T) {
			t.Parallel()
			ok, err := charsetUsingBOM(tc.Charset, []byte{})
			if ok || err == nil {
				t.Error("an error was expected")
			}
		})
	}
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
			//t.Parallel()

			def := &editorconfig.Definition{
				Charset: tc.Charset,
			}

			r := bytes.NewReader(tc.File)
			for _, err := range validate(r, l, def) {
				if err != nil {
					t.Errorf("no errors were expected, got %s", err)
				}
			}
		})
	}
}

func TestEndOfLine(t *testing.T) {
	tests := []struct {
		Name      string
		EndOfLine string
		Line      []byte
	}{
		{
			Name:      "crlf",
			EndOfLine: "crlf",
			Line:      []byte("\r\n"),
		}, {
			Name:      "lf",
			EndOfLine: "lf",
			Line:      []byte("\n"),
		}, {
			Name:      "cr",
			EndOfLine: "cr",
			Line:      []byte("\r"),
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			err := endOfLine(tc.EndOfLine, tc.Line)
			if err != nil {
				t.Errorf("no errors were expected, got %s", err)
			}
		})
	}
}

func TestEndOfLineFailures(t *testing.T) {
	tests := []struct {
		Name      string
		EndOfLine string
		Line      []byte
		Position  int
	}{
		{
			Name:      "cr instead of crlf",
			EndOfLine: "crlf",
			Line:      []byte("\r"),
			Position:  1,
		}, {
			Name:      "lf instead of crlf",
			EndOfLine: "crlf",
			Line:      []byte("[*]\n"),
			Position:  4,
		}, {
			Name:      "cr instead of lf",
			EndOfLine: "lf",
			Line:      []byte("\r"),
			Position:  1,
		}, {
			Name:      "crlf instead of lf",
			EndOfLine: "lf",
			Line:      []byte("\r\n"),
			Position:  2,
		}, {
			Name:      "crlf instead of cr",
			EndOfLine: "cr",
			Line:      []byte("\r\n"),
			Position:  2,
		}, {
			Name:      "lf instead of cr",
			EndOfLine: "cr",
			Line:      []byte("hello\n"),
			Position:  6,
		}, {
			Name:      "unknown eol",
			EndOfLine: "lfcr",
			Line:      []byte("\n\r"),
			Position:  -1,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			err := endOfLine(tc.EndOfLine, tc.Line)
			ve, ok := err.(validationError)
			if tc.Position >= 0 {
				if !ok {
					t.Errorf("a validationError was expected, got %v", err)
				}
				if tc.Position != ve.position {
					t.Errorf("position mismatch %d, got %d", tc.Position, ve.position)
				}
			} else if err == nil {
				t.Error("an error was expected")
			}
		})
	}
}

func TestTrimTrailingWhitespace(t *testing.T) {
	tests := []struct {
		Name string
		Line []byte
	}{
		{
			Name: "crlf",
			Line: []byte("\r\n"),
		}, {
			Name: "cr",
			Line: []byte("\r"),
		}, {
			Name: "lf",
			Line: []byte("\n"),
		}, {
			Name: "words",
			Line: []byte("hello world."),
		}, {
			Name: "noeol",
			Line: []byte(""),
		}, {
			Name: "nbsp",
			Line: []byte{0xA0},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			err := trimTrailingWhitespace(tc.Line)
			if err != nil {
				t.Errorf("no errors were expected, got %s", err)
			}
		})
	}
}

func TestTrimTrailingWhitespaceFailure(t *testing.T) {
	tests := []struct {
		Name string
		Line []byte
	}{
		{
			Name: "space",
			Line: []byte(" \r\n"),
		}, {
			Name: "tab",
			Line: []byte("\t"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			err := trimTrailingWhitespace(tc.Line)
			if err == nil {
				t.Error("an error was expected")
			}
		})
	}
}

func TestIndentStyle(t *testing.T) {
	tests := []struct {
		Name        string
		IndentSize  int
		IndentStyle string
		Line        []byte
	}{
		{
			Name:        "empty line of tab",
			IndentSize:  1,
			IndentStyle: "tab",
			Line:        []byte("\t\t\t\t."),
		}, {
			Name:        "empty line of spaces",
			IndentSize:  1,
			IndentStyle: "space",
			Line:        []byte("      ."),
		}, {
			Name:        "three spaces",
			IndentSize:  3,
			IndentStyle: "space",
			Line:        []byte("            ."),
		}, {
			Name:        "three tabs",
			IndentSize:  3,
			IndentStyle: "tab",
			Line:        []byte("\t\t\t\t."),
		}, {
			Name:        "unset",
			IndentSize:  5,
			IndentStyle: "unset",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			err := indentStyle(tc.IndentStyle, tc.IndentSize, tc.Line)
			if err != nil {
				t.Errorf("no errors were expected, got %s", err)
			}
		})
	}
}

func TestIndentStyleFailure(t *testing.T) {
	tests := []struct {
		Name        string
		IndentSize  int
		IndentStyle string
		Line        []byte
	}{
		{
			Name:        "mix of tab and spaces",
			IndentSize:  2,
			IndentStyle: "space",
			Line:        []byte("  \t."),
		}, {
			Name:        "mix of tabs and space spaces",
			IndentSize:  2,
			IndentStyle: "tab",
			Line:        []byte("\t \t."),
		}, {
			Name:        "three spaces +1",
			IndentSize:  3,
			IndentStyle: "space",
			Line:        []byte("    ."),
		}, {
			Name:        "three spaces -1",
			IndentSize:  3,
			IndentStyle: "space",
			Line:        []byte("  ."),
		}, {
			Name:        "invalid size",
			IndentSize:  -1,
			IndentStyle: "space",
			Line:        []byte("."),
		}, {
			Name:        "invalid style",
			IndentSize:  3,
			IndentStyle: "fubar",
			Line:        []byte("."),
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			err := indentStyle(tc.IndentStyle, tc.IndentSize, tc.Line)
			if err == nil {
				t.Error("an error was expected")
			}
		})
	}
}

func TestCheckBlockComment(t *testing.T) {
	tests := []struct {
		Name     string
		Position int
		Prefix   []byte
		Line     []byte
	}{
		{
			Name:     "Java",
			Position: 5,
			Prefix:   []byte{'*'},
			Line:     []byte("\t\t\t\t *\r\n"),
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			err := checkBlockComment(tc.Position, tc.Prefix, tc.Line)
			if err != nil {
				t.Errorf("no errors were expected, got %s", err)
			}
		})
	}
}

func TestMaxLineLength(t *testing.T) {
	tests := []struct {
		Name          string
		MaxLineLength int
		TabWidth      int
		Line          []byte
	}{
		{
			Name:          "no limits",
			MaxLineLength: 0,
			TabWidth:      0,
			Line:          []byte("\r\n"),
		}, {
			Name:          "some limit",
			MaxLineLength: 1,
			TabWidth:      0,
			Line:          []byte(".\r\n"),
		}, {
			Name:          "some limit",
			MaxLineLength: 10,
			TabWidth:      0,
			Line:          []byte("0123456789\n"),
		}, {
			Name:          "tabs",
			MaxLineLength: 5,
			TabWidth:      2,
			Line:          []byte("\t\t.\n"),
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			err := maxLineLength(tc.MaxLineLength, tc.TabWidth, tc.Line)
			if err != nil {
				t.Errorf("no errors were expected, got %s", err)
			}
		})
	}
}

func TestMaxLineLengthFailure(t *testing.T) {
	tests := []struct {
		Name          string
		MaxLineLength int
		TabWidth      int
		Line          []byte
	}{
		{
			Name:          "small limit",
			MaxLineLength: 1,
			TabWidth:      1,
			Line:          []byte("..\r\n"),
		}, {
			Name:          "small limit and tab",
			MaxLineLength: 2,
			TabWidth:      2,
			Line:          []byte("\t.\r\n"),
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			err := maxLineLength(tc.MaxLineLength, tc.TabWidth, tc.Line)
			if err == nil {
				t.Error("an error was expected")
			}
		})
	}
}
