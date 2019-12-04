// +build linux

package eclint

import "github.com/rakyll/magicmime"

// LoadMime opens the libmagic database
func LoadMime() error {
	return magicmime.Open(magicmime.MAGIC_NONE)
}

// UnloadMime closes the libmagic database
func UnloadMime() {
	magicmime.Close()
}

// TypeByBuffer return the type as discovered by libmagic
func TypeByBuffer(buf []byte) (string, error) {
	return magicmime.TypeByBuffer(buf)
}
