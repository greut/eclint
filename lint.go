package eclint

import (
	"bufio"
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

// UnsetValue is the value equivalent to an empty / unset one.
const UnsetValue = "unset"

// validate is where the validations rules are applied
func validate(r io.Reader, log logr.Logger, def *editorconfig.Definition) []error { //nolint:funlen,gocyclo
	var buf *bytes.Buffer
	// chardet uses a 8192 bytebuf for detection
	bufSize := 8192

	indentSize, _ := strconv.Atoi(def.IndentSize)

	var charset string
	var lastLine []byte
	var lastIndex int

	var insideBlockComment bool
	var blockCommentStart []byte
	var blockComment []byte
	var blockCommentEnd []byte
	if def.IndentStyle != "" && def.IndentStyle != UnsetValue {
		bs, ok := def.Raw["block_comment_start"]
		if ok && bs != "" && bs != UnsetValue {
			blockCommentStart = []byte(bs)
			bc, ok := def.Raw["block_comment"]
			if ok && bc != "" && bs != UnsetValue {
				blockComment = []byte(bc)
			}

			be, ok := def.Raw["block_comment_end"]
			if !ok || be == "" || be == UnsetValue {
				return []error{fmt.Errorf("block_comment_end was expected, none were found")}
			}
			blockCommentEnd = []byte(be)
		}
	}

	maxLength := 0
	tabWidth := def.TabWidth
	if mll, ok := def.Raw["max_line_length"]; ok && mll != "off" && mll != UnsetValue {
		ml, err := strconv.Atoi(mll)
		if err != nil || ml < 0 {
			return []error{fmt.Errorf("max_line_length expected a non-negative number, got %s", mll)}
		}
		maxLength = ml
		if tabWidth <= 0 {
			tabWidth = DefaultTabWidth
		}
	}

	errs := ReadLines(r, func(index int, data []byte) error {
		var err error

		// The first line may contain the BOM for detecting some encodings
		if index == 0 {
			if def.Charset != "utf-8" && def.Charset != "latin1" {
				charset = detectCharsetUsingBOM(data)

				if def.Charset != "" && charset != def.Charset {
					return ValidationError{
						Message:  fmt.Sprintf("no %s prefix were found (got %q)", def.Charset, charset),
						Position: 0,
						Index:    index,
						Line:     data,
					}
				}
			}

			if charset == "" && def.Charset != "" {
				buf = bytes.NewBuffer(make([]byte, 0))
			}
		}

		// The last line may not have the expected ending.
		if lastLine != nil && def.EndOfLine != "" {
			err = endOfLine(def.EndOfLine, lastLine)
			// XXX not so nice hack
			if ve, ok := err.(ValidationError); ok {
				ve.Line = lastLine
				ve.Index = lastIndex

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

		if def.IndentStyle != "" && def.IndentStyle != UnsetValue {
			if insideBlockComment && blockCommentEnd != nil {
				insideBlockComment = !isBlockCommentEnd(blockCommentEnd, data)
			}

			err = indentStyle(def.IndentStyle, indentSize, data)
			if err != nil && insideBlockComment && blockComment != nil {
				// The indentation may fail within a block comment.
				if ve, ok := err.(ValidationError); ok {
					err = checkBlockComment(ve.Position, blockComment, data)
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
			// Remove any BOM from the first line.
			d := data
			if index == 0 && charset != "" {
				for _, bom := range [][]byte{utf8Bom} {
					if bytes.HasPrefix(data, bom) {
						d = data[len(utf8Bom):]
						break
					}
				}
			}
			err = MaxLineLength(maxLength, tabWidth, d)
		}

		// Enrich the error with the line number
		if ve, ok := err.(ValidationError); ok {
			ve.Line = data
			ve.Index = index
			return ve
		}

		return err
	})

	if buf != nil && buf.Len() > 0 {
		err := detectCharset(def.Charset, buf.Bytes())
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
			if ve, ok := err.(ValidationError); ok {
				ve.Line = lastLine
				ve.Index = lastIndex
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
				return fmt.Errorf("%v cannot be overridden yet, pr welcome", nk)
			case "insert_final_newline":
				return fmt.Errorf("%v cannot be overridden yet, pr welcome", nk)
			}
		}
	}
	return nil
}

// Lint does the hard work of validating the given file.
func Lint(filename string, log logr.Logger) []error {
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

	r := bufio.NewReader(fp)

	errs := validate(r, log, def)

	// Enrich the errors with the filename
	for i, err := range errs {
		if ve, ok := err.(ValidationError); ok {
			ve.Filename = filename
			errs[i] = ve
		} else if err != nil {
			errs[i] = fmt.Errorf("%s:%w", filename, err)
		}
	}

	return errs
}
