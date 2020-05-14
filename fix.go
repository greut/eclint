package eclint

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/go-logr/logr"
)

// FixWithDefinition does the hard work of validating the given file.
func FixWithDefinition(d *editorconfig.Definition, filename string, log logr.Logger) error {
	def, err := newDefinition(d)
	if err != nil {
		return err
	}

	stat, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("cannot stat %s. %w", filename, err)
	}

	if stat.IsDir() {
		log.V(2).Info("skipped directory")
		return nil
	}

	fileSize := stat.Size()
	mode := stat.Mode()

	r, err := fixWithFilename(def, filename, fileSize, log)
	if err != nil {
		return err
	}

	if r == nil {
		return nil
	}

	// XXX keep mode as is.
	fp, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer fp.Close()

	n, err := io.Copy(fp, r)
	log.V(1).Info("bytes written", "total", n)

	return err
}

func fixWithFilename(def *definition, filename string, fileSize int64, log logr.Logger) (io.Reader, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s. %w", filename, err)
	}

	defer fp.Close()

	r := bufio.NewReader(fp)

	ok, err := probeReadable(fp, r)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s. %w", filename, err)
	}

	if !ok {
		log.V(2).Info("skipped unreadable or empty file")
		return nil, nil
	}

	charset, isBinary, err := ProbeCharsetOrBinary(r, def.Charset, log)
	if err != nil {
		return nil, err
	}

	if isBinary {
		log.V(2).Info("binary file detected and skipped")
		return nil, nil
	}

	log.V(2).Info("charset probed", "charset", charset)

	return fix(r, fileSize, charset, log, def)
}

func fix(r io.Reader, fileSize int64, charset string, log logr.Logger, def *definition) (io.Reader, error) {
	buf := bytes.NewBuffer([]byte{})

	errs := ReadLines(r, fileSize, func(index int, data []byte, isEOF bool) error {
		if def.EndOfLine != "" && !isEOF {
			data = bytes.TrimRight(data, "\r\n")

			switch def.EndOfLine {
			case "cr":
				data = append(data, '\r')
			case "crlf":
				data = append(data, '\r', '\n')
			case "lf":
				data = append(data, '\n')
			default:
				return fmt.Errorf("unsupported EndOfLine value %s", def.EndOfLine)
			}
		}

		_, err := buf.Write(data)
		return err
	})

	if len(errs) != 0 {
		return nil, errs[0]
	}

	return buf, nil
}
