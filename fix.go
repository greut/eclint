package eclint

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/go-logr/logr"
)

// FixWithDefinition does the hard work of validating the given file.
func FixWithDefinition(ctx context.Context, d *editorconfig.Definition, filename string) error {
	def, err := newDefinition(d)
	if err != nil {
		return err
	}

	stat, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("cannot stat %s. %w", filename, err)
	}

	log := logr.FromContextOrDiscard(ctx)

	if stat.IsDir() {
		log.V(2).Info("skipped directory")

		return nil
	}

	fileSize := stat.Size()
	mode := stat.Mode()

	r, err := fixWithFilename(ctx, def, filename, fileSize)
	if err != nil {
		return fmt.Errorf("cannot fix %s: %w", filename, err)
	}

	if r == nil {
		return nil
	}

	// XXX keep mode as is.
	fp, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC, mode) //nolint:nosnakecase
	if err != nil {
		return fmt.Errorf("cannot open %s using %s: %w", filename, mode, err)
	}
	defer fp.Close()

	n, err := io.Copy(fp, r)
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	log.V(1).Info("bytes written", "total", n)

	return nil
}

func fixWithFilename(ctx context.Context, def *definition, filename string, fileSize int64) (io.Reader, error) {
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

	log := logr.FromContextOrDiscard(ctx)

	if !ok {
		log.V(2).Info("skipped unreadable or empty file")

		return nil, nil
	}

	charset, isBinary, err := ProbeCharsetOrBinary(ctx, r, def.Charset)
	if err != nil {
		return nil, err
	}

	if isBinary {
		log.V(2).Info("binary file detected and skipped")

		return nil, nil
	}

	log.V(2).Info("charset probed", "charset", charset)

	return fix(ctx, r, fileSize, charset, def)
}

func fix( //nolint:funlen
	_ context.Context,
	r io.Reader,
	fileSize int64,
	_ string,
	def *definition,
) (io.Reader, error) {
	buf := bytes.NewBuffer([]byte{})

	size := def.IndentSize
	if def.TabWidth != 0 {
		size = def.TabWidth
	}

	if size == 0 {
		// Indent size default == 2
		size = 2
	}

	var c []byte

	var x []byte

	switch def.IndentStyle {
	case SpaceValue:
		c = bytes.Repeat([]byte{space}, size)
		x = []byte{tab}
	case TabValue:
		c = []byte{tab}
		x = bytes.Repeat([]byte{space}, size)
	case "", UnsetValue:
		size = 0
	default:
		return nil, fmt.Errorf(
			"%w: %q is an invalid value of indent_style, want tab or space",
			ErrConfiguration,
			def.IndentStyle,
		)
	}

	eol, err := def.EOL()
	if err != nil {
		return nil, fmt.Errorf("cannot get EOL: %w", err)
	}

	trimTrailingWhitespace := false
	if def.TrimTrailingWhitespace != nil {
		trimTrailingWhitespace = *def.TrimTrailingWhitespace
	}

	errs := ReadLines(r, fileSize, func(index int, data []byte, isEOF bool) error {
		if size != 0 {
			data = fixTabAndSpacePrefix(data, c, x)
		}

		if trimTrailingWhitespace {
			data = fixTrailingWhitespace(data)
		}

		if def.EndOfLine != "" && !isEOF {
			data = bytes.TrimRight(data, "\r\n")

			data = append(data, eol...)
		}

		_, err := buf.Write(data)
		if err != nil {
			return fmt.Errorf("error writing into buffer: %w", err)
		}

		return nil
	})

	if len(errs) != 0 {
		return nil, errs[0]
	}

	return buf, nil
}

// fixTabAndSpacePrefix replaces any `x` by `c` in the given `data`.
func fixTabAndSpacePrefix(data []byte, c []byte, x []byte) []byte {
	newData := make([]byte, 0, len(data))

	i := 0
	for i < len(data) {
		if bytes.HasPrefix(data[i:], c) {
			i += len(c)

			newData = append(newData, c...)

			continue
		}

		if bytes.HasPrefix(data[i:], x) {
			i += len(x)

			newData = append(newData, c...)

			continue
		}

		return append(newData, data[i:]...)
	}

	return data
}

// fixTrailingWhitespace replaces any whitespace or tab from the end of the line.
func fixTrailingWhitespace(data []byte) []byte {
	i := len(data) - 1

	// u -> v is the range to clean
	u := len(data)

	v := u //nolint: ifshort

outer:
	for i >= 0 {
		switch data[i] {
		case '\r', '\n':
			i--
			u--
			v--
		case ' ', '\t':
			i--
			u--
		default:
			break outer
		}
	}

	if u != v {
		data = append(data[:u], data[v:]...)
	}

	return data
}
