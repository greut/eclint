package eclint

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/gogs/chardet"
)

// probeCharsetOrBinary does all the probes to detect the encoding
// or whether it is a binary file.
func probeCharsetOrBinary(r *bufio.Reader, charset string, log logr.Logger) (string, bool, error) {
	bs, err := r.Peek(512)
	if len(bs) == 0 || (err != nil && err != io.EOF) {
		return "", false, err
	}

	cs, err := probeCharset(bs, charset, log)
	if err != nil {
		return "", false, err
	}
	if cs == "" {
		ok, er := probeMagic(bs, log)
		if !ok && er == nil {
			ok, er = probeBinary(bs, log)
		}
		if er != nil {
			return "", false, fmt.Errorf("cannot probe binary. %w", er)
		}
		return "", ok, nil
	}
	return cs, false, nil
}

// probeMagic searches for some text-baesd binary files such as PDF.
func probeMagic(bs []byte, log logr.Logger) (bool, error) {
	if bytes.HasPrefix(bs, []byte("%PDF-")) {
		log.V(2).Info("magic for PDF was found", "prefix", bs[0:7])
		return true, nil
	}
	return false, nil
}

// probeBinary tells if the reader is likely to be binary
//
// checking for \0 is a weak strategy.
func probeBinary(bs []byte, log logr.Logger) (bool, error) {
	cont := 0
	for _, b := range bs {
		switch {
		case b>>6 == 0x02:
			// found continuation, but no cont available, break
			if cont <= 0 {
				return true, nil
			}
			cont--
		case b>>5 == 0x06:
			// found leading of two bytes
			if cont > 0 {
				return true, nil
			}
			cont = 1
		case b>>4 == 0x0e:
			// found leading of three bytes
			if cont > 0 {
				return true, nil
			}
			cont = 2
		case b>>3 == 0x1e:
			// found leading of four bytes
			if cont > 0 {
				return true, nil
			}
			cont = 3
		case b == '\000':
			return true, nil
		}
	}

	return false, nil
}

func probeCharset(bs []byte, charset string, log logr.Logger) (string, error) {
	// empty files are valid text files
	if len(bs) == 0 {
		return "", nil
	}

	var cs string
	// The first line may contain the BOM for detecting some encodings
	if charset != Utf8 && charset != "latin1" {
		cs = detectCharsetUsingBOM(bs)

		if charset != "" && cs != charset {
			return "", ValidationError{
				Message: fmt.Sprintf("no %s prefix were found (got %q)", charset, cs),
			}
		}
		log.V(3).Info("detect using BOM", "charset", charset)
	}

	if cs == "" {
		cs, err := detectCharset(charset, bs)
		if err != nil {
			return "", err
		}

		if charset != "" && charset != cs {
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
	if err != nil && err != io.EOF {
		fi, err := fp.Stat()
		if err != nil {
			return false, err
		}

		if fi.IsDir() {
			return false, nil
		}

		return false, err
	}
	return err != io.EOF, nil
}

// detectCharsetUsingBOM checks the charset via the first bytes of the first line
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

// detectCharset detects the file encoding
func detectCharset(charset string, data []byte) (string, error) {
	if charset == "" {
		return charset, nil
	}

	d := chardet.NewTextDetector()
	results, err := d.DetectAll(data)
	if err != nil {
		return "", fmt.Errorf("charset detection failure %s", err)
	}

	for i, result := range results {
		if strings.HasPrefix(result.Charset, "ISO-8859-") {
			result.Charset = "ASCII"
		}
		switch result.Charset {
		case "UTF-8":
			return Utf8, nil
		case "ASCII":
			if charset == Utf8 {
				return Utf8, nil
			}
			return "latin1", nil
		default:
			if i == 0 {
				charset = result.Charset
			} else {
				charset = fmt.Sprintf("%s,%s", charset, result.Charset)
			}
		}
	}

	return "", fmt.Errorf("got the following charset(s) %q which are not supported", charset)
}
