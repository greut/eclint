package eclint

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-logr/logr"
)

// ListFiles returns the list of files based on the input.
//
// When its empty, it relies on `git ls-files` first, which
// whould fail if `git` is not present or the current working
// directory is not managed by it. In that case, it work the
// current working directory.
//
// When args are given, it recursively walks into them.
func ListFiles(log logr.Logger, args ...string) ([]string, error) {
	if len(args) == 0 {
		fs, err := GitLsFiles(log, ".")
		if err == nil {
			return fs, nil
		}

		log.Error(err, "git ls-files failure")

		args = append(args, ".")
	}

	return Walk(log, args...)
}

// Walk iterates on each path item recursively.
func Walk(log logr.Logger, paths ...string) ([]string, error) {
	files := make([]string, 0)

	for _, path := range paths {
		err := filepath.Walk(path, func(p string, _ os.FileInfo, e error) error {
			if e != nil {
				return e
			}
			log.V(2).Info("index", "filename", p)
			files = append(files, p)

			return nil
		})
		if err != nil {
			return files, err
		}
	}

	return files, nil
}

// GitLsFiles returns the list of file based on what is in the git index.
//
// -z is mandatory as some repositories non-ASCII file names which creates
// quoted and escaped file names. This method also returns directories for
// any submodule there is. Submodule will be skipped afterwards and thus
// not checked.
func GitLsFiles(log logr.Logger, path string) ([]string, error) {
	output, err := exec.Command("git", "ls-files", "-z", path).Output()
	if err != nil {
		return nil, err
	}

	fs := bytes.Split(output, []byte{0})
	// last line is empty
	files := make([]string, len(fs)-1)
	for i := 0; i < len(files); i++ {
		p := string(fs[i])
		log.V(2).Info("index", "filename", p)
		files[i] = p
	}

	return files, nil
}