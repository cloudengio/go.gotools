# go.gotools

![linux](https://github.com/cloudengio/go.gotools/actions/workflows/linux.yml/badge.svg)
![macos](https://github.com/cloudengio/go.gotools/actions/workflows/macos.yml/badge.svg)
![windows](https://github.com/cloudengio/go.gotools/actions/workflows/windows.yml/badge.svg)
![CodeQL](https://github.com/cloudengio/go.gotools/actions/workflows/codeql.yml/badge.svg)

go tools and packages for manipulating go source code and
generating documentation.

`go.gotools` contains the following cmds and packages:

- [cloudeng.io/go/cmd/gomarkdown](cmd/gomarkdown/README.md): generate markdown for go packages and commands.
- [cloudeng.io/go/cmd/gousage](cmd/gousage/README.md): generate go code that
captures the output of --help as a package comment.
- [cloudeng.io/go/cmd/golocate](cmd/golocate/README.md): locating interface implementations, functions and comments.
- [cloudeng.io/go/cmd/goannotate](cmd/goannotate/README.md): extensible, configurable
annotations such as for copyright/licenses, extensible logging etc.
- [cloudeng.io/go/locate](locate/README.md): support for locating interface implementations,
functions etc.
- [cloudeng.io/go/derive](derive/README.md): support for deriving go code from existing
types and function signatures.

The following commands are used to maintain the package
and command README.md's for this repository and to ensure
that the appropriate copyright and license headers are present.

```go
go run cloudeng.io/go/cmd/gousage --overwrite ./...
go run cloudeng.io/go/cmd/goannotate --config=copyright-annotation.yaml --annotation=cloudeng-copyright ./...
go run cloudeng.io/go/cmd/gomarkdown --overwrite --circleci=cloudengio/go.gotools --goreportcard ./...
```
