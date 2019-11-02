package main

import (
	"testing"
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
