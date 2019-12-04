// +build windows darwin

package eclint

import "errors"

// LoadMime opens the libmagic database
func LoadMime() error {
	return errors.New("libmagic doesn't work on Windows")
}

// UnloadMime closes the libmagic database
func UnloadMime() {}

// TypeByBuffer return the type as discovered by libmagic
func TypeByBuffer(buf []byte) (string, error) {
	return "", errors.New("not implemented error")
}
