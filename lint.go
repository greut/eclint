package eclint

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/go-logr/logr"
)

// DefaultTabWidth sets the width of a tab used when counting the line length.
const DefaultTabWidth = 8

const (
	// UnsetValue is the value equivalent to an empty / unset one.
	UnsetValue = "unset"
	// TabValue is the value representing tab indentation (the ugly one).
	TabValue = "tab"
	// SpaceValue is the value representing space indentation (the good one).
	SpaceValue = "space"
	// Utf8 is the ubiquitous character set.
	Utf8 = "utf-8"
	// Latin1 is the legacy 7-bits character set.
	Latin1 = "latin1"
)

// Lint does the hard work of validating the given file.
func Lint(ctx context.Context, filename string) []error {
	def, err := editorconfig.GetDefinitionForFilename(filename)
	if err != nil {
		return []error{fmt.Errorf("cannot open file %s. %w", filename, err)}
	}

	return LintWithDefinition(ctx, def, filename)
}

// LintWithDefinition does the hard work of validating the given file.
func LintWithDefinition(ctx context.Context, d *editorconfig.Definition, filename string) []error { //nolint:funlen
	log := logr.FromContextOrDiscard(ctx)

	def, err := newDefinition(d)
	if err != nil {
		return []error{err}
	}

	stat, err := os.Stat(filename)
	if err != nil {
		return []error{fmt.Errorf("cannot stat %s. %w", filename, err)}
	}

	if stat.IsDir() {
		log.V(2).Info("skipped directory")

		return nil
	}

	fileSize := stat.Size()

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

	charset, isBinary, err := ProbeCharsetOrBinary(ctx, r, def.Charset)
	if err != nil {
		return []error{err}
	}

	if isBinary {
		log.V(2).Info("binary file detected and skipped")

		return nil
	}

	log.V(2).Info("charset probed", "charset", charset)

	errs := validate(ctx, r, fileSize, charset, def)

	// Enrich the errors with the filename
	for i, err := range errs {
		var ve ValidationError
		if ok := errors.As(err, &ve); ok {
			ve.Filename = filename
			errs[i] = ve
		} else if err != nil {
			errs[i] = err
		}
	}

	return errs
}

// validate is where the validations rules are applied.
func validate( //nolint:cyclop,gocognit,funlen
	ctx context.Context,
	r io.Reader,
	fileSize int64,
	charset string,
	def *definition,
) []error {
	return ReadLines(r, fileSize, func(index int, data []byte, isEOF bool) error {
		var err error

		if ctx.Err() != nil {
			return fmt.Errorf("read lines got interrupted: %w", ctx.Err())
		}

		if isEOF {
			if def.InsertFinalNewline != nil {
				err = checkInsertFinalNewline(data, *def.InsertFinalNewline)
			}
		} else {
			if def.EndOfLine != "" && def.EndOfLine != UnsetValue {
				err = endOfLine(def.EndOfLine, data)
			}
		}

		if err == nil && //nolint:nestif
			def.IndentStyle != "" &&
			def.IndentStyle != UnsetValue &&
			def.Definition.IndentSize != UnsetValue {
			err = indentStyle(def.IndentStyle, def.IndentSize, data)
			if err != nil && def.InsideBlockComment && def.BlockComment != nil {
				// The indentation may fail within a block comment.
				var ve ValidationError
				if ok := errors.As(err, &ve); ok {
					err = checkBlockComment(ve.Position, def.BlockComment, data)
				}
			}

			if def.InsideBlockComment && def.BlockCommentEnd != nil {
				def.InsideBlockComment = !isBlockCommentEnd(def.BlockCommentEnd, data)
			}

			if err == nil && !def.InsideBlockComment && def.BlockCommentStart != nil {
				def.InsideBlockComment = isBlockCommentStart(def.BlockCommentStart, data)
			}
		}

		if err == nil && def.TrimTrailingWhitespace != nil && *def.TrimTrailingWhitespace {
			err = checkTrimTrailingWhitespace(data)
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
		var ve ValidationError
		if ok := errors.As(err, &ve); ok {
			ve.Line = data
			ve.Index = index

			return ve
		}

		return err
	})
}
