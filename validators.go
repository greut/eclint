package eclint

import (
	"bytes"
	"errors"
	"fmt"
)

const (
	cr    = '\r'
	lf    = '\n'
	tab   = '\t'
	space = ' '
)

var (
	utf8Bom    = []byte{0xef, 0xbb, 0xbf} // nolint:gochecknoglobals
	utf16leBom = []byte{0xff, 0xfe}       // nolint:gochecknoglobals
	utf16beBom = []byte{0xfe, 0xff}       // nolint:gochecknoglobals
	utf32leBom = []byte{0xff, 0xfe, 0, 0} // nolint:gochecknoglobals
	utf32beBom = []byte{0, 0, 0xfe, 0xff} // nolint:gochecknoglobals
)

// ErrConfiguration represents an error in the editorconfig value.
var ErrConfiguration = errors.New("configuration error")

// ValidationError is a rich type containing information about the error.
type ValidationError struct {
	Message  string
	Filename string
	Line     []byte
	Index    int
	Position int
}

func (e ValidationError) String() string {
	return e.Error()
}

// Error builds the error string.
func (e ValidationError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.Filename, e.Index+1, e.Position+1, e.Message)
}

// endOfLines checks the line ending.
func endOfLine(eol string, data []byte) error {
	switch eol {
	case "lf":
		if !bytes.HasSuffix(data, []byte{lf}) || bytes.HasSuffix(data, []byte{cr, lf}) {
			return ValidationError{
				Message:  "line does not end with lf (`\\n`)",
				Position: len(data),
			}
		}
	case "crlf":
		if !bytes.HasSuffix(data, []byte{cr, lf}) {
			return ValidationError{
				Message:  "line does not end with crlf (`\\r\\n`)",
				Position: len(data),
			}
		}
	case "cr":
		if !bytes.HasSuffix(data, []byte{cr}) {
			return ValidationError{
				Message:  "line does not end with cr (`\\r`)",
				Position: len(data),
			}
		}
	default:
		return fmt.Errorf("%w: %q is an invalid value for eol, want cr, crlf, or lf", ErrConfiguration, eol)
	}

	return nil
}

// indentStyle checks that the line beginnings are either space or tabs.
func indentStyle(style string, size int, data []byte) error {
	var c byte

	var x byte

	switch style {
	case SpaceValue:
		c = space
		x = tab
	case TabValue:
		c = tab
		x = space
		size = 1
	case UnsetValue:
		return nil
	default:
		return fmt.Errorf("%w: %q is an invalid value of indent_style, want tab or space", ErrConfiguration, style)
	}

	if size == 0 {
		return nil
	}

	if size < 0 {
		return fmt.Errorf("%w: %d is an invalid value of indent_size, want a number or unset", ErrConfiguration, size)
	}

	for i := 0; i < len(data); i++ {
		if data[i] == c {
			continue
		}

		if data[i] == x {
			return ValidationError{
				Message:  fmt.Sprintf("indentation style mismatch expected %q (%s) got %q", c, style, x),
				Position: i,
			}
		}

		if data[i] == cr || data[i] == lf || (size > 0 && i%size == 0) {
			break
		}

		return ValidationError{
			Message:  fmt.Sprintf("indentation size doesn't match expected %d, got %d", size, i),
			Position: i,
		}
	}

	return nil
}

// checkInsertFinalNewline checks whenever the final line contains a newline or not.
func checkInsertFinalNewline(data []byte, insertFinalNewline bool) error {
	if len(data) == 0 {
		return nil
	}

	lastChar := data[len(data)-1]
	if lastChar != cr && lastChar != lf {
		if insertFinalNewline {
			return ValidationError{
				Message:  "the final newline is missing",
				Position: len(data),
			}
		}
	} else {
		if !insertFinalNewline {
			return ValidationError{
				Message:  "an extraneous final newline was found",
				Position: len(data),
			}
		}
	}

	return nil
}

// checkTrimTrailingWhitespace lints any spaces before the final newline.
func checkTrimTrailingWhitespace(data []byte) error {
	for i := len(data) - 1; i >= 0; i-- {
		if data[i] == cr || data[i] == lf {
			continue
		}

		if data[i] == space || data[i] == tab {
			return ValidationError{
				Message:  "line has some trailing whitespaces",
				Position: i,
			}
		}

		break
	}

	return nil
}

// isBlockCommentStart tells you when a block comment started on this line.
func isBlockCommentStart(start []byte, data []byte) bool {
	for i := 0; i < len(data); i++ {
		if data[i] == space || data[i] == tab {
			continue
		}

		return bytes.HasPrefix(data[i:], start)
	}

	return false
}

// checkBlockComment checks the line is a valid block comment.
func checkBlockComment(i int, prefix []byte, data []byte) error {
	for ; i < len(data); i++ {
		if data[i] == space || data[i] == tab {
			continue
		}

		if !bytes.HasPrefix(data[i:], prefix) {
			return ValidationError{
				Message:  fmt.Sprintf("block_comment prefix %q was expected inside a block comment", string(prefix)),
				Position: i,
			}
		}

		break
	}

	return nil
}

// isBlockCommentEnd tells you when a block comment end on this line.
func isBlockCommentEnd(end []byte, data []byte) bool {
	for i := len(data) - 1; i > 0; i-- {
		if data[i] == cr || data[i] == lf {
			continue
		}

		return bytes.HasSuffix(data[:i], end)
	}

	return false
}

// MaxLineLength checks the length of a given line.
//
// It assumes UTF-8 and will count as one runes. The first byte has no prefix
// 0xxxxxxx, 110xxxxx, 1110xxxx, 11110xxx, 111110xx, etc. and the following byte
// the 10xxxxxx prefix which are skipped.
func MaxLineLength(maxLength int, tabWidth int, data []byte) error {
	length := 0
	breakingPosition := 0

	for i := 0; i < len(data); i++ {
		if data[i] == cr || data[i] == lf {
			break
		}

		switch {
		case data[i] == tab:
			length += tabWidth
		case (data[i] >> 6) == 0b10:
			// skip 0x10xxxxxx that are UTF-8 continuation markers
		default:
			length++
		}

		if length > maxLength && breakingPosition == 0 {
			breakingPosition = i
		}
	}

	if length > maxLength {
		return ValidationError{
			Message:  fmt.Sprintf("line is too long (%d > %d)", length, maxLength),
			Position: breakingPosition,
		}
	}

	return nil
}
