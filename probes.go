package eclint

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/gogs/chardet"
)

// ProbeCharsetOrBinary does all the probes to detect the encoding
// or whether it is a binary file.
func ProbeCharsetOrBinary(ctx context.Context, r *bufio.Reader, charset string) (string, bool, error) {
	bs, err := r.Peek(512)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", false, fmt.Errorf("cannot peek into reader: %w", err)
	}

	isBinary := probeMagic(ctx, bs)

	if !isBinary {
		isBinary = probeBinary(ctx, bs)
	}

	if isBinary {
		return "", true, nil
	}

	cs, err := probeCharset(ctx, bs, charset)
	if err != nil {
		return "", false, fmt.Errorf("cannot probe charset: %w", err)
	}

	return cs, false, nil
}

// probeMagic searches for some text-baesd binary files such as PDF.
func probeMagic(ctx context.Context, bs []byte) bool {
	log := logr.FromContextOrDiscard(ctx)

	if bytes.HasPrefix(bs, []byte("%PDF-")) {
		log.V(2).Info("magic for PDF was found", "prefix", bs[0:7])

		return true
	}

	return false
}

// probeBinary tells if the reader is likely to be binary
//
// checking for \0 is a weak strategy.
func probeBinary(_ context.Context, bs []byte) bool {
	cont := 0

	l := len(bs)
	for i := 0; i < l; i++ {
		b := bs[i]

		switch {
		case b&0b1000_0000 == 0x00:
			continue

		case b&0b1100_0000 == 0b1000_0000:
			// found continuation, probably binary
			return true

		case b&0b1110_0000 == 0b1100_0000:
			// found leading of two bytes
			cont = 1
		case b&0b1111_0000 == 0b1110_0000:
			// found leading of three bytes
			cont = 2
		case b&0b1111_1000 == 0b1111_0000:
			// found leading of four bytes
			cont = 3
		case b == 0x00:
			// found NUL byte, probably binary
			return true
		}

		for ; cont > 0 && i < l-1; cont-- {
			i++
			b = bs[i]

			if b&0b1100_0000 != 0b1000_0000 {
				// found something different than a continuation,
				// probably binary
				return true
			}
		}
	}

	return cont > 0
}

func probeCharset(ctx context.Context, bs []byte, charset string) (string, error) {
	log := logr.FromContextOrDiscard(ctx)

	// empty files are valid text files
	if len(bs) == 0 {
		return charset, nil
	}

	var cs string
	// The first line may contain the BOM for detecting some encodings
	if charset != Utf8 && charset != Latin1 {
		cs = detectCharsetUsingBOM(bs)

		if charset != "" && cs != charset {
			return "", ValidationError{
				Message: fmt.Sprintf("no %s prefix were found, got %q", charset, cs),
			}
		}

		log.V(3).Info("detect using BOM", "charset", charset)
	}

	if cs == "" && charset != "" {
		c, err := detectCharset(charset, bs)
		if err != nil {
			return "", err
		}

		cs = c

		// latin1 is a strict subset of utf-8
		if charset != cs {
			return "", ValidationError{
				Message: fmt.Sprintf("detected charset %q does not match expected %q", cs, charset),
			}
		}

		log.V(3).Info("detect using chardet", "charset", charset)
	}

	return cs, nil
}

// probeReadable tries to read the file. When empty or a directory
// it's considered non-readable with no errors. Otherwise the error
// should be caught.
func probeReadable(fp *os.File, r *bufio.Reader) (bool, error) {
	// Sanity check that the file can be read.
	_, err := r.Peek(1)
	if err != nil && !errors.Is(err, io.EOF) {
		if !errors.Is(err, io.EOF) {
			return false, nil
		}

		fi, err := fp.Stat()
		if err != nil {
			return false, fmt.Errorf("cannot stat file: %w", err)
		}

		if fi.IsDir() {
			return false, nil
		}

		return false, nil
	}

	return true, nil
}

// detectCharsetUsingBOM checks the charset via the first bytes of the first line.
func detectCharsetUsingBOM(data []byte) string {
	switch {
	case bytes.HasPrefix(data, utf32leBom):
		return "utf-32le"
	case bytes.HasPrefix(data, utf32beBom):
		return "utf-32be"
	case bytes.HasPrefix(data, utf16leBom):
		return "utf-16le"
	case bytes.HasPrefix(data, utf16beBom):
		return "utf-16be"
	case bytes.HasPrefix(data, utf8Bom):
		return "utf-8 bom"
	}

	return ""
}

// detectCharset detects the file encoding.
func detectCharset(charset string, data []byte) (string, error) {
	if charset == "" {
		return charset, nil
	}

	d := chardet.NewTextDetector()

	results, err := d.DetectAll(data)
	if err != nil {
		return "", fmt.Errorf("charset detection failure: %w", err)
	}

	hasLatin1 := false
	hasUtf8 := false

	for _, result := range results {
		if strings.HasPrefix(result.Charset, "ISO-8859-") {
			hasLatin1 = true
		} else if result.Charset == "UTF-8" {
			hasUtf8 = true
		}
	}

	if hasUtf8 && charset == Utf8 {
		return charset, nil
	}

	if hasLatin1 && charset == Latin1 {
		return charset, nil
	}

	return "", fmt.Errorf("%w: got the following charset(s) %q which are not supported", ErrConfiguration, charset)
}
