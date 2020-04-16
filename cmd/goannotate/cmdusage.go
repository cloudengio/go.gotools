// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Usage of goannotate:
// goannotate provides a configurable and extensible set of annotators
// that can be used to add/remove statements from large bodies of go source code.
//
//   -annotation string
//     	annotation to be applied
//   -config string
//     	yaml configuration file (default "config.yaml")
//   -list
//     	list available annotators
//   -list-config
//     	list available annotations and their configurations
//   -verbose
//     	display verbose debug info
//   -write-dir string
//     	if set, specify an alternate directory to write modified files to, otherwise files are modified in place.
package main
