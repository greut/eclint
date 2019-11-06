package main

import (
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
