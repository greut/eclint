package eclint_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	tlogr "github.com/go-logr/logr/testing"
	"gitlab.com/greut/eclint"
)

const (
	// testdataSimple contains a sample editorconfig directory with
	// some errors.
	testdataSimple = "testdata/simple"
)

func TestListFiles(t *testing.T) {
	l := tlogr.TestLogger{}
	d := testdataSimple

	fs := 0
	fsChan, errChan := eclint.ListFilesContext(context.TODO(), l, d)

outer:
	for {
		select {
		case err, ok := <-errChan:
			if ok && err != nil {
				t.Fatal(err)
			}

		case _, ok := <-fsChan:
			if !ok {
				break outer
			}
			fs++
		}
	}

	if fs != 5 {
		t.Errorf("%s should have five files, got %d", d, fs)
	}
}

func TestListFilesNoArgs(t *testing.T) {
	skipNoGit(t)

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

	fs := 0
	fsChan, errChan := eclint.ListFilesContext(context.TODO(), l)
outer:
	for {
		select {
		case err, ok := <-errChan:
			if ok && err != nil {
				t.Fatal(err)
			}
		case _, ok := <-fsChan:
			if !ok {
				break outer
			}
			fs++
		}
	}

	if fs != 3 {
		t.Errorf("%s should have three files, got %d", d, fs)
	}
}

func TestListFilesNoGit(t *testing.T) {
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

	_, errChan := eclint.ListFilesContext(context.TODO(), l)

	err, ok := <-errChan
	if !ok || err == nil {
		t.Errorf("an error was expected, got nothing")
	}

	fs := 0
	fsChan, errChan := eclint.ListFilesContext(context.TODO(), l, ".")

outer:
	for {
		select {
		case err, ok := <-errChan:
			if ok && err != nil {
				t.Fatal(err)
			}
		case _, ok := <-fsChan:
			if !ok {
				break outer
			}
			fs++
		}
	}

	if fs != 1 {
		t.Errorf("%s should have one file, got %d", d, fs)
	}
}

func TestWalk(t *testing.T) {
	d := testdataSimple

	fs := 0
	fsChan, errChan := eclint.WalkContext(context.TODO(), d)

outer:
	for {
		select {
		case err, ok := <-errChan:
			if ok && err != nil {
				t.Fatal(err)
			}
		case _, ok := <-fsChan:
			if !ok {
				break outer
			}
			fs++
		}
	}

	if fs != 5 {
		t.Errorf("%s should have five files, got %d", d, fs)
	}
}

func TestGitLsFiles(t *testing.T) {
	skipNoGit(t)

	l := tlogr.TestLogger{}
	d := testdataSimple

	fs, err := eclint.GitLsFiles(l, d)
	if err != nil {
		t.Fatal(err)
	}

	if len(fs) != 3 {
		t.Errorf("%s should have two files, got %d", d, len(fs))
	}
}

func TestGitLsFilesFailure(t *testing.T) {
	skipNoGit(t)

	l := tlogr.TestLogger{}
	d := fmt.Sprintf("/tmp/eclint/%d", os.Getpid())

	err := os.MkdirAll(d, 0700)
	if err != nil {
		t.Fatal(err)
	}

	_, err = eclint.GitLsFiles(l, d)
	if err == nil {
		t.Error("an error was expected")
	}
}

func skipNoGit(t *testing.T) {
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		t.Skip("skipping test requiring .git to be present")
	}
}
