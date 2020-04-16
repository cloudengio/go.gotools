# [cloudeng.io/go/cmd/gousage](https://pkg.go.dev/cloudeng.io/go/cmd/gousage?tab=doc)
[![CircleCI](https://circleci.com/gh/cloudengio/go.gotools.svg?style=svg)](https://circleci.com/gh/cloudengio/go.gotools) [![Go Report Card](https://goreportcard.com/badge/cloudeng.io/go/cmd/gousage)](https://goreportcard.com/report/cloudeng.io/go/cmd/gousage)


# Usage of `gousage`

`gousage` is a utility for generating a go source code file containing the
usage information for command packages.

`gousage` can be applied to multiple packages and will generate its output for
each package in that package's directory.

For example, the following will generate cmdusage.go files for every main
package found in ./...

    go run cloudeng.io/go/cmd/`gousage` --overwrite ./...

# Command line flags

    -go-output string
      	name of generated go file. (default "cmdusage.go")
    -overwrite
      	overwrite existing file.

