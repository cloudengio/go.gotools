// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Usage of gomarkdown:
//
// gomarkdown is a utility for generating markdown from go source comments.
// It mirrors the behaviour of go doc but producing markdown output instead.
// It is most useful for packages hosted on github to generate a README.md
// for each package or command.
//
// gomarkdown can be applied to multiple packages and will generate a README.md
// for each package in that package's directory.
//
// For example, the following will generate a README.md in every package and
// command under ./...
//
//   go run cloudeng.io/go/cmd/gomarkdown --overwrite ./...
//
//   -circleci string
//     	set to the circleci project to insert a circleci build badge
//   -gopkg string
//     	link to this site for full godoc and godoc examples (default "pkg.go.dev")
//   -goreportcard
//     	insert a link to goreportcard.com
//   -markdown string
//     	markdown style to use, currently only github is supported. (default "github")
//   -md-output string
//     	name of markdown output file. (default "README.md")
//   -overwrite
//     	overwrite existing file.
package main
