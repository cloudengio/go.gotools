# [main](https://pkg.go.dev/cloudeng.io/go/cmd/golocate?tab=doc)

# Command cloudeng.io/go/cmd/golocate

Usage of golocate:

golocate is a utility for locating interface implementations, functions and
comments in go source code using the parsed representation of the code
rather than simple text search.

Locate all instances of io.Writer ./...

    go run . --interfaces io.Writer ./...

Locate all exported functions in ./...

    go run . --functions='.*' ./...

Locate all comments in ./...

    go run . --comments='.*' ./...

The output of golocate is limited right now but is easily extended as uses
cases arise. Currently locating interface implementations is the most
useful.

    -comments string
      	if set, find all comments that match this regular expression in the specified packages.
    -functions string
      	if set, find all functions whose name matches this regular expression.
    -interfaces string
      	if set, find all implementations of these interfaces in the speficied packages. The package local component of the interface name is treated as a regular expression

