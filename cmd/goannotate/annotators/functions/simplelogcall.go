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

// SimpleLogCall represents a function call generator for a logging
// call with the same signature as log.Callf and fmt.Printf.
type SimpleLogCall struct {
	EssentialOptions `yaml:",inline"`
	ContextType      string `yaml:"contextType" annotator:"type for the context parameter and result."`
}

// SimpleLogCallDescription documents SimpleLogCall.
const SimpleLogCallDescription = `
SimpleLogCall provides a functon call generator for generating calls to
functions with the same signature log.Callf and fmt.Printf.
`

func init() {
	RegisterCallGenerator(&SimpleLogCall{})
}

// UnmarshalYAML implements functions.CallGenerator.
func (sl *SimpleLogCall) UnmarshalYAML(buf []byte) error {
	return yaml.Unmarshal(buf, sl)
}

// Describe implements functions.CallGenerator.
func (sl *SimpleLogCall) Describe() string {
	return internal.MustDescribe(sl, SimpleLogCallDescription)
}

// Import implements functions.CallGenerator.
func (sl *SimpleLogCall) Import() string {
	return sl.ImportPath
}

var simpleCallTemplate = template.Must(template.New("call").Parse(`{{.FunctionName}}({{.Params}})`))

func (sl *SimpleLogCall) Generate(_ *token.FileSet, fn *types.Func, _ *ast.FuncDecl) (string, error) {
	sig := fn.Type().(*types.Signature)
	var ignore []int
	_, hasContext := derive.HasCustomContext(sig, sl.ContextType)
	if hasContext {
		ignore = append(ignore, 0)
	}
	params, paramArgs := derive.ArgsForParams(sig, ignore...)
	call := &strings.Builder{}
	data := struct {
		*SimpleLogCall
		Params string
	}{
		SimpleLogCall: sl,
		Params:        flatten(quote(params), paramArgs),
	}
	if err := simpleCallTemplate.Execute(call, data); err != nil {
		return "", err
	}
	return call.String(), nil
}
