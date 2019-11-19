package main

import (
	"fmt"
	"os"
	"testing"

	tlogr "github.com/go-logr/logr/testing"
)

const (
	// testdataSimple contains a sample editorconfig directory with
	// some errors.
	testdataSimple = "testdata/simple"
)

func TestListFiles(t *testing.T) {
	l := tlogr.TestLogger{}
	d := testdataSimple
	fs, err := listFiles(l, d)
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) != 2 {
		t.Errorf("%s should have two files, got %d", d, len(fs))
	}
}

func TestListFilesNoArgs(t *testing.T) {
	l := tlogr.TestLogger{}
	d := testdataSimple

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()

	err = os.Chdir(d)
	if err != nil {
		t.Fatal(err)
	}

	fs, err := listFiles(l)
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) != 2 {
		t.Errorf("%s should have two files, got %d", d, len(fs))
	}
}

func TestListFilesNoGit(t *testing.T) {
	// FIXME... should be the null logger, right?
	l := tlogr.NullLogger{}
	d := fmt.Sprintf("/tmp/eclint/%d", os.Getpid())

	err := os.MkdirAll(d, 0700)
	if err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()
	err = os.Chdir(d)
	if err != nil {
		t.Fatal(err)
	}

	fs, err := listFiles(l)
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) != 0 {
		t.Errorf("%s should have two files, got %d", d, len(fs))
	}
}

func TestWalk(t *testing.T) {
	l := tlogr.TestLogger{}
	d := testdataSimple
	fs, err := walk(l, d)
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) != 2 {
		t.Errorf("%s should have two files, got %d", d, len(fs))
	}
}

func TestGitLsFiles(t *testing.T) {
	l := tlogr.TestLogger{}
	d := testdataSimple

	fs, err := gitLsFiles(l, d)
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) != 2 {
		t.Errorf("%s should have two files, got %d", d, len(fs))
	}
}

func TestGitLsFilesFailure(t *testing.T) {
	l := tlogr.TestLogger{}
	d := fmt.Sprintf("/tmp/eclint/%d", os.Getpid())

	err := os.MkdirAll(d, 0700)
	if err != nil {
		t.Fatal(err)
	}

	_, err = gitLsFiles(l, d)
	if err == nil {
		t.Error("an error was expected")
	}
}
