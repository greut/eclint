package eclint_test

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"testing"
	"unicode/utf16"

	tlogr "github.com/go-logr/logr/testing"
	"gitlab.com/greut/eclint"
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

func TestProbeCharsetOfBinary(t *testing.T) {
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

			r := bytes.NewReader(tc.File)
			br := bufio.NewReader(r)

			charset, ok, err := eclint.ProbeCharsetOrBinary(br, tc.Charset, l)
			if err != nil {
				t.Errorf("no errors were expected, got %s", err)
			}

			if ok {
				t.Errorf("no binary should have been detected")
			}

			if charset != tc.Charset {
				t.Errorf("bad charset. expected %s, got %s", tc.Charset, charset)
			}
		})
	}
}
