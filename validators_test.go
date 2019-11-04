package main

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unicode/utf16"

	"github.com/editorconfig/editorconfig-core-go/v2"
	tlogr "github.com/go-logr/logr/testing"
)

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
				t.Errorf("No errors where expected, got %s", err)
			}
		})
	}
}

func TestEndOfLineFailures(t *testing.T) {
	tests := []struct {
		Name      string
		EndOfLine string
		Line      []byte
	}{
		{
			Name:      "cr instead of crlf",
			EndOfLine: "crlf",
			Line:      []byte("\r"),
		}, {
			Name:      "lf instead of crlf",
			EndOfLine: "crlf",
			Line:      []byte("\n"),
		}, {
			Name:      "cr instead of lf",
			EndOfLine: "lf",
			Line:      []byte("\r"),
		}, {
			Name:      "crlf instead of lf",
			EndOfLine: "lf",
			Line:      []byte("\r\n"),
		}, {
			Name:      "crlf instead of cr",
			EndOfLine: "cr",
			Line:      []byte("\r\n"),
		}, {
			Name:      "lf instead of cr",
			EndOfLine: "cr",
			Line:      []byte("\n"),
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			err := endOfLine(tc.EndOfLine, tc.Line)
			if err == nil {
				t.Error("An error was expected")
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
				t.Errorf("No errors where expected, got %s", err)
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
				t.Error("An error was expected")
			}
		})
	}
}

func utf16le(s string) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, []uint16{0xfeff})
	binary.Write(buf, binary.LittleEndian, utf16.Encode([]rune(s)))
	return buf.Bytes()
}

func utf16be(s string) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, []uint16{0xfeff})
	binary.Write(buf, binary.BigEndian, utf16.Encode([]rune(s)))
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
			//t.Parallel()

			def := &editorconfig.Definition{
				Charset: tc.Charset,
			}

			r := bytes.NewReader(tc.File)
			if err := validate(r, l, def); err != nil {
				t.Errorf("No errors where expected, got %s", err)
			}
		})
	}
}

func TestInsertFinalNewline(t *testing.T) {
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
			if err := validate(r, l, def); err != nil {
				t.Errorf("No errors where expected, got %s", err)
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

			if err := validate(r, l, def); err == nil {
				t.Error("An error was expected")
			}
		})
	}
}
