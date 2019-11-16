package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/go-logr/logr"
)

// DefaultTabWidth sets the width of a tab used when counting the line length
const DefaultTabWidth = 8

// validate is where the validations rules are applied
func validate(r io.Reader, log logr.Logger, def *editorconfig.Definition) []error { //nolint:gocyclo
	var buf *bytes.Buffer
	// chardet uses a 8192 bytebuf for detection
	bufSize := 8192

	indentSize, _ := strconv.Atoi(def.IndentSize)

	var lastLine []byte
	var lastIndex int

	var insideBlockComment bool
	var blockCommentStart []byte
	var blockComment []byte
	var blockCommentEnd []byte
	if def.IndentStyle != "" && def.IndentStyle != "unset" {
		bs, ok := def.Raw["block_comment_start"]
		if ok && bs != "" && bs != "unset" {
			blockCommentStart = []byte(bs)
			bc, ok := def.Raw["block_comment"]
			if ok && bc != "" && bs != "unset" {
				blockComment = []byte(bc)
			}

			be, ok := def.Raw["block_comment_end"]
			if !ok || be == "" || be == "unset" {
				return []error{fmt.Errorf("block_comment_end was expected, none were found")}
			}
			blockCommentEnd = []byte(be)
		}
	}

	maxLength := 0
	tabWidth := def.TabWidth
	if mll, ok := def.Raw["max_line_length"]; ok && mll != "off" && mll != "unset" {
		ml, err := strconv.Atoi(mll)
		if err != nil || ml < 0 {
			return []error{fmt.Errorf("max_line_length expected a non-negative number, got %s", mll)}
		}
		maxLength = ml
		if tabWidth <= 0 {
			tabWidth = DefaultTabWidth
		}
	}

	errs := readLines(r, func(index int, data []byte) error {
		var err error

		// The first line may contain the BOM for detecting some encodings
		if index == 1 && def.Charset != "" {
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
			// XXX not so nice hack
			if ve, ok := err.(validationError); ok {
				ve.line = lastLine
				ve.index = lastIndex

				lastLine = data
				lastIndex = index

				return ve
			}
		}

		lastLine = data
		lastIndex = index

		if buf != nil && buf.Len() < bufSize {
			if _, err := buf.Write(data); err != nil {
				log.Error(err, "cannot write into file buffer", "line", index)
			}
		}

		if def.IndentStyle != "" && def.IndentStyle != "unset" {
			if insideBlockComment && blockCommentEnd != nil {
				insideBlockComment = !isBlockCommentEnd(blockCommentEnd, data)
			}

			err = indentStyle(def.IndentStyle, indentSize, data)
			if err != nil && insideBlockComment && blockComment != nil {
				// The indentation may fail within a block comment.
				if ve, ok := err.(validationError); ok {
					err = checkBlockComment(ve.position-1, blockComment, data)
				}
			}

			if err == nil && !insideBlockComment && blockCommentStart != nil {
				insideBlockComment = isBlockCommentStart(blockCommentStart, data)
			}
		}

		if err == nil && def.TrimTrailingWhitespace != nil && *def.TrimTrailingWhitespace {
			err = trimTrailingWhitespace(data)
		}

		if err == nil && maxLength > 0 && tabWidth > 0 {
			err = maxLineLength(maxLength, tabWidth, data)
		}

		// Enrich the error with the line number
		if ve, ok := err.(validationError); ok {
			ve.line = data
			ve.index = index
			return ve
		}

		return err
	})

	if buf != nil && buf.Len() > 0 {
		err := charset(def.Charset, buf.Bytes())
		if err != nil {
			errs = append(errs, err)
		}
	}

	if lastLine != nil && def.InsertFinalNewline != nil {
		var err error
		var lastChar byte
		if len(lastLine) > 0 {
			lastChar = lastLine[len(lastLine)-1]
		}

		if lastChar != 0x0 && lastChar != cr && lastChar != lf {
			if *def.InsertFinalNewline {
				err = fmt.Errorf("missing the final newline")
			}
		} else {
			if def.EndOfLine != "" {
				err = endOfLine(def.EndOfLine, lastLine)
			}

			if err != nil {
				if !*def.InsertFinalNewline {
					err = fmt.Errorf("found an extraneous final newline")
				} else {
					err = nil
				}
			}
		}

		if err != nil {
			if ve, ok := err.(validationError); ok {
				ve.line = lastLine
				ve.index = lastIndex
				errs = append(errs, ve)
			} else {
				errs = append(errs, err)
			}
		}
	}

	return errs
}

func overrideUsingPrefix(def *editorconfig.Definition, prefix string) error {
	for k, v := range def.Raw {
		if strings.HasPrefix(k, prefix) {
			nk := k[len(prefix):]
			def.Raw[nk] = v
			switch nk {
			case "indent_style":
				def.IndentStyle = v
			case "indent_size":
				def.IndentSize = v
			case "charset":
				def.Charset = v
			case "end_of_line":
				def.EndOfLine = v
			case "tab_width":
				i, err := strconv.Atoi(v)
				if err != nil {
					return fmt.Errorf("tab_width cannot be set. %w", err)
				}
				def.TabWidth = i
			case "trim_trailing_whitespace":
				return fmt.Errorf("%v cannot be overriden yet, pr welcome", nk)
			case "insert_final_newline":
				return fmt.Errorf("%v cannot be overriden yet, pr welcome", nk)
			}
		}
	}
	return nil
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

	err = overrideUsingPrefix(def, "eclint_")
	if err != nil {
		return []error{err}
	}

	errs := validate(fp, log, def)

	// Enrich the errors with the filename
	for i, err := range errs {
		if ve, ok := err.(validationError); ok {
			ve.filename = filename
			errs[i] = ve
		} else if err != nil {
			errs[i] = fmt.Errorf("%s:%w", filename, err)
		}
	}

	return errs
}
