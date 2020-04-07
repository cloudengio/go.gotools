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
	ConfigFileFlag string
	AnnotationFlag string
	ListFlag       bool
	VerboseFlag    bool
	ProgressFlag   bool
)

const defaultConfigFile = "$HOME/.goannoate/config.yaml"

func init() {
	flag.StringVar(&ConfigFileFlag, "config", os.ExpandEnv(defaultConfigFile), "yaml configuration file")
	flag.StringVar(&AnnotationFlag, "annotation", "", "annotation to be applied")
	flag.BoolVar(&ListFlag, "list", false, "list available annotations")
	flag.BoolVar(&VerboseFlag, "verbose", false, "display verbose debug info")
}

func handleDebug(ctx context.Context, cfg Debug) (func(), error) {
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
	annotators.Verbose = VerboseFlag

	config, err := ConfigFromFile(ConfigFileFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	if ListFlag {
		fmt.Println(available())
		return
	}

	cleanup, err := handleDebug(ctx, config.Debug)
	if err != nil {
		cmdutil.Exit("failed to configure debugging/profiling: %v\n", err)
	}
	defer cleanup()

	cmdutil.HandleSignals(cancel, os.Interrupt, os.Kill)
	cmdutil.HandleSignals(cleanup, os.Interrupt, os.Kill)

	if !flags.ExactlyOneSet(AnnotationFlag) {
		cmdutil.Exit("--annotation must be specified\n")
	}
	names := bySuffix(AnnotationFlag)
	switch len(names) {
	case 0:
		cmdutil.Exit("no annotator found for %v\n", AnnotationFlag)
	case 1:
	default:
		cmdutil.Exit("multiple annotators found for %v: %v\n", AnnotationFlag, strings.Join(names, ", "))
	}
	an := annotators.Lookup(names[0])
	if an == nil {
		cmdutil.Exit("unrecognised annotation: %v\n%v\n", AnnotationFlag, available())
	}
	if err := an.Do(ctx, config.Packages); err != nil {
		cmdutil.Exit("%v", err)
	}
}

func available() string {
	out := strings.Builder{}
	for _, name := range annotators.Available() {
		an := annotators.Lookup(name)
		out.WriteString(an.Describe())
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
