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

	r, fixed, err := fixWithFilename(ctx, def, filename, fileSize)
	if err != nil {
		return fmt.Errorf("cannot fix %s: %w", filename, err)
	}

	if !fixed {
		log.V(1).Info("no fixes to apply", "filename", filename)

		return nil
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

	log.V(1).Info("bytes written", "filename", filename, "total", n)

	return nil
}

func fixWithFilename(ctx context.Context, def *definition, filename string, fileSize int64) (io.Reader, bool, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return nil, false, fmt.Errorf("cannot open %s. %w", filename, err)
	}

	defer fp.Close()

	r := bufio.NewReader(fp)

	ok, err := probeReadable(fp, r)
	if err != nil {
		return nil, false, fmt.Errorf("cannot read %s. %w", filename, err)
	}

	log := logr.FromContextOrDiscard(ctx)

	if !ok {
		log.V(2).Info("skipped unreadable or empty file")

		return nil, false, nil
	}

	charset, isBinary, err := ProbeCharsetOrBinary(ctx, r, def.Charset)
	if err != nil {
		return nil, false, err
	}

	if isBinary {
		log.V(2).Info("binary file detected and skipped")

		return nil, false, nil
	}

	log.V(2).Info("charset probed", "charset", charset)

	return fix(ctx, r, fileSize, charset, def)
}

func fix( //nolint:funlen,cyclop
	ctx context.Context,
	r io.Reader,
	fileSize int64,
	_ string,
	def *definition,
) (io.Reader, bool, error) {
	log := logr.FromContextOrDiscard(ctx)

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
		return nil, false, fmt.Errorf(
			"%w: %q is an invalid value of indent_style, want tab or space",
			ErrConfiguration,
			def.IndentStyle,
		)
	}

	eol, err := def.EOL()
	if err != nil {
		return nil, false, fmt.Errorf("cannot get EOL: %w", err)
	}

	trimTrailingWhitespace := false
	if def.TrimTrailingWhitespace != nil {
		trimTrailingWhitespace = *def.TrimTrailingWhitespace
	}

	fixed := false
	errs := ReadLines(r, fileSize, func(index int, data []byte, isEOF bool) error {
		var f bool
		if size != 0 {
			data, f = fixTabAndSpacePrefix(data, c, x)
			fixed = fixed || f
		}

		if trimTrailingWhitespace {
			data, f = fixTrailingWhitespace(data)
			fixed = fixed || f
		}

		if def.EndOfLine != "" && !isEOF {
			data, f = fixEndOfLine(data, eol)
			fixed = fixed || f
		}

		_, err := buf.Write(data)
		if err != nil {
			return fmt.Errorf("error writing into buffer: %w", err)
		}

		log.V(2).Info("fix line", "index", index, "fixed", fixed)

		return nil
	})

	if len(errs) != 0 {
		return nil, false, errs[0]
	}

	if def.InsertFinalNewline != nil {
		f := fixInsertFinalNewline(buf, *def.InsertFinalNewline, eol)
		fixed = fixed || f
	}

	return buf, fixed, nil
}

// fixEndOfLine replaces any non eol suffix by the given one.
func fixEndOfLine(data []byte, eol []byte) ([]byte, bool) {
	fixed := false

	if !bytes.HasSuffix(data, eol) {
		fixed = true
		data = bytes.TrimRight(data, "\r\n")
		data = append(data, eol...)
	}

	return data, fixed
}

// fixTabAndSpacePrefix replaces any `x` by `c` in the given `data`.
func fixTabAndSpacePrefix(data []byte, c []byte, x []byte) ([]byte, bool) {
	newData := make([]byte, 0, len(data))

	fixed := false

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

			fixed = true

			continue
		}

		return append(newData, data[i:]...), fixed
	}

	return data, fixed
}

// fixTrailingWhitespace replaces any whitespace or tab from the end of the line.
func fixTrailingWhitespace(data []byte) ([]byte, bool) {
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

	// If u != v then the line has been fixed.
	fixed := u != v
	if fixed {
		data = append(data[:u], data[v:]...)
	}

	return data, fixed
}

// fixInsertFinalNewline modifies buf to fix the existence of a final newline.
// Line endings are assumed to already be consistent within the buffer.
// A nil buffer or an empty buffer is returned as is.
func fixInsertFinalNewline(buf *bytes.Buffer, insertFinalNewline bool, endOfLine []byte) bool {
	fixed := false

	if buf == nil || buf.Len() == 0 {
		return fixed
	}

	if insertFinalNewline {
		if !bytes.HasSuffix(buf.Bytes(), endOfLine) {
			fixed = true

			buf.Write(endOfLine)
		}
	} else {
		for bytes.HasSuffix(buf.Bytes(), endOfLine) {
			fixed = true
			buf.Truncate(buf.Len() - len(endOfLine))
		}
	}

	return fixed
}
