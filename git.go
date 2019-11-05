package main

import (
	"os/exec"
	"strings"
)

func gitLsFiles() ([]string, error) {
	lines, err := exec.Command("git", "ls-files").Output()
	if err != nil {
		return nil, err
	}

	files := strings.Split(string(lines), "\n")
	return files[:len(files)-1], nil
}
