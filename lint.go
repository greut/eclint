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

// DefaultTabWidth sets the width of a tab used when counting the line length
const DefaultTabWidth = 8

const (
	// UnsetValue is the value equivalent to an empty / unset one.
	UnsetValue = "unset"
	// TabValue is the value representing tab indentation (the ugly one)
	TabValue = "tab"
	// SpaceValue is the value representing space indentation (the good one)
	SpaceValue = "space"
	// Utf8 is the ubiquitous character set
	Utf8 = "utf-8"
)

// validate is where the validations rules are applied
func validate( //nolint:funlen,gocyclo
	r io.Reader, charset string,
	log logr.Logger,
	def *definition,
) []error {
	var lastLine []byte

	var lastIndex int

	errs := ReadLines(r, func(index int, data []byte) error {
		var err error

		// The last line may not have the expected ending.
		if lastLine != nil && def.EndOfLine != "" && def.EndOfLine != UnsetValue {
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

		if def.IndentStyle != "" && def.IndentStyle != UnsetValue && def.Definition.IndentSize != UnsetValue {
			if def.InsideBlockComment && def.BlockCommentEnd != nil {
				def.InsideBlockComment = !isBlockCommentEnd(def.BlockCommentEnd, data)
			}

			err = indentStyle(def.IndentStyle, def.IndentSize, data)
			if err != nil && def.InsideBlockComment && def.BlockComment != nil {
				// The indentation may fail within a block comment.
				if ve, ok := err.(ValidationError); ok {
					err = checkBlockComment(ve.Position, def.BlockComment, data)
				}
			}

			if err == nil && !def.InsideBlockComment && def.BlockCommentStart != nil {
				def.InsideBlockComment = isBlockCommentStart(def.BlockCommentStart, data)
			}
		}

		if err == nil && def.TrimTrailingWhitespace != nil && *def.TrimTrailingWhitespace {
			err = trimTrailingWhitespace(data)
		}

		if err == nil && def.MaxLength > 0 {
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
			err = MaxLineLength(def.MaxLength, def.TabWidth, d)
		}

		// Enrich the error with the line number
		if ve, ok := err.(ValidationError); ok {
			ve.Line = data
			ve.Index = index
			return ve
		}

		return err
	})

	if lastLine != nil && def.InsertFinalNewline != nil {
		var err error

		var lastChar byte

		if len(lastLine) > 0 {
			lastChar = lastLine[len(lastLine)-1]
		}

		if lastChar != 0x0 {
			if lastChar != cr && lastChar != lf {
				if *def.InsertFinalNewline {
					err = fmt.Errorf("missing the final newline")
				}
			} else {
				if !*def.InsertFinalNewline {
					err = fmt.Errorf("found an extraneous final newline")
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

// Lint does the hard work of validating the given file.
func Lint(filename string, log logr.Logger) []error {
	def, err := editorconfig.GetDefinitionForFilename(filename)
	if err != nil {
		return []error{fmt.Errorf("cannot open file %s. %w", filename, err)}
	}

	return LintWithDefinition(def, filename, log)
}

// LintWithDefinition does the hard work of validating the given file.
func LintWithDefinition(d *editorconfig.Definition, filename string, log logr.Logger) []error {
	def, err := newDefinition(d)
	if err != nil {
		return []error{err}
	}

	fp, err := os.Open(filename)
	if err != nil {
		return []error{fmt.Errorf("cannot open %s. %w", filename, err)}
	}

	defer fp.Close()

	r := bufio.NewReader(fp)

	ok, err := probeReadable(fp, r)
	if err != nil {
		return []error{fmt.Errorf("cannot read %s. %w", filename, err)}
	}

	if !ok {
		log.V(2).Info("skipped unreadable or empty file")
		return nil
	}

	charset, isBinary, err := ProbeCharsetOrBinary(r, def.Charset, log)
	if err != nil {
		return []error{err}
	}

	if isBinary {
		log.V(2).Info("binary file detected and skipped")
		return nil
	}

	log.V(2).Info("charset probed", "charset", charset)

	errs := validate(r, charset, log, def)

	// Enrich the errors with the filename
	for i, err := range errs {
		if ve, ok := err.(ValidationError); ok {
			ve.Filename = filename
			errs[i] = ve
		} else if err != nil {
			errs[i] = err
		}
	}

	return errs
}
