package eclint_test

import (
	"context"
	"fmt"
	"testing"

	"gitlab.com/greut/eclint"
)

func TestLintSimple(t *testing.T) {
	ctx := context.TODO()

	for _, err := range eclint.Lint(ctx, "testdata/simple/simple.txt") {
		if err != nil {
			t.Errorf("no errors where expected, got %s", err)
		}
	}
}

func TestLintMissing(t *testing.T) {
	ctx := context.TODO()

	errs := eclint.Lint(ctx, "testdata/missing/file")
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
	ctx := context.TODO()

	errs := eclint.Lint(ctx, "testdata/invalid/.editorconfig")
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
	ctx := context.TODO()

	for _, f := range []string{"a", "b"} {
		for _, err := range eclint.Lint(ctx, fmt.Sprintf("./testdata/block_comments/%s", f)) {
			if err != nil {
				t.Fatalf("no errors where expected, got %s", err)
			}
		}
	}
}

func TestBlockCommentInvalidSpec(t *testing.T) {
	ctx := context.TODO()

	for _, f := range []string{"c"} {
		errs := eclint.Lint(ctx, fmt.Sprintf("./testdata/block_comments/%s", f))
		if len(errs) == 0 {
			t.Errorf("one error was expected, got none")
		}
	}
}

func TestLintCharset(t *testing.T) {
	ctx := context.TODO()

	for _, f := range []string{"latin1", "utf8"} {
		for _, err := range eclint.Lint(ctx, fmt.Sprintf("./testdata/charset/%s.txt", f)) {
			if err != nil {
				t.Errorf("no errors where expected, got %s", err)
			}
		}
	}
}

func TestLintImages(t *testing.T) {
	ctx := context.TODO()

	for _, f := range []string{"edcon_tool.png", "edcon_tool.pdf", "hello.txt.gz"} {
		for _, err := range eclint.Lint(ctx, fmt.Sprintf("./testdata/images/%s", f)) {
			if err != nil {
				t.Fatalf("no errors where expected, got %s", err)
			}
		}
	}
}

func TestMaxLineLengthValidSpec(t *testing.T) {
	ctx := context.TODO()

	for _, f := range []string{"a", "b"} {
		for _, err := range eclint.Lint(ctx, fmt.Sprintf("./testdata/max_line_length/%s", f)) {
			if err != nil {
				t.Fatalf("no errors where expected, got %s", err)
			}
		}
	}
}

func TestMaxLineLengthInvalidSpec(t *testing.T) {
	ctx := context.TODO()

	for _, f := range []string{"c"} {
		errs := eclint.Lint(ctx, fmt.Sprintf("./testdata/max_line_length/%s", f))
		if len(errs) == 0 {
			t.Errorf("one error was expected, got none")
		}
	}
}

func TestInsertFinalNewlineSpec(t *testing.T) {
	ctx := context.TODO()

	for _, f := range []string{"with_final_newline.txt", "no_final_newline.md"} {
		for _, err := range eclint.Lint(ctx, fmt.Sprintf("./testdata/insert_final_newline/%s", f)) {
			if err != nil {
				t.Fatalf("no errors where expected, got %s", err)
			}
		}
	}
}

func TestInsertFinalNewlineInvalidSpec(t *testing.T) {
	ctx := context.TODO()

	for _, f := range []string{"no_final_newline.txt", "with_final_newline.md"} {
		errs := eclint.Lint(ctx, fmt.Sprintf("./testdata/insert_final_newline/%s", f))
		if len(errs) == 0 {
			t.Errorf("one error was expected, got none")
		}
	}
}
