// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/doc"
	"go/doc/comment"
	"go/format"
	"go/token"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/packages"
)

// any line with fewer than the requested number of words and which ends in
// a colon is deemed to be a heading.
func guessHeadings(level, words int, text string) string {
	out := strings.Builder{}
	sc := bufio.NewScanner(bytes.NewBufferString(text))
	heading := strings.Repeat("#", level) + " "
	for sc.Scan() {
		line := sc.Text()
		if len(line) == 0 {
			out.WriteString("\n")
			continue
		}
		first, _ := utf8.DecodeRuneInString(line)
		fields := strings.Fields(line)
		l := len(fields)
		if unicode.IsUpper(first) && l > 0 && l < words && strings.HasSuffix(fields[l-1], ":") {
			out.WriteString(heading)
			fields[l-1] = strings.TrimSuffix(fields[l-1], ":")
			out.WriteString(strings.Join(fields, " "))
			out.WriteString("\n")
			continue
		}
		out.WriteString(line)
		out.WriteString("\n")
	}
	return out.String()
}

func highlight(word, text string) string {
	out := strings.Builder{}
	sc := bufio.NewScanner(bytes.NewBufferString(text))
	pre := regexp.MustCompile(`(\s+)` + word)
	post := regexp.MustCompile(word + `(\s+)`)
	quoted := "`" + word + "`"
	for sc.Scan() {
		line := pre.ReplaceAllString(sc.Text(), "$1"+quoted)
		line = post.ReplaceAllString(line, quoted+"$1")
		out.WriteString(line)
		out.WriteString("\n")
	}
	return out.String()
}

func goComment(text string) string {
	out := strings.Builder{}
	sc := bufio.NewScanner(bytes.NewBufferString(text))
	for sc.Scan() {
		out.WriteString("// ")
		out.WriteString(sc.Text())
		out.WriteString("\n")
	}
	return out.String()
}

func filterGoGenerate(text string) string {
	out := strings.Builder{}
	sc := bufio.NewScanner(bytes.NewBufferString(text))
	for sc.Scan() {
		if strings.HasPrefix(sc.Text(), "go:generate") {
			continue
		}
		out.WriteString(sc.Text())
		out.WriteString("\n")
	}
	return out.String()
}

type outputState struct {
	markdownGoCodeStart string
	doc                 *doc.Package
	pkg                 *packages.Package
	options             outputOptions
	tplFuncs            map[string]interface{}
}

type outputOptions struct {
	markdownFlavour string
	goPkgSite       string
	circleciProject string
	goreportcard    bool
}

type outputOption func(o *outputOptions)

func markdownFlavour(f string) outputOption {
	return func(o *outputOptions) {
		o.markdownFlavour = f
	}
}

func goPkgSite(s string) outputOption {
	return func(o *outputOptions) {
		o.goPkgSite = s
	}
}

func goreportcard(b bool) outputOption {
	return func(o *outputOptions) {
		o.goreportcard = b
	}
}

func circleciProject(p string) outputOption {
	return func(o *outputOptions) {
		o.circleciProject = p
	}
}

func newOutputState(doc *doc.Package, pkg *packages.Package, opts ...outputOption) *outputState {
	st := &outputState{doc: doc, pkg: pkg}
	for _, fn := range opts {
		fn(&st.options)
	}
	switch st.options.markdownFlavour {
	case "github":
		st.markdownGoCodeStart = "```go"
	default:
		st.markdownGoCodeStart = "```"
	}
	st.tplFuncs = map[string]interface{}{
		"badges":           st.badges,
		"codeStart":        st.codeStart,
		"codeEnd":          st.codeEnd,
		"comment":          st.comment,
		"filterGoGenerate": filterGoGenerate,
		"func":             st.funcDecl,
		"gocomment":        goComment,
		"guessHeadings":    guessHeadings,
		"highlight":        highlight,
		"type":             st.typeDecl,
		"value":            st.valueDecl,
		"join":             strings.Join,
	}
	switch st.options.goPkgSite {
	case "pkg.go.dev":
		st.tplFuncs["packageLink"] = st.pkgGoDevPackageLink
		st.tplFuncs["exampleLink"] = st.pkgGoDevExampleLink
	case "godoc.org":
		st.tplFuncs["packageLink"] = st.godocOrgPackageLink
		st.tplFuncs["exampleLink"] = st.godocOrgExampleLink
	default:
		panic(fmt.Sprintf("unsupported go pkg site: %v", st.options.goPkgSite))
	}
	return st
}

func (st *outputState) badges() string {
	var badges []string
	if ci := st.options.circleciProject; len(ci) > 0 {
		badge := fmt.Sprintf("[![CircleCI](https://circleci.com/gh/%v.svg?style=svg)](https://circleci.com/gh/%v)", ci, ci)
		badges = append(badges, badge)
	}
	if st.options.goreportcard {
		badge := fmt.Sprintf("[![Go Report Card](https://goreportcard.com/badge/%s)](https://goreportcard.com/report/%s)", st.pkg.PkgPath, st.pkg.PkgPath)
		badges = append(badges, badge)
	}
	return strings.Join(badges, " ")
}

func (st *outputState) valueDecl(decl *ast.GenDecl) string {
	switch decl.Tok {
	case token.CONST, token.VAR:
		out := &strings.Builder{}
		for _, spec := range decl.Specs {
			if err := format.Node(out, st.pkg.Fset, spec); err != nil {
				fmt.Fprintf(os.Stderr, "%v: failed to format const or var node: %v", st.pkg.PkgPath, err)
			}
			out.WriteString("\n")
		}
		return out.String()
	}
	return "unsupported"
}

