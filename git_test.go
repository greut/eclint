package main

import (
	"fmt"
	"os"
	"testing"
)

func TestGitLsFiles(t *testing.T) {
	d := "testdata/simple"
	fs, err := gitLsFiles(d)
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) != 2 {
		t.Errorf("%s should have two files, got %d", d, len(fs))
	}
}

func TestGitLsFilesFailure(t *testing.T) {
	d := fmt.Sprintf("/tmp/eclint/%d", os.Getpid())
	err := os.MkdirAll(d, 0700)
	if err != nil {
		t.Fatal(err)
	}

	_, err = gitLsFiles(d)
	if err == nil {
		t.Error("an error was expected")
	}
}
