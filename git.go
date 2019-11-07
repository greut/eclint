package main

import (
	"bytes"
	"os/exec"
)

// gitLsFiles returns the list of file based on what is in the git index.
//
// -z is mandatory as some repositories non-ASCII file names which creates
// quoted and escaped file names.
func gitLsFiles(path string) ([]string, error) {
	output, err := exec.Command("git", "ls-files", "-z", path).Output()
	if err != nil {
		return nil, err
	}

	fs := bytes.Split(output, []byte{0})
	// last line is empty
	files := make([]string, len(fs)-1)
	for i := 0; i < len(files); i++ {
		files[i] = string(fs[i])
	}
	return files, nil
}
