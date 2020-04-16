// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const usage = `
gousage is a utility for generating a go source code file containing
the usage information for command packages.

gousage can be applied to multiple packages and will generate its output
for each package in that package's directory.

For example, the following will generate cmdusage.go files for every
main package found in ./...

  go run cloudeng.io/go/cmd/gousage --overwrite ./...
`

func init() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage of %s:\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(out, "%s\n", usage)
		flag.PrintDefaults()
	}
}
