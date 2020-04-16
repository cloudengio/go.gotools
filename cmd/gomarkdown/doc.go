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

For commands, gomarkdown can also generate a go file containing a comment
with the usage for that command, obtained by running it with --help.
With this generated file in place, gomarkdown can generate a useful markdown
file for the command.

gomarkdown can be applied to multiple packages and will generate a README.md
for each package in that package's directory.
`

func init() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage of %s:\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(out, "%s\n", usage)
		flag.PrintDefaults()
	}
}
