// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/pprof"
	"strings"

	"cloudeng.io/cmdutil"
	"cloudeng.io/cmdutil/flags"
	"cloudeng.io/go/cmd/goannotate/annotators"
)

var (
	configFileFlag string
	annotationFlag string
	writeDirFlag   string
	listFlag       bool
	listConfigFlag bool
	verboseFlag    bool
)

const defaultConfigFile = "config.yaml"

func init() {
	flag.StringVar(&configFileFlag, "config", os.ExpandEnv(defaultConfigFile), "yaml configuration file")
	flag.StringVar(&annotationFlag, "annotation", "", "annotation to be applied")
	flag.StringVar(&writeDirFlag, "write-dir", "", "if set, specify an alternate directory to write modified files to, otherwise files are modified in place.")
	flag.BoolVar(&listFlag, "list", false, "list available annotators")
	flag.BoolVar(&listConfigFlag, "list-config", false, "list available annotations and their configurations")
	flag.BoolVar(&verboseFlag, "verbose", false, "display verbose debug info")
}

func handleDebug(_ context.Context, cfg debug) (func(), error) {
	var cpu io.WriteCloser
	if filename := os.ExpandEnv(cfg.CPUProfile); len(filename) > 0 {
		var err error
		cpu, err = os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
		if err != nil {
			return func() {}, err
		}
		if err := pprof.StartCPUProfile(cpu); err != nil {
			cpu.Close()
			return func() {}, err
		}
		fmt.Printf("writing cpu profile to: %v\n", filename)
	}
	return func() {
		pprof.StopCPUProfile()
		if cpu != nil {
			cpu.Close()
		}
	}, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	flag.Parse()
	annotators.Verbose = verboseFlag

	if listFlag {
		fmt.Println(describe(annotators.Registered()))
		return
	}

	config, err := configFromFile(configFileFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	if listConfigFlag {
		fmt.Println(describe(annotators.Available()))
		return
	}

	cleanup, err := handleDebug(ctx, config.Debug)
	if err != nil {
		cmdutil.Exit("failed to configure debugging/profiling: %v\n", err)
	}
	defer cleanup()

	cmdutil.HandleSignals(cancel, os.Interrupt, os.Kill)
	cmdutil.HandleSignals(cleanup, os.Interrupt, os.Kill)

	if !flags.ExactlyOneSet(annotationFlag) {
		cmdutil.Exit("--annotation must be specified\n")
	}
	names := bySuffix(annotationFlag)
	switch len(names) {
	case 0:
		cmdutil.Exit("no annotator found for %v\n", annotationFlag)
	case 1:
	default:
		cmdutil.Exit("multiple annotators found for %v: %v\n", annotationFlag, strings.Join(names, ", "))
	}
	an := annotators.Lookup(names[0])
	if an == nil {
		cmdutil.Exit("unrecognised annotation: %v\n%v\n", annotationFlag, describe(annotators.Available()))
	}
	if err := an.Do(ctx, writeDirFlag, flag.Args()); err != nil {
		cmdutil.Exit("%v", err)
	}
}

func describe(names []string) string {
	out := strings.Builder{}
	for _, name := range names {
		out.WriteString(annotators.Description(name))
		out.WriteString("\n")
	}
	return out.String()
}

func bySuffix(suffix string) []string {
	found := []string{}
	suffix = strings.ToLower(suffix)
	for _, name := range annotators.Available() {
		lname := strings.ToLower(name)
		if strings.HasSuffix(lname, suffix) {
			found = append(found, name)
		}
	}
	return found
}
