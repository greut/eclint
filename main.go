package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
)

var (
	version = "dev"
)

// option contains the environment of the program.
type option struct {
	isTerminal        bool
	noColors          bool
	showAllErrors     bool
	summary           bool
	showErrorQuantity int
	exclude           string
	log               logr.Logger
	stdout            io.Writer
}

func main() { //nolint:funlen
	flagVersion := false
	opt := option{
		stdout:            os.Stdout,
		showErrorQuantity: 10,
		log:               klogr.New(),
		isTerminal:        terminal.IsTerminal(syscall.Stdout),
	}

	// Flags
	klog.InitFlags(nil)
	flag.BoolVar(&flagVersion, "version", false, "print the version number")
	flag.BoolVar(&opt.noColors, "no_colors", false, "enable or disable colors")
	flag.BoolVar(&opt.summary, "summary", false, "enable the summary view")
	flag.BoolVar(
		&opt.showAllErrors,
		"show_all_errors",
		false,
		fmt.Sprintf("display all errors for each file (otherwise %d are kept)", opt.showErrorQuantity),
	)
	flag.StringVar(&opt.exclude, "exclude", "", "paths to exclude")
	flag.Parse()

	if flagVersion {
		fmt.Fprintf(opt.stdout, "eclint %s\n", version)
		return
	}

	args := flag.Args()
	files, err := listFiles(opt.log, args...)
	if err != nil {
		opt.log.Error(err, "error while handling the arguments")
		flag.Usage()
		os.Exit(1)
		return
	}

	opt.log.V(1).Info("files", "count", len(files), "exclude", opt.exclude)

	if opt.summary {
		opt.showAllErrors = true
		opt.showErrorQuantity = int(^uint(0) >> 1)
	}

	c := 0
	for _, filename := range files {
		// Skip excluded files
		if opt.exclude != "" {
			ok, err := editorconfig.FnmatchCase(opt.exclude, filename)
			if err != nil {
				opt.log.Error(err, "exclude pattern failure", "exclude", opt.exclude)
				c++
				break
			}
			if ok {
				continue
			}
		}

		c += lintAndPrint(opt, filename)
	}

	if c > 0 {
		opt.log.V(1).Info("Some errors were found.", "count", c)
		os.Exit(1)
	}
}
