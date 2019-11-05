package main

import (
	"bytes"
	"encoding/binary"
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
			//t.Parallel()

			def := &editorconfig.Definition{
				Charset: tc.Charset,
			}

			r := bytes.NewReader(tc.File)
			for _, err := range validate(r, l, def) {
				if err != nil {
					t.Errorf("no errors where expected, got %s", err)
				}
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
			for _, err := range validate(r, l, def) {
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

			for _, err := range validate(r, l, def) {
				if err == nil {
					t.Error("an error was expected")
				}
			}
		})
	}
}

func TestLintSimple(t *testing.T) {
	l := tlogr.TestLogger{}

	for _, err := range lint("testdata/simple/simple.txt", l) {
		if err != nil {
			t.Errorf("no errors where expected, got %s", err)
		}
	}
}

func TestLintMissing(t *testing.T) {
	l := tlogr.TestLogger{}

	for _, err := range lint("testdata/missing/file", l) {
		if err == nil {
			t.Error("an error was expected")
		}
		return
	}
	t.Error("an error was expected, got none")
}

func TestLintInvalid(t *testing.T) {
	l := tlogr.TestLogger{}

	for _, err := range lint("testdata/invalid/.editorconfig", l) {
		if err == nil {
			t.Error("an error was expected")
		}
		return
	}
	t.Error("an error was expected, got none")
}
