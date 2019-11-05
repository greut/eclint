package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/go-logr/logr"
)

// validate is where the validations rules are applied
func validate(r io.Reader, log logr.Logger, def *editorconfig.Definition) []error {
	var buf *bytes.Buffer
	// chardet uses a 8192 bytebuf for detection
	bufSize := 8192

	indentSize, _ := strconv.Atoi(def.IndentSize)

	var lastLine []byte
	errs := readLines(r, func(index int, data []byte) error {
		var err error

		// The first line may contain the BOM for detecting some encodings
		if index == 0 && def.Charset != "" {
			ok, err := charsetUsingBOM(def.Charset, data)
			if err != nil {
				return err
			}

			if !ok {
				buf = bytes.NewBuffer(make([]byte, 0))
			}
		}

		// The last line may not have the expected ending.
		if lastLine != nil && def.EndOfLine != "" {
			err = endOfLine(def.EndOfLine, lastLine)
		}

		lastLine = data

		if buf != nil && buf.Len() < bufSize {
			if _, err := buf.Write(data); err != nil {
				log.Error(err, "cannot write into file buffer", "line", index)
			}
		}

		if err == nil && def.IndentStyle != "" {
			err = indentStyle(def.IndentStyle, indentSize, data)
		}

		if err == nil && def.TrimTrailingWhitespace != nil && *def.TrimTrailingWhitespace {
			err = trimTrailingWhitespace(data)
		}

		if err != nil {
			return fmt.Errorf("line %d: %s", index, err)
		}

		return nil
	})

	if buf != nil && buf.Len() > 0 {
		err := charset(def.Charset, buf.Bytes())
		errs = append(errs, err)
	}

	if lastLine != nil && def.InsertFinalNewline != nil {
		var lastChar byte
		if len(lastLine) > 0 {
			lastChar = lastLine[len(lastLine)-1]
		}

		if lastChar != 0x0 && lastChar != '\r' && lastChar != '\n' {
			if *def.InsertFinalNewline {
				err := fmt.Errorf("missing the final newline")
				errs = append(errs, err)
			}
		} else {
			if def.EndOfLine != "" {
				err := endOfLine(def.EndOfLine, lastLine)
				errs = append(errs, err)
			}

			if !*def.InsertFinalNewline {
				err := fmt.Errorf("found an extraneous final newline")
				errs = append(errs, err)
			}
		}
	}

	return errs
}

func lint(filename string, log logr.Logger) []error {
	// XXX editorconfig should be able to treat a flux of
	// filenames with caching capabilities.
	def, err := editorconfig.GetDefinitionForFilename(filename)
	if err != nil {
		return []error{fmt.Errorf("cannot open file %s. %w", filename, err)}
	}
	log.V(1).Info("lint", "filename", filename)

	fp, err := os.Open(filename)
	if err != nil {
		return []error{err}
	}
	defer fp.Close()

	return validate(fp, log, def)
}
