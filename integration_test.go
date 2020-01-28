package eclint_test

import (
	"fmt"
	"testing"

	tlogr "github.com/go-logr/logr/testing"
	"gitlab.com/greut/eclint"
)

func TestLintSimple(t *testing.T) {
	l := tlogr.TestLogger{}

	for _, err := range eclint.Lint("testdata/simple/simple.txt", l) {
		if err != nil {
			t.Errorf("no errors where expected, got %s", err)
		}
	}
}

func TestLintMissing(t *testing.T) {
	l := tlogr.TestLogger{}

	errs := eclint.Lint("testdata/missing/file", l)
	if len(errs) == 0 {
		t.Error("an error was expected, got none")
	}

	for _, err := range errs {
		if err == nil {
			t.Error("an error was expected")
		}
	}
}

func TestLintInvalid(t *testing.T) {
	l := tlogr.TestLogger{}

	errs := eclint.Lint("testdata/invalid/.editorconfig", l)
	if len(errs) == 0 {
		t.Error("an error was expected, got none")
	}

	for _, err := range errs {
		if err == nil {
			t.Error("an error was expected")
		}
	}
}

func TestBlockCommentValidSpec(t *testing.T) {
	l := tlogr.TestLogger{}

	for _, f := range []string{"a", "b"} {
		for _, err := range eclint.Lint(fmt.Sprintf("./testdata/block_comments/%s", f), l) {
			if err != nil {
				t.Fatalf("no errors where expected, got %s", err)
			}
		}
	}
}

func TestBlockCommentInvalidSpec(t *testing.T) {
	l := tlogr.TestLogger{}

	for _, f := range []string{"c"} {
		errs := eclint.Lint(fmt.Sprintf("./testdata/block_comments/%s", f), l)
		if len(errs) == 0 {
			t.Errorf("one error was expected, got none")
		}
	}
}

func TestLintCharset(t *testing.T) {
	l := tlogr.TestLogger{}

	for _, f := range []string{"latin1", "utf8"} {
		for _, err := range eclint.Lint(fmt.Sprintf("./testdata/charset/%s.txt", f), l) {
			if err != nil {
				t.Errorf("no errors where expected, got %s", err)
			}
		}
	}
}

func TestLintImages(t *testing.T) {
	l := tlogr.TestLogger{}

	for _, f := range []string{"edcon_tool.png", "edcon_tool.pdf", "hello.txt.gz"} {
		for _, err := range eclint.Lint(fmt.Sprintf("./testdata/images/%s", f), l) {
			if err != nil {
				t.Fatalf("no errors where expected, got %s", err)
			}
		}
	}
}

func TestMaxLineLengthValidSpec(t *testing.T) {
	l := tlogr.TestLogger{}

	for _, f := range []string{"a", "b"} {
		for _, err := range eclint.Lint(fmt.Sprintf("./testdata/max_line_length/%s", f), l) {
			if err != nil {
				t.Fatalf("no errors where expected, got %s", err)
			}
		}
	}
}

func TestMaxLineLengthInvalidSpec(t *testing.T) {
	l := tlogr.TestLogger{}

	for _, f := range []string{"c"} {
		errs := eclint.Lint(fmt.Sprintf("./testdata/max_line_length/%s", f), l)
		if len(errs) == 0 {
			t.Errorf("one error was expected, got none")
		}
	}
}

func TestInsertFinalNewlineSpec(t *testing.T) {
	l := tlogr.TestLogger{}

	for _, f := range []string{"with_final_newline.txt", "no_final_newline.md"} {
		for _, err := range eclint.Lint(fmt.Sprintf("./testdata/insert_final_newline/%s", f), l) {
			if err != nil {
				t.Fatalf("no errors where expected, got %s", err)
			}
		}
	}
}

func TestInsertFinalNewlineInvalidSpec(t *testing.T) {
	l := tlogr.TestLogger{}

	for _, f := range []string{"no_final_newline.txt", "with_final_newline.md"} {
		errs := eclint.Lint(fmt.Sprintf("./testdata/insert_final_newline/%s", f), l)
		if len(errs) == 0 {
			t.Errorf("one error was expected, got none")
		}
	}
}
