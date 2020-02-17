package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/mattn/go-colorable"
	"gitlab.com/greut/eclint"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
)

var (
	version = "dev"
)

const (
	overridePrefix = "eclint_"
)

func main() { //nolint:funlen
	flagVersion := false
	color := "auto"
	cpuprofile := ""
	memprofile := ""
	log := klogr.New()
	opt := eclint.Option{
		Stdout:            os.Stdout,
		ShowErrorQuantity: 10,
		IsTerminal:        terminal.IsTerminal(int(syscall.Stdout)), //nolint: unconvert
		Log:               log,
	}

	if runtime.GOOS == "windows" {
		opt.Stdout = colorable.NewColorableStdout()
	}

	// Flags
	klog.InitFlags(nil)
	flag.BoolVar(&flagVersion, "version", false, "print the version number")
	flag.StringVar(&color, "color", color, `use color when printing; can be "always", "auto", or "never"`)
	flag.BoolVar(&opt.Summary, "summary", false, "enable the summary view")
	flag.BoolVar(
		&opt.ShowAllErrors,
		"show_all_errors",
		false,
		fmt.Sprintf("display all errors for each file (otherwise %d are kept)", opt.ShowErrorQuantity),
	)
	flag.IntVar(
		&opt.ShowErrorQuantity,
		"show_error_quantity",
		opt.ShowErrorQuantity,
		"display only the first n errors (0 means all)",
	)
	flag.StringVar(&opt.Exclude, "exclude", "", "paths to exclude")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to `file`")
	flag.StringVar(&memprofile, "memprofile", "", "write mem profile to `file`")
	flag.Parse()

	if flagVersion {
		fmt.Fprintf(opt.Stdout, "eclint %s\n", version)
		return
	}

	switch color {
	case "always":
		opt.IsTerminal = true
	case "never":
		opt.NoColors = true
	}

	if opt.Summary {
		opt.ShowAllErrors = true
	}

	if opt.ShowAllErrors {
		opt.ShowErrorQuantity = 0
	}

	if opt.Exclude != "" {
		_, err := editorconfig.FnmatchCase(opt.Exclude, "dummy")
		if err != nil {
			log.Error(err, "exclude pattern failure", "exclude", opt.Exclude)
			flag.Usage()
			return
		}
	}

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Error(err, "could not create CPU profile", "cpuprofile", cpuprofile)
			os.Exit(1)

			return
		}
		defer f.Close()

		if err := pprof.StartCPUProfile(f); err != nil {
			log.Error(err, "could not start CPU profile")
		}
	}

	args := flag.Args()

	config := &editorconfig.Config{
		Parser: editorconfig.NewCachedParser(),
	}

	c := 0

	fileChan, errChan := eclint.ListFilesContext(context.Background(), log, args...)
outter:
	for {
		select {
		case err, ok := <-errChan:
			if ok {
				log.Error(err, "cannot list files")
			}
			break outter

		case filename, ok := <-fileChan:
			if !ok {
				break outter
			}

			// Skip excluded files
			if opt.Exclude != "" {
				ok, err := editorconfig.FnmatchCase(opt.Exclude, filename)
				if err != nil {
					log.Error(err, "exclude pattern failure", "exclude", opt.Exclude)
					c++

					break
				}

				if ok {
					continue
				}
			}

			def, err := config.Load(filename)
			if err != nil {
				log.Error(err, "cannot open file", "filename", filename)
				c++

				break
			}

			err = eclint.OverrideDefinitionUsingPrefix(def, overridePrefix)
			if err != nil {
				log.Error(err, "overriding the definition failed", "prefix", overridePrefix)
				c++

				break
			}

			errs := eclint.LintWithDefinition(def, filename, opt.Log.WithValues("filename", filename))
			c += len(errs)

			if err := eclint.PrintErrors(opt, filename, errs); err != nil {
				log.Error(err, "print errors failure", "filename", filename)
			}
		}
	}

	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			log.Error(err, "could not create memory profile", "memprofile", memprofile)
		}
		defer f.Close()

		runtime.GC() // get up-to-date statistics

		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Error(err, "could not write memory profile")
		}
	}

	if cpuprofile != "" {
		pprof.StopCPUProfile()
	}

	if c > 0 {
		opt.Log.V(1).Info("Some errors were found.", "count", c)
		os.Exit(1)
	}
}
