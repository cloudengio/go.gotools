// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package functions

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"
	"text/template"

	"cloudeng.io/go/cmd/goannotate/annotators/internal"
	"cloudeng.io/go/derive"
	"gopkg.in/yaml.v2"
)

// LogCallWithContext represents a function call generator for a logging
// call with the following signature:
//
//	func (ctx <contextType>, functionName, callerLocation, format string, arguments ...interface{}) func(ctx <contextType>, format string, namedResults ...interface{})
//
// See LogCallWithContextDescription for a complete description.
type LogCallWithContext struct {
	EssentialOptions `yaml:",inline"`
	ContextType      string `yaml:"contextType" annotator:"type for the context parameter and result."`
}

// LogCallWithContextDescription documents LogCallWithContext.
const LogCallWithContextDescription = `
LogCallWithContext provides a functon call generator for generating calls to
functions with the following signature:

  func (ctx <contextType>, functionName, format string, arguments ...interface{}) func(ctx <contextType>, format string, namedResults ...interface{}) 

These are invoked via defer as show below:

  defer <call>(ctx, "<function-name>",  "<format>", <parameters>....)(ctx, "<format>", <results>)

The actual type of the context is determined by the ContextType configuration
field. The parameters and named results are captured and passed to the logging
call according to cloudeng.io/go/derive.ArgsForParams and ArgsForResults.
The logging function must return a function that is defered to capture named
results and log them on function exit.
`

func init() {
	RegisterCallGenerator(&LogCallWithContext{})
}

// UnmarshalYAML implements functions.CallGenerator.
func (lc *LogCallWithContext) UnmarshalYAML(buf []byte) error {
	return yaml.Unmarshal(buf, lc)
}

// Describe implements functions.CallGenerator.
func (lc *LogCallWithContext) Describe() string {
	return internal.MustDescribe(lc, LogCallWithContextDescription)
}

// Import implements functions.CallGenerator.
func (lc *LogCallWithContext) Import() string {
	return lc.ImportPath
}

const ctxCallTemplateText = `defer {{.FunctionName}}({{.ContextParam}}, "{{.LoggedFunction}}", {{.Params}})({{.ContextParam}}, {{.Results}})`

var ctxCallTemplate = template.Must(template.New("call").Parse(ctxCallTemplateText))

func (lc *LogCallWithContext) Generate(_ *token.FileSet, fn *types.Func, _ *ast.FuncDecl) (string, error) {
	sig := fn.Type().(*types.Signature)
	var ignore []int
	ctxParam, hasContext := derive.HasCustomContext(sig, lc.ContextType)
	if hasContext {
		ignore = append(ignore, 0)
	}
	params, paramArgs := derive.ArgsForParams(sig, ignore...)
	results, resultArgs := derive.ArgsForResults(sig)
	if !hasContext || len(ctxParam) == 0 || ctxParam == "_" {
		ctxParam = "nil"
	}
	call := &strings.Builder{}
	data := struct {
		*LogCallWithContext
		// Used for templates.
		LoggedFunction string
		ContextParam   string
		Params         string
		Results        string
	}{
		LogCallWithContext: lc,
		ContextParam:       ctxParam,
		LoggedFunction:     fn.Pkg().Path() + "." + fn.Name(),
		Params:             flatten(quote(params), paramArgs),
		Results:            flatten(quote(results), resultArgs),
	}
	if err := ctxCallTemplate.Execute(call, data); err != nil {
		return "", err
	}
	return call.String(), nil
}
