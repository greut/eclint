package main

import (
	"fmt"

	"github.com/gogs/chardet"
)

// endOfLines checks the line ending
func endOfLine(eol string, data []byte) error {
	l := len(data)
	switch eol {
	case "lf":
		if l > 0 && data[l-1] != '\n' {
			return fmt.Errorf("line does not end with lf (`\\n`)")
		}
		if l > 1 && data[l-2] == '\r' {
			return fmt.Errorf("line should not end with crlf (`\\r\\n`)")
		}
	case "crlf":
		if (l > 0 && data[l-1] != '\n') || l == 1 || (l > 1 && data[l-2] != '\r') {
			return fmt.Errorf("line does not end with crlf (`\\r\\n`)")
		}
	case "cr":
		if l > 0 && data[l-1] != '\r' {
			return fmt.Errorf("line does not end with cr (`\\r`)")
		}
	default:
		return fmt.Errorf("%q is an invalid value for eol, want cr, crlf, or lf", eol)
	}

	return nil
}

// charset checks the file encoding
func charset(charset string, data []byte) error {
	d := chardet.NewTextDetector()
	results, err := d.DetectAll(data)
	if err != nil {
		return fmt.Errorf("charset detection failure %s", err)
	}

	for _, result := range results {
		log.V(4).Info("Charset", "result", result.Charset, "confidence", result.Confidence, "want", charset)
		switch charset {
		case "utf-8":
			if result.Charset == "UTF-8" {
				return nil
			}
		case "utf-8 bom":
			if result.Charset == "UTF-8" {
				return nil
			}
		case "utf-16be":
			if result.Charset == "UTF-16BE" {
				return nil
			}
		case "utf-16le":
			if result.Charset == "UTF-16LE" {
				return nil
			}
		case "latin1":
			if result.Charset == "ISO-8859-1" {
				return nil
			}
		default:
			return fmt.Errorf("%q is an invalid value for charset", charset)
		}
	}

	if len(results) > 0 {
		return fmt.Errorf("detected charset %q does not match expected %q", results[0].Charset, charset)
	}

	return nil
}

// indentStyle checks that the line beginnings are either space or tabs
func indentStyle(style string, size int, data []byte) error {
	var c byte
	var x byte
	switch style {
	case "space":
		c = ' '
		x = '\t'
	case "tab":
		c = '\t'
		x = ' '
		size = 1
	default:
		return fmt.Errorf("%q is an invalid value of indent_style, want tab or space", style)
	}

	for i := 0; i < len(data); i++ {
		if data[i] == c {
			continue
		}
		if data[i] == x {
			return fmt.Errorf("pos %d: indentation style mismatch expected %s", i, style)
		}
		if data[i] == '\r' || data[i] == '\n' || i%size == 0 {
			break
		}
		return fmt.Errorf("pos %d: indentation size doesn't match expected %d, got %d", i, size, i)
	}

	return nil
}

// trimTrailingWhitespace
func trimTrailingWhitespace(data []byte) error {
	for i := len(data) - 1; i >= 0; i-- {
		if data[i] == '\r' || data[i] == '\n' {
			continue
		}
		if data[i] == ' ' || data[i] == '\t' {
			return fmt.Errorf("pos %d: looks like a trailing whitespace", i)
		}
		break
	}
	return nil
}
