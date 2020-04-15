// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package annotators

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"
	"text/template"

	"cloudeng.io/errors"
	"cloudeng.io/go/derive"
	"cloudeng.io/go/locate"
	"cloudeng.io/go/locate/locateutil"
	"cloudeng.io/text/edit"
	"golang.org/x/tools/go/packages"
	"gopkg.in/yaml.v2"
)

// AddLogCall represents an annotator for adding a function call that logs
// the entry and exit to every function and method that is matched by the
// locator.
type AddLogCall struct {
	Type                 string   `annotator:"name of annotator type."`
	Name                 string   `annotator:"name of annotation."`
	Packages             []string `annotator:"packages to be annotated"`
	Interfaces           []string `annotator:"list of interfaces whose implementations are to have logging calls added to them."`
	Functions            []string `annotator:"list of functionms that are to have function calls added to them."`
	ContextType          string   `yaml:"contextType" annotator:"type for the context parameter and result."`
	Import               string   `annotator:"import patrh for the logging function."`
	Logcall              string   `annotator:"invocation for the logging function."`
	IgnoreEmptyFunctions bool     `yaml:"ignoreEmptyFunctions" annotator:"if set empty functions are ignored."`
	Concurrency          int      `annotator:"the number of goroutines to use, zero for a sensible default."`

	// Used for templates.
	FunctionName string `yaml:",omitempty"`
	Tag          string `yaml:",omitempty"`
	ContextParam string `yaml:",omitempty"`
	Location     string `yaml:",omitempty"`
	Params       string `yaml:",omitempty"`
	Results      string `yaml:",omitempty"`
}

func init() {
	Register(&AddLogCall{})
}

// New implements annotators.Annotator.
func (lc *AddLogCall) New(name string) Annotation {
	return &AddLogCall{Name: name}
}

// UnmarshalYAML implements annotators.Annotation.
func (lc *AddLogCall) UnmarshalYAML(buf []byte) error {
	return yaml.Unmarshal(buf, lc)
}

// Describe implements annotators.Annotation.
func (lc *AddLogCall) Describe() string {
	return MustDescribe(lc, `an annotator to add function calls that are intended to log entry
and exit from functions. The signature of these calls is:

func (ctx <contextType>, functionName, callerLocation, format string, arguments ...interface{}) func(ctx <contextType>, format string, namedResults ...interface{}) 

and their invocation:

defer <call>(ctx, "<function-name>", "<location>", "<format>", <parameters>....)(ctx, "<format>", <results>)

The actual type of the context is determied by the ContextType configuration
field. The parameters and named results are captured and passed to the logging
call according to cloudeng.io/go/derive.ArgsForParams and ArgsForResults.
The logging function must return a function that is defered to capture named
results and log them on function exit.
`)
}

// Do implements annotators.Annotation.
func (lc *AddLogCall) Do(ctx context.Context, root string, pkgs []string) error {
	locator := locate.New(
		concurrencyOpt(lc.Concurrency),
		locate.Trace(Verbosef),
		locate.IgnoreMissingFuctionsEtc(),
	)
	locator.AddInterfaces(lc.Interfaces...)
	locator.AddFunctions(lc.Functions...)
	locator.AddPackages(pkgs...)
	Verbosef("locating functions to be annotated with a logcall...")
	if err := locator.Do(ctx); err != nil {
		return fmt.Errorf("failed to locate functions and/or interface implementations: %v\n", err)
	}

	dirty := map[string]bool{}
	edits := map[string][]edit.Delta{}
	errs := &errors.M{}
	locator.WalkFunctions(func(fullname string,
		pkg *packages.Package,
		file *ast.File,
		fn *types.Func,
		decl *ast.FuncDecl,
		implements []string) {
		if lc.IgnoreEmptyFunctions && !locateutil.HasBody(decl) {
			return
		}
		invovation, comment, err := lc.annotationForFunc(pkg.Fset, fn, decl)
		if err != nil {
			errs.Append(err)
			return
		}
		if lc.alreadyAnnotated(pkg.Fset, file, fn, decl, comment) {
			Verbosef("%v: already annotated\n", fullname)
			return
		}
		lbrace := pkg.Fset.PositionFor(decl.Body.Lbrace, false)
		delta := edit.InsertString(lbrace.Offset+1, invovation+" // "+comment)
		edits[lbrace.Filename] = append(edits[lbrace.Filename], delta)
		dirty[lbrace.Filename] = true
		Verbosef("function: %v @ %v\n", fullname, lbrace)
	})

	importStatement := "\n" + `import "` + lc.Import + `"` + "\n"

	locator.WalkFiles(func(filename string,
		pkg *packages.Package,
		comments ast.CommentMap,
		file *ast.File,
		mask locate.HitMask) {
		if !dirty[filename] || ((mask | locate.HasFunction) == 0) {
			return
		}
		if lc.alreadyImported(file, lc.Import) {
			Verbosef("%v: %v: already imported\n", filename, lc.Import)
			return
		}
		_, end := locateutil.ImportBlock(file)
		if end == token.NoPos {
			end = file.Name.End()
		}
		pos := pkg.Fset.PositionFor(end, false)
		delta := edit.InsertString(pos.Offset, importStatement)
		edits[pos.Filename] = append(edits[pos.Filename], delta)
		Verbosef("import: %v @ %v\n", lc.Import, pos)
	})

	if err := errs.Err(); err != nil {
		return err
	}
	return applyEdits(ctx, computeOutputs(root, edits), edits)
}

