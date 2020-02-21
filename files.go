package eclint

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-logr/logr"
)

// ListFilesContext lists the files in an asynchronous fashion
//
// When its empty, it relies on `git ls-files` first, which
// whould fail if `git` is not present or the current working
// directory is not managed by it. In that case, it work the
// current working directory.
//
// When args are given, it recursively walks into them.
func ListFilesContext(ctx context.Context, log logr.Logger, args ...string) (<-chan string, <-chan error) {
	if len(args) > 0 {
		return WalkContext(ctx, args...)
	}

	return GitLsFilesContext(ctx, ".")
}

// ListFiles returns the list of files based on the input.
//
// Deprecated: use ListFilesContext
func ListFiles(log logr.Logger, args ...string) ([]string, error) {
	filesChan, errChan := ListFilesContext(context.Background(), log, args...)

	return ConsumeFilesContext(context.Background(), log, filesChan, errChan)
}

// WalkContext iterates on each path item recursively (asynchronously)
//
// Future work: use godirwalk
func WalkContext(ctx context.Context, paths ...string) (<-chan string, <-chan error) {
	filesChan := make(chan string, 128)
	errChan := make(chan error, 1)

	go func() {
		defer close(filesChan)
		defer close(errChan)

		for _, path := range paths {
			err := filepath.Walk(path, func(filename string, _ os.FileInfo, e error) error {
				if e != nil {
					return e
				}

				select {
				case filesChan <- filename:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			})

			if err != nil {
				errChan <- err
				break
			}
		}
	}()

	return filesChan, errChan
}

// Walk iterates on each path item recursively.
//
// Deprecated: use WalkContext
func Walk(log logr.Logger, paths ...string) ([]string, error) {
	filesChan, errChan := WalkContext(context.Background(), paths...)

	return ConsumeFilesContext(context.Background(), log, filesChan, errChan)
}

// AsyncGitLsFiles returns the list of file base on what is in the git index (asynchronously)
//
// -z is mandatory as some repositories non-ASCII file names which creates
// quoted and escaped file names. This method also returns directories for
// any submodule there is. Submodule will be skipped afterwards and thus
// not checked.
//
// Future work: use go-cmd for async call
func GitLsFilesContext(ctx context.Context, path string) (<-chan string, <-chan error) {
	filesChan := make(chan string, 128)
	errChan := make(chan error, 1)

	go func() {
		defer close(filesChan)
		defer close(errChan)

		output, err := exec.CommandContext(ctx, "git", "ls-files", "-z", path).Output()
		if err != nil {
			if e, ok := err.(*exec.ExitError); ok {
				if e.ExitCode() == 128 {
					err = fmt.Errorf("not a git repository")
				} else {
					err = fmt.Errorf("git ls-files failed with %s", e.Stderr)
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

// GitLsFiles returns the list of file based on what is in the git index.
func GitLsFiles(log logr.Logger, path string) ([]string, error) {
	filesChan, errChan := GitLsFilesContext(context.Background(), path)

	return ConsumeFilesContext(context.Background(), log, filesChan, errChan)
}

// ConsumeFilesContext is the helper function that consumes the files channel.
func ConsumeFilesContext(
	ctx context.Context,
	log logr.Logger,
	filesChan <-chan string,
	errChan <-chan error,
) ([]string, error) {
	files := make([]string, 0)

	for filesChan != nil {
		select {
		case filename, ok := <-filesChan:
			if ok {
				log.V(2).Info("index", "filename", filename)
				files = append(files, filename)
			} else {
				filesChan = nil
			}
		case err, ok := <-errChan:
			if ok {
				return nil, err
			}

			errChan = nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return files, nil
}
