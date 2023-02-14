// Usage of gousage:
//
// gousage is a utility for generating a go source code file containing
// the usage information for command packages.
//
// gousage can be applied to multiple packages and will generate its output
// for each package in that package's directory.
//
// For example, the following will generate cmdusage.go files for every
// main package found in ./...
//
//	go run cloudeng.io/go/cmd/gousage --overwrite ./...
//
// Command line flags:
//
//	-go-output string
//	  	name of generated go file. (default "cmdusage.go")
//	-overwrite
//	  	overwrite existing file.
package main
