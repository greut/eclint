package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"syscall"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/go-logr/logr"
	"github.com/mattn/go-colorable"
	"gitlab.com/greut/eclint"
	"golang.org/x/term"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
)

var version = "dev"

const (
	overridePrefix = "eclint_"
)

func main() { // nolint: funlen
	flagVersion := false
	color := "auto"
	cpuprofile := ""
	memprofile := ""

	// hack to ensure other deferrable are executed beforehand.
	retcode := 0

	defer func() { os.Exit(retcode) }()

	log := klogr.New()
	defer klog.Flush()

	opt := &eclint.Option{
		Stdout:            os.Stdout,
		ShowErrorQuantity: 10,
		IsTerminal:        term.IsTerminal(int(syscall.Stdout)), //nolint: unconvert
	}

	if runtime.GOOS == "windows" {
		opt.Stdout = colorable.NewColorableStdout()
	}

	// Flags
	klog.InitFlags(nil)
	flag.BoolVar(&flagVersion, "version", false, "print the version number")
	flag.StringVar(&color, "color", color, `use color when printing; can be "always", "auto", or "never"`)
	flag.BoolVar(&opt.Summary, "summary", opt.Summary, "enable the summary view")
	flag.BoolVar(&opt.FixAllErrors, "fix", opt.FixAllErrors, "enable fixing instead of error reporting")
	flag.BoolVar(
		&opt.ShowAllErrors,
		"show_all_errors",
		opt.ShowAllErrors,
		fmt.Sprintf("display all errors for each file (otherwise %d are kept)", opt.ShowErrorQuantity),
	)
	flag.IntVar(
		&opt.ShowErrorQuantity,
		"show_error_quantity",
		opt.ShowErrorQuantity,
		"display only the first n errors (0 means all)",
	)
	flag.StringVar(&opt.Exclude, "exclude", opt.Exclude, "paths to exclude")
	flag.StringVar(&cpuprofile, "cpuprofile", cpuprofile, "write cpu profile to `file`")
	flag.StringVar(&memprofile, "memprofile", memprofile, "write mem profile to `file`")
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

			retcode = 1

			return
		}

		defer f.Close()

		if err := pprof.StartCPUProfile(f); err != nil {
			log.Error(err, "could not start CPU profile")
		}
	}

	ctx := logr.NewContext(context.Background(), log)
	c, err := processArgs(ctx, opt, flag.Args())
	if err != nil {
		log.Error(err, "linting failure")

		retcode = 2

		return
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
		log.V(1).Info("some errors were found.", "count", c)

		retcode = 1
	}
}

func processArgs(ctx context.Context, opt *eclint.Option, args []string) (int, error) { // nolint:funlen
	log := logr.FromContextOrDiscard(ctx)
	c := 0

	config := &editorconfig.Config{
		Parser: editorconfig.NewCachedParser(),
	}

	fileChan, errChan := eclint.ListFilesContext(ctx, args...)

	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()

		case err, ok := <-errChan:
			if ok {
				log.Error(err, "cannot list files")

				return 0, err
			}

		case filename, ok := <-fileChan:
			if !ok {
				return c, nil
			}

			log := log.WithValues("filename", filename)

			// Skip excluded files
			if opt.Exclude != "" {
				ok, err := editorconfig.FnmatchCase(opt.Exclude, filename)
				if err != nil {
					log.Error(err, "exclude pattern failure", "exclude", opt.Exclude)

					return 0, err
				}

				if ok {
					continue
				}
			}

			def, err := config.Load(filename)
			if err != nil {
				log.Error(err, "cannot open file")

				return 0, err
			}

			err = eclint.OverrideDefinitionUsingPrefix(def, overridePrefix)
			if err != nil {
				log.Error(err, "overriding the definition failed", "prefix", overridePrefix)

				return 0, err
			}

			// Linting vs Fixing
			if !opt.FixAllErrors {
				errs := eclint.LintWithDefinition(ctx, def, filename)
				c += len(errs)

				if err := eclint.PrintErrors(ctx, opt, filename, errs); err != nil {
					log.Error(err, "print errors failure")

					return 0, err
				}
			} else {
				err := eclint.FixWithDefinition(ctx, def, filename)
				if err != nil {
					log.Error(err, "fixing errors failure")

					return 0, err
				}
			}
		}
	}
}
