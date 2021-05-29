package eclint

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/go-logr/logr"
	"github.com/karrick/godirwalk"
)

// ListFilesContext lists the files in an asynchronous fashion
//
// When its empty, it relies on `git ls-files` first, which
// whould fail if `git` is not present or the current working
// directory is not managed by it. In that case, it work the
// current working directory.
//
// When args are given, it recursively walks into them.
func ListFilesContext(ctx context.Context, args ...string) (<-chan string, <-chan error) {
	if len(args) > 0 {
		return WalkContext(ctx, args...)
	}

	dir := "."

	log := logr.FromContextOrDiscard(ctx)

	log.V(3).Info("fallback to `git ls-files`", "dir", dir)

	return GitLsFilesContext(ctx, dir)
}

// WalkContext iterates on each path item recursively (asynchronously).
//
// Future work: use godirwalk.
func WalkContext(ctx context.Context, paths ...string) (<-chan string, <-chan error) {
	filesChan := make(chan string, 128)
	errChan := make(chan error, 1)

	go func() {
		defer close(filesChan)
		defer close(errChan)

		for _, path := range paths {
			err := godirwalk.Walk(path, &godirwalk.Options{
				Callback: func(filename string, de *godirwalk.Dirent) error {
					select {
					case filesChan <- filename:
						return nil
					case <-ctx.Done():
						return fmt.Errorf("walking dir got interrupted: %w", ctx.Err())
					}
				},
				Unsorted: true,
			})
			if err != nil {
				errChan <- err

				break
			}
		}
	}()

	return filesChan, errChan
}

// GitLsFilesContext returns the list of file base on what is in the git index (asynchronously).
//
// -z is mandatory as some repositories non-ASCII file names which creates
// quoted and escaped file names. This method also returns directories for
// any submodule there is. Submodule will be skipped afterwards and thus
// not checked.
func GitLsFilesContext(ctx context.Context, path string) (<-chan string, <-chan error) {
	filesChan := make(chan string, 128)
	errChan := make(chan error, 1)

	go func() {
		defer close(filesChan)
		defer close(errChan)

		output, err := exec.CommandContext(ctx, "git", "ls-files", "-z", path).Output()
		if err != nil {
			var e *exec.ExitError
			if ok := errors.As(err, &e); ok {
				if e.ExitCode() == 128 {
					err = fmt.Errorf("not a git repository: %w", e)
				} else {
					err = fmt.Errorf("git ls-files failed with %s: %w", e.Stderr, e)
				}
			}

			errChan <- err

			return
		}

		fs := bytes.Split(output, []byte{0})
		// last line is empty
		for _, f := range fs[:len(fs)-1] {
			select {
			case filesChan <- string(f):
				// everything is good
			case <-ctx.Done():
				return
			}
		}
	}()

	return filesChan, errChan
}
