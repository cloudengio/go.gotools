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
gomarkdown is a utility for generating markdown from go source comments.
It mirrors the behaviour of go doc but producing markdown output instead.
It is most useful for packages hosted on github to generate a README.md
for each package or command.

gomarkdown can be applied to multiple packages and will generate a README.md
for each package in that package's directory.

For example, the following will generate a README.md in every package and
command under ./...

  go run cloudeng.io/go/cmd/gomarkdown --overwrite ./...
`

func init() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage of %s:\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(out, "%s\n", usage)
		flag.PrintDefaults()
	}
}
