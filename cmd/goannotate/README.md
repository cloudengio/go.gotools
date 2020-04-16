# [cloudeng.io/go/cmd/goannotate](https://pkg.go.dev/cloudeng.io/go/cmd/goannotate?tab=doc)
[![CircleCI](https://circleci.com/gh/cloudengio/go.gotools.svg?style=svg)](https://circleci.com/gh/cloudengio/go.gotools) [![Go Report Card](https://goreportcard.com/badge/cloudeng.io/go/cmd/goannotate)](https://goreportcard.com/report/cloudeng.io/go/cmd/goannotate)


Usage of `goannotate`: `goannotate` provides a configurable and extensible set
of annotators that can be used to add/remove statements from large bodies of
go source code.

# Command line flags

    -annotation string
      	annotation to be applied
    -config string
      	yaml configuration file (default "config.yaml")
    -list
      	list available annotators
    -list-config
      	list available annotations and their configurations
    -verbose
      	display verbose debug info
    -write-dir string
      	if set, specify an alternate directory to write modified files to, otherwise files are modified in place.

