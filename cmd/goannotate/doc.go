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

const usage = `goannotate provides a configurable and extensible set of annotators
that can be used to add/remove statements from large bodies of go source code.
`

func init() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage of %s:\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(out, "%s\nCommand line flags:\n", usage)
		flag.PrintDefaults()
	}
}
