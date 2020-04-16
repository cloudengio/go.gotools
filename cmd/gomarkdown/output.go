package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/doc"
	"go/format"
	"go/token"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

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

var exitStatusRE = regexp.MustCompile(`exit status \d+$`)

func filterExitStatus(text string) string {
	out := strings.Builder{}
	sc := bufio.NewScanner(bytes.NewBufferString(text))
	first := true
	var prev string
	for sc.Scan() {
		if !first {
			out.WriteString(prev)
			out.WriteString("\n")
		}
		prev = sc.Text()
		first = false
	}
	if !exitStatusRE.MatchString(prev) {
		out.WriteString(prev)
		out.WriteString("\n")
	}
	return out.String()
}

type outputState struct {
	markdownGoCodeStart string
	doc                 *doc.Package
	pkg                 *packages.Package
	tplFuncs            map[string]interface{}
}

func newOutputState(markdownFlavour, goPkgSite string, doc *doc.Package, pkg *packages.Package) *outputState {
	st := &outputState{doc: doc, pkg: pkg}
	switch markdownFlavour {
	case "github":
		st.markdownGoCodeStart = "```go"
	default:
		st.markdownGoCodeStart = "```"
	}
	st.tplFuncs = map[string]interface{}{
		"codeStart":        st.codeStart,
		"codeEnd":          st.codeEnd,
		"comment":          st.comment,
		"filterGoGenerate": filterGoGenerate,
		"gocomment":        goComment,
		"func":             st.funcDecl,
		"type":             st.typeDecl,
		"value":            st.valueDecl,
		"join":             strings.Join,
	}
	switch goPkgSite {
	case "pkg.go.dev":
		st.tplFuncs["packageLink"] = st.pkgGoDevPackageLink
		st.tplFuncs["exampleLink"] = st.pkgGoDevExampleLink
	case "godoc.org":
		st.tplFuncs["packageLink"] = st.godocOrgPackageLink
		st.tplFuncs["exampleLink"] = st.godocOrgExampleLink
	default:
		panic(fmt.Sprintf("unsupported go pkg site: %v", goPkgSite))
	}
	return st
}

func (st *outputState) valueDecl(decl *ast.GenDecl) string {
	switch decl.Tok {
	case token.CONST, token.VAR:
		if ns := len(decl.Specs); ns != 1 {
			panic(fmt.Sprintf("unexpected # (%v) of Specs for const: %#v", ns, decl))
		}
		spec := decl.Specs[0].(*ast.ValueSpec)
		out := &strings.Builder{}
		format.Node(out, st.pkg.Fset, spec)
		return out.String()
	}
	return "unsupported"
}

func (st *outputState) funcDecl(decl *ast.FuncDecl) string {
	out := &strings.Builder{}
	format.Node(out, st.pkg.Fset, decl)
	return out.String()
}

func (st *outputState) typeDecl(decl *ast.GenDecl) string {
	out := &strings.Builder{}
	format.Node(out, st.pkg.Fset, decl)
	return out.String()
}

func (st *outputState) codeStart() string {
	return st.markdownGoCodeStart
}

func (st *outputState) codeEnd() string {
	return "```"
}

func (st *outputState) godocOrgPackageLink() string {
	return fmt.Sprintf("[%s](https://godoc.org/%s)", st.doc.Name, st.doc.ImportPath)
}

func (st *outputState) godocOrgExampleLink(eg *doc.Example) string {
	return fmt.Sprintf("[Example%s](https://godoc.org/%s#example-%s)",
		eg.Name, st.doc.ImportPath, eg.Name)
}

func (st *outputState) pkgGoDevPackageLink() string {
	return fmt.Sprintf("[%s](https://pkg.go.dev/%s?tab=doc)",
		st.doc.Name, st.doc.ImportPath)
}

func (st *outputState) pkgGoDevExampleLink(eg *doc.Example) string {
	return fmt.Sprintf("[Example%s](https://pkg.go.dev/%s?tab=doc#example-%s)",
		eg.Name, st.doc.ImportPath, eg.Name)
}

func (st *outputState) comment(indent, preIndent int, text string) string {
	out := &strings.Builder{}
	doc.ToText(out,
		text,
		strings.Repeat(" ", indent),
		strings.Repeat(" ", preIndent),
		80-indent-preIndent)
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

func (st *outputState) outputCommand(help string) (string, error) {
	tpl, err := template.New("markdown").Funcs(st.tplFuncs).Parse(markdownCommandTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to create template: %v", err)
	}
	out := &strings.Builder{}
	err = tpl.Execute(out, st.doc)
	return out.String(), err
}

func (st *outputState) outputGodoc(help string) (string, error) {
	tpl, err := template.New("markdown").Funcs(st.tplFuncs).Parse(godocTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to create template: %v", err)
	}
	tmp := struct {
		Usage string
	}{
		Usage: help,
	}
	out := &strings.Builder{}
	err = tpl.Execute(out, tmp)
	return out.String(), err
}

var markdownPackageTemplate = `# {{packageLink}}

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
{{end}}
{{end}}

{{- if .Examples}}
## Examples

{{range .Examples}}### {{exampleLink . }}
{{comment 0 4 .Doc}}
{{end}}
{{end}}
`

var markdownCommandTemplate = `# {{packageLink}}

# Command {{.ImportPath}}

{{comment 0 4 .Doc}}
`

var godocTemplate = `{{gocomment .Usage}}

package main
`
