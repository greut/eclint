package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
)

var (
	version = "dev"
	log     logr.Logger
)

func walk(paths ...string) ([]string, error) {
	files := make([]string, 0)
	for _, path := range paths {
		err := filepath.Walk(path, func(p string, i os.FileInfo, e error) error {
			if e != nil {
				return e
			}
			mode := i.Mode()
			if mode.IsRegular() && !mode.IsDir() {
				log.V(4).Info("index %s\n", p)
				_, err := filepath.Abs(p)
				if err != nil {
					return err
				}
				files = append(files, p)
			}
			return nil
		})
		if err != nil {
			return files, err
		}
	}
	return files, nil
}

func main() {
	var flagVersion bool

	exclude := ""

	klog.InitFlags(nil)
	flag.BoolVar(&flagVersion, "version", false, "print the version number")
	flag.StringVar(&exclude, "exclude", "", "paths to exclude")
	flag.Parse()

	if flagVersion {
		fmt.Printf("eclint %s\n", version)
		return
	}

	log = klogr.New()

	args := flag.Args()
	var files []string
	if len(args) == 0 {
		fs, err := gitLsFiles(".")
		if err != nil {
			args = append(args, ".")

			fs, err := walk(args...)
			if err != nil {
				log.Error(err, "error while handling the arguments")
				flag.Usage()
				os.Exit(1)
				return
			}

			files = fs
		} else {
			files = fs
		}
	}

	log.V(1).Info("files", "count", len(files), "exclude", exclude)

	c := 0
	for _, filename := range files {
		// Skip excluded files
		if exclude != "" {
			ok, err := editorconfig.FnmatchCase(exclude, filename)
			if err != nil {
				log.Error(err, "exclude pattern failure", "exclude", exclude)
				fmt.Printf("exclude pattern failure %s", err)
				c++
				break
			}
			if ok {
				continue
			}
		}

		d := 0
		errs := lint(filename, log)
		for _, err := range errs {
			if err != nil {
				log.V(4).Info("lint error", "filename", filename, "error", err)
				if d == 0 {
					fmt.Printf("%s:\n", filename)
				}
				fmt.Printf("\t%s\n", err)
				d++
				c++
			}
		}
	}
	if c > 0 {
		log.V(1).Info("Some errors were found.", "count", c)
		os.Exit(1)
	}
}