func (st *outputState) funcDecl(decl *ast.FuncDecl) string {
	out := &strings.Builder{}
	if err := format.Node(out, st.pkg.Fset, decl); err != nil {
		at := st.pkg.Fset.PositionFor(decl.Pos(), false)
		fmt.Fprintf(os.Stderr, "%v: failed to format function declaration at %v: %v", st.pkg.PkgPath, at, err)
	}
	return out.String()
}

func (st *outputState) typeDecl(decl *ast.GenDecl) string {
	out := &strings.Builder{}
	if err := format.Node(out, st.pkg.Fset, decl); err != nil {
		at := st.pkg.Fset.PositionFor(decl.Pos(), false)
		fmt.Fprintf(os.Stderr, "%v: failed to format type declaration at %v: %v", st.pkg.PkgPath, at, err)
	}
	return out.String()
}

func (st *outputState) codeStart() string {
	return st.markdownGoCodeStart
}

func (st *outputState) codeEnd() string {
	return "```"
}

func (st *outputState) godocOrgPackageLink() string {
	return fmt.Sprintf("[%s](https://godoc.org/%s)", st.pkg.PkgPath, st.doc.ImportPath)
}

func (st *outputState) godocOrgExampleLink(eg *doc.Example) string {
	return fmt.Sprintf("[Example%s](https://godoc.org/%s#example-%s)",
		eg.Name, st.doc.ImportPath, eg.Name)
}

func (st *outputState) pkgGoDevPackageLink() string {
	return fmt.Sprintf("[%s](https://pkg.go.dev/%s?tab=doc)",
		st.pkg.PkgPath, st.doc.ImportPath)
}

func (st *outputState) pkgGoDevExampleLink(eg *doc.Example) string {
	return fmt.Sprintf("[Example%s](https://pkg.go.dev/%s?tab=doc#example-%s)",
		eg.Name, st.doc.ImportPath, eg.Name)
}

func (st *outputState) comment(indent, preIndent int, text string) string {
	out := &strings.Builder{}
	d := new(doc.Package).Parser().Parse(text)
	pr := &comment.Printer{
		TextPrefix:     strings.Repeat(" ", indent),
		TextCodePrefix: strings.Repeat(" ", preIndent),
		TextWidth:      80 - indent - preIndent,
	}
	out.Write(pr.Text(d))
	return out.String()
}

func (st *outputState) outputPackage() (string, error) {
	tpl, err := template.New("markdown").Funcs(st.tplFuncs).Parse(markdownPackageTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to create template: %v", err)
	}
	out := &strings.Builder{}
	err = tpl.Execute(out, st.doc)
	return out.String(), err
}

func (st *outputState) outputCommand() (string, error) {
	tpl, err := template.New("markdown").Funcs(st.tplFuncs).Parse(markdownCommandTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to create template: %v", err)
	}
	out := &strings.Builder{}
	tmp := struct {
		*doc.Package
		CommandName string
	}{
		st.doc,
		path.Base(st.pkg.PkgPath),
	}
	err = tpl.Execute(out, tmp)
	return out.String(), err
}

var markdownPackageTemplate = `# Package {{packageLink}}
{{if badges}}{{badges}}
{{end}}
{{codeStart}}
import {{.ImportPath}}
{{codeEnd}}

{{filterGoGenerate .Doc | comment 0 4}}

{{- if .Consts}}
## Constants
{{range .Consts}}### {{join .Names ", "}}
{{codeStart}}
{{value .Decl}}
{{codeEnd}}
{{comment 0 4 .Doc}}
{{end}}
{{end}}

{{- if .Vars}}
## Variables
{{range .Vars}}### {{join .Names ", "}}
{{codeStart}}
{{value .Decl}}
{{codeEnd}}
{{comment 0 4 .Doc}}
{{end}}
{{end}}

{{- if .Funcs}}
## Functions
{{range .Funcs}}### Func {{.Name}}
{{codeStart}}
{{func .Decl}}
{{codeEnd}}
{{comment 0 4 .Doc}}
{{end}}
{{end}}

{{- if .Types}}
## Types
{{range .Types}}### Type {{.Name}}
{{codeStart}}
{{type .Decl}}
{{codeEnd}}
{{comment 0 4 .Doc}}

{{- if .Consts}}
### Constants
{{range .Consts}}### {{join .Names ", "}}
{{codeStart}}
{{value .Decl}}
{{codeEnd}}
{{comment 0 4 .Doc}}
{{end}}
{{end}}
{{- if .Vars}}
### Variables
{{range .Vars}}### {{join .Names ", "}}
{{codeStart}}
{{value .Decl}}
{{codeEnd}}
{{comment 0 4 .Doc}}
{{end}}
{{end}}
{{- if .Funcs}}
### Functions
{{range .Funcs}}
{{codeStart}}
{{func .Decl}}
{{codeEnd}}
{{comment 0 4 .Doc}}
{{end}}
{{end}}
{{- if .Methods}}
### Methods
{{range .Methods}}
{{codeStart}}
{{func .Decl}}
{{codeEnd}}
{{comment 0 4 .Doc}}
{{end}}
{{end}}

{{end}}
{{end}}

{{- if .Examples}}
## Examples
{{range .Examples}}### {{exampleLink . }}
{{comment 0 4 .Doc}}
{{end}}
{{end}}

{{- if .Notes}}
{{range $marker, $notes := .Notes}}
### {{$marker}}
{{range $notes}}- {{.UID}}: {{.Body}}{{end}}
{{end}}
{{end}}

`

var markdownCommandTemplate = `# {{packageLink}}
{{if badges}}{{badges}}
{{end}}

{{comment 0 4 .Doc| guessHeadings 1 5 | highlight .CommandName }}
`
