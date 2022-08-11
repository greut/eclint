package eclint_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"testing"
	"unicode/utf16"

	"gitlab.com/greut/eclint"
)

func utf16le(s string) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, []uint16{0xfeff})        //nolint:errcheck
	binary.Write(buf, binary.LittleEndian, utf16.Encode([]rune(s))) //nolint:errcheck

	return buf.Bytes()
}

func utf16be(s string) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, []uint16{0xfeff})        //nolint:errcheck
	binary.Write(buf, binary.BigEndian, utf16.Encode([]rune(s))) //nolint:errcheck

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
			File:    []byte{'h', 'i', ' ', 0xf0, 0x9f, 0x92, 0xa9, '!'},
		}, {
			Name:    "empty utf-8",
			Charset: "utf-8",
			File:    []byte(""),
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

	ctx := context.TODO()

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			r := bytes.NewReader(tc.File)
			br := bufio.NewReader(r)

			charset, ok, err := eclint.ProbeCharsetOrBinary(ctx, br, tc.Charset)
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

func TestProbeCharsetOfBinaryFailure(t *testing.T) {
	tests := []struct {
		Name    string
		Charset string
		File    []byte
	}{
		{
			Name:    "utf-8 vs latin1",
			Charset: "latin1",
			File:    []byte{'h', 'i', ' ', 0xf0, 0x9f, 0x92, 0xa9, '!'},
		},
	}

	ctx := context.TODO()

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			r := bytes.NewReader(tc.File)
			br := bufio.NewReader(r)

			charset, ok, err := eclint.ProbeCharsetOrBinary(ctx, br, tc.Charset)
			if err == nil {
				t.Errorf("an error was expected, got charset %s, %v", charset, ok)
			}
		})
	}
}

func TestProbeCharsetOfBinaryForBinary(t *testing.T) {
	tests := []struct {
		Name    string
		Charset string
		File    []byte
	}{
		{
			Name: "euro but reversed",
			File: []byte{0xac, 0x82, 0xe2},
		}, {
			Name: "euro but trucated",
			File: []byte{0xe2, 0x82},
		}, {
			Name: "poop but middle only",
			File: []byte{0x9f, 0x92, 0xa9},
		}, {
			Name: "poop emoji but trucated",
			File: []byte{0xf0, 0x9f, 0x92},
		},
	}

	ctx := context.TODO()

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			r := bytes.NewReader(tc.File)
			br := bufio.NewReader(r)

			charset, ok, err := eclint.ProbeCharsetOrBinary(ctx, br, "")
			if err != nil {
				t.Errorf("no errors were expected %s", err)
			}

			if !ok {
				t.Errorf("binary should have been detected got %s", charset)
			}
		})
	}
}
