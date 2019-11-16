package main

import (
	"bytes"
	"fmt"

	"github.com/gogs/chardet"
)

const (
	cr    = '\r'
	lf    = '\n'
	tab   = '\t'
	space = ' '
)

// validationError is a rich type containing information about the error
type validationError struct {
	error    string
	filename string
	line     []byte
	index    int
	position int
}

func (e validationError) String() string {
	return e.Error()
}

// Error builds the error string.
func (e validationError) Error() string {
	return fmt.Sprintf("%d:%d: %s", e.index, e.position, e.error)
}

// endOfLines checks the line ending
func endOfLine(eol string, data []byte) error {
	switch eol {
	case "lf":
		if !bytes.HasSuffix(data, []byte{lf}) || bytes.HasSuffix(data, []byte{cr, lf}) {
			return validationError{
				error:    "line does not end with lf (`\\n`)",
				position: len(data),
			}
		}
	case "crlf":
		if !bytes.HasSuffix(data, []byte{cr, lf}) {
			return validationError{
				error:    "line does not end with crlf (`\\r\\n`)",
				position: len(data),
			}
		}
	case "cr":
		if !bytes.HasSuffix(data, []byte{cr}) {
			return validationError{
				error:    "line does not end with cr (`\\r`)",
				position: len(data),
			}
		}
	default:
		return fmt.Errorf("%q is an invalid value for eol, want cr, crlf, or lf", eol)
	}

	return nil
}

// charsetUsingBOM checks the charset via the first bytes of the first line
func charsetUsingBOM(charset string, data []byte) (bool, error) {
	switch charset {
	case "utf-8 bom":
		if !bytes.HasPrefix(data, []byte{0xef, 0xbb, 0xbf}) {
			return false, validationError{error: "no UTF-8 BOM were found"}
		}
	case "utf-16le":
		if !bytes.HasPrefix(data, []byte{0xff, 0xfe}) {
			return false, validationError{error: "no UTF-16LE BOM were found"}
		}
	case "utf-16be":
		if !bytes.HasPrefix(data, []byte{0xfe, 0xff}) {
			return false, validationError{error: "no UTF-16BE BOM were found"}
		}
	case "utf-32le":
		if !bytes.HasPrefix(data, []byte{0xff, 0xfe, 0, 0}) {
			return false, validationError{error: "no UTF-32LE BOM were found"}
		}
	case "utf-32be":
		if !bytes.HasPrefix(data, []byte{0, 0, 0xfe, 0xff}) {
			return false, validationError{error: "no UTF-32BE BOM were found"}
		}
	default:
		return false, nil
	}
	return true, nil
}

// charset checks the file encoding
func charset(charset string, data []byte) error {
	d := chardet.NewTextDetector()
	results, err := d.DetectAll(data)
	if err != nil {
		return fmt.Errorf("charset detection failure %s", err)
	}

	for _, result := range results {
		switch charset {
		case "utf-8":
			if result.Charset == "UTF-8" {
				return nil
			}
		case "latin1":
			if result.Charset == "ISO-8859-1" {
				return nil
			}
		default:
			return fmt.Errorf("charset %q is invalid or should have been detected using its BOM already", charset)
		}
	}

	if len(results) > 0 {
		return validationError{
			error: fmt.Sprintf("detected charset %q does not match expected %q", results[0].Charset, charset),
		}
	}

	return nil
}

// indentStyle checks that the line beginnings are either space or tabs
func indentStyle(style string, size int, data []byte) error {
	var c byte
	var x byte
	switch style {
	case "space":
		c = space
		x = tab
	case "tab":
		c = tab
		x = space
		size = 1
	case "unset":
		return nil
	default:
		return fmt.Errorf("%q is an invalid value of indent_style, want tab or space", style)
	}

	for i := 0; i < len(data); i++ {
		if data[i] == c {
			continue
		}
		if data[i] == x {
			return validationError{
				error:    fmt.Sprintf("indentation style mismatch expected %q (%s) got %q", c, style, x),
				position: i + 1,
			}
		}
		if data[i] == cr || data[i] == lf || (size > 0 && i%size == 0) {
			break
		}
		return validationError{
			error:    fmt.Sprintf("indentation size doesn't match expected %d, got %d", size, i),
			position: i + 1,
		}
	}

	return nil
}

// trimTrailingWhitespace
func trimTrailingWhitespace(data []byte) error {
	for i := len(data) - 1; i >= 0; i-- {
		if data[i] == cr || data[i] == lf {
			continue
		}
		if data[i] == space || data[i] == tab {
			return validationError{
				error:    "line has some trailing whitespaces",
				position: i + 1,
			}
		}
		break
	}
	return nil
}

// isBlockCommentStart tells you when a block comment started on this line
func isBlockCommentStart(start []byte, data []byte) bool {
	for i := 0; i < len(data); i++ {
		if data[i] == space || data[i] == tab {
			continue
		}
		return bytes.HasPrefix(data[i:], start)
	}
	return false
}

// checkBlockComment checks the line is a valid block comment
func checkBlockComment(i int, prefix []byte, data []byte) error {
	for ; i < len(data); i++ {
		if data[i] == space || data[i] == tab {
			continue
		}
		if !bytes.HasPrefix(data[i:], prefix) {
			return validationError{
				error:    fmt.Sprintf("block_comment prefix %q was expected inside a block comment", string(prefix)),
				position: i + 1,
			}
		}
		break
	}
	return nil
}

// isBlockCommentEnd tells you when a block comment end on this line
func isBlockCommentEnd(end []byte, data []byte) bool {
	for i := len(data) - 1; i > 0; i-- {
		if data[i] == cr || data[i] == lf {
			continue
		}
		return bytes.HasSuffix(data[:i], end)
	}
	return false
}

// maxLineLength checks the length of a given line
func maxLineLength(maxLength int, tabWidth int, data []byte) error {
	length := 0
	breakingPosition := 0
	for i := 0; i < len(data); i++ {
		if data[i] == cr || data[i] == lf {
			break
		}
		if data[i] == tab {
			length += tabWidth
		} else {
			length++
		}
		if length > maxLength && breakingPosition == 0 {
			breakingPosition = i
		}
	}

	if length > maxLength {
		return validationError{
			error:    fmt.Sprintf("line is too long (%d > %d)", length+1, maxLength),
			position: breakingPosition,
		}
	}

	return nil
}