const commentTemplateText = `DO NOT EDIT, AUTO GENERATED BY {{.Tag}}#{{.Name}}`
const callTemplateText = `
defer {{.Logcall}}({{.ContextParam}}, "{{.FunctionName}}", "{{.Location}}", {{.Params}})({{.ContextParam}}, {{.Results}})`

var callTemplate = template.Must(template.New("call").Parse(callTemplateText))
var commentTemplate = template.Must(template.New("comment").Parse(commentTemplateText))

func quote(s string) string {
	return `"` + s + `"`
}

func flatten(format string, args []string) string {
	if len(args) == 0 {
		return format
	}
	return format + ", " + strings.Join(args, ", ")
}

func (lc *AddLogCall) alreadyImported(file *ast.File, path string) bool {
	path = `"` + path + `"`
	for _, im := range file.Imports {
		if im.Path.Value == path {
			return true
		}
	}
	return false
}

func (lc *AddLogCall) alreadyAnnotated(fset *token.FileSet, file *ast.File, fn *types.Func, decl *ast.FuncDecl, comment string) bool {
	if !locateutil.HasBody(decl) {
		return false
	}
	deferStmt, ok := decl.Body.List[0].(*ast.DeferStmt)
	if !ok {
		return false
	}
	cmap := ast.NewCommentMap(fset, file, file.Comments)
	comments := cmap[deferStmt]
	for _, c := range comments {
		if c := c.Text(); strings.HasPrefix(c, comment) {
			return true
		}
	}
	return false
}

func (lc *AddLogCall) annotationForFunc(fset *token.FileSet, fn *types.Func, decl *ast.FuncDecl) (string, string, error) {
	sig := fn.Type().(*types.Signature)
	var ignore []int
	ctxParam, hasContext := derive.HasCustomContext(sig, lc.ContextType)
	if hasContext {
		ignore = append(ignore, 0)
	}
	params, paramArgs := derive.ArgsForParams(sig, ignore...)
	results, resultArgs := derive.ArgsForResults(sig)
	if !hasContext {
		ctxParam = "nil"
	}
	call, comment := &strings.Builder{}, &strings.Builder{}
	pos := fset.Position(decl.Pos())
	parent, base := filepath.Base(filepath.Dir(pos.Filename)), filepath.Base(pos.Filename)
	location := fmt.Sprintf("%s%c%s:%d", parent, filepath.Separator, base, pos.Line)
	lc.ContextParam = ctxParam
	lc.Tag = lc.Type
	lc.FunctionName = fn.Pkg().Path() + "." + fn.Name()
	lc.Location = location
	lc.Params = flatten(quote(params), paramArgs)
	lc.Results = flatten(quote(results), resultArgs)
	if err := callTemplate.Execute(call, lc); err != nil {
		return "", "", err
	}
	if err := commentTemplate.Execute(comment, lc); err != nil {
		return "", "", err
	}
	return call.String(), comment.String(), nil
}
