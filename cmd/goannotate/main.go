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
	WriteDirFlag   string
	ListFlag       bool
	ListConfigFlag bool
	VerboseFlag    bool
)

const defaultConfigFile = "config.yaml"

func init() {
	flag.StringVar(&ConfigFileFlag, "config", os.ExpandEnv(defaultConfigFile), "yaml configuration file")
	flag.StringVar(&AnnotationFlag, "annotation", "", "annotation to be applied")
	flag.StringVar(&WriteDirFlag, "write-dir", "", "if set, specify an alternate directory to write modified files to, otherwise files are modified in place.")
	flag.BoolVar(&ListFlag, "list", false, "list available annotators")
	flag.BoolVar(&ListConfigFlag, "list-config", false, "list available annotations and their configurations")
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

	if ListFlag {
		fmt.Println(describe(annotators.Registered()))
		return
	}

	config, err := ConfigFromFile(ConfigFileFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	if ListConfigFlag {
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
		cmdutil.Exit("unrecognised annotation: %v\n%v\n", AnnotationFlag, describe(annotators.Available()))
	}
	if err := an.Do(ctx, WriteDirFlag, flag.Args()); err != nil {
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
