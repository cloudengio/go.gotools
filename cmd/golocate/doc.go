package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const usage = `
golocate is a utility for locating interface implementations, functions
and comments in go source code using the parsed representation of the code
rather than simple text search.

Locate all instances of io.Writer ./...
  go run . --interfaces io.Writer ./...

Locate all exported functions in ./...
  go run . --functions='.*' ./...

Locate all comments in ./...
  go run . --comments='.*' ./...

The output of golocate is limited right now but is easily extended as
uses cases arise. Currently locating interface implementations is the
most useful.
`

func init() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage of %s:\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(out, "%s\n", usage)
		flag.PrintDefaults()
	}
}
