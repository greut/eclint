package main

import "fmt"

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
		if l > 0 && data[l-1] != '\n' || (l > 1 && data[l-2] != '\r') {
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
