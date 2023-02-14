// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/doc"
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

var exitStatusRE = regexp.MustCompile(`exit status \d+$`)

func filterUsage(text string) string {
	out := strings.Builder{}
	sc := bufio.NewScanner(bytes.NewBufferString(text))
	first := true
	var prev string
	for sc.Scan() {
		if strings.HasPrefix(sc.Text(), "go:generate") {
			continue
		}
		if strings.Contains(sc.Text(), "flag: help requested") {
			continue
		}
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
	doc      *doc.Package
	pkg      *packages.Package
	tplFuncs map[string]interface{}
}

func newOutputState(doc *doc.Package, pkg *packages.Package) *outputState {
	st := &outputState{doc: doc, pkg: pkg}
	st.tplFuncs = map[string]interface{}{
		"gocomment": goComment,
		"join":      strings.Join,
	}
	return st
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

var godocTemplate = `{{gocomment .Usage}}package main
`
