package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/editorconfig/editorconfig-core-go/v2"
)

var (
	version = "dev"
)

func walk(paths ...string) ([]string, error) {
	files := make([]string, 0)
	for _, path := range paths {
		log.Printf("enter %s\n", path)
		err := filepath.Walk(path, func(p string, i os.FileInfo, e error) error {
			mode := i.Mode()
			if mode.IsRegular() && !mode.IsDir() {
				log.Printf("add %s\n", p)
				abs, err := filepath.Abs(p)
				if err != nil {
					return err
				}
				files = append(files, abs)
			}
			return nil
		})
		if err != nil {
			return files, err
		}
	}
	return files, nil
}

func lint(filename string) error {
	// XXX editorconfig should be able to treat a flux of
	// filenames with caching capabilities.
	def, err := editorconfig.GetDefinitionForFilename(filename)
	if err != nil {
		return fmt.Errorf("Cannot open file %s. %w", filename, err)
	}
	log.Printf("lint %s %#v", filename, def)

	fp, err := os.Open(filename)
	if err != nil {
		return err
	}

	err = readLines(fp, func(index int, data []byte) error {
		log.Printf("line %d: %s", index, data)
		if def.EndOfLine != "" {
			err := endOfLine(def.EndOfLine, data)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

func main() {
	var flagVersion bool

	flag.BoolVar(&flagVersion, "version", false, "print the version number")
	flag.Parse()

	if flagVersion {
		fmt.Printf("ec %s\n", version)
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		args = append(args, ".")
	}

	files, err := walk(args...)
	if err != nil {
		log.Printf("[ERROR] %v", err)
		os.Exit(1)
		return
	}
	log.Printf("%d files found", len(files))

	c := 0
	for _, file := range files {
		err := lint(file)
		if err != nil {
			log.Printf("[ERROR] %v", err)
			c++
		}
	}
	if c > 0 {
		os.Exit(1)
	}
}
