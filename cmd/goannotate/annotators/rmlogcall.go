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
	"regexp"

	"cloudeng.io/go/cmd/goannotate/annotators/internal"
	"cloudeng.io/go/locate"
	"cloudeng.io/go/locate/locateutil"
	"cloudeng.io/text/edit"
	"golang.org/x/tools/go/packages"
	"gopkg.in/yaml.v2"
)

// RmLogCall represents an annotor for removing logging calls.
type RmLogCall struct {
	EssentialOptions `yaml:",inline"`
	LocateOptions    `yaml:",inline"`

	FunctionNameRE string `yaml:"functionNameRE" annotator:"the function call (regexp) to be removed"`
	Comment        string `yaml:"comment" annotator:"optional comment that must appear in the comments associated with the function call if it is to be removed."`
	Deferred       bool   `yaml:"deferred" annotator:"if set requires that the function to be removed must be defered."`
}

func init() {
	Register(&RmLogCall{})
}

// New implements annotators.Annotator.
func (rc *RmLogCall) New(name string) Annotation {
	n := &RmLogCall{}
	n.Name = name
	return n
}

// UnmarshalYAML implements annotators.Annotation.
func (rc *RmLogCall) UnmarshalYAML(buf []byte) error {
	return yaml.Unmarshal(buf, rc)
}

// Describe implements annotators.Annotation.
func (rc *RmLogCall) Describe() string {
	return internal.MustDescribe(rc, "an annotator that removes instances of calls to functions.")
}

// Do implements annotators.Annotation.
func (rc *RmLogCall) Do(ctx context.Context, root string, pkgs []string) error {
	logcallRE, err := regexp.Compile(rc.FunctionNameRE)
	if err != nil {
		return err
	}
	locator := locate.New(
		concurrencyOpt(rc.Concurrency),
		locate.Trace(Verbosef),
		locate.IgnoreMissingFuctionsEtc(),
	)
	locator.AddInterfaces(rc.Interfaces...)
	locator.AddFunctions(rc.Functions...)
	if len(pkgs) == 0 {
		pkgs = rc.Packages
	}
	locator.AddPackages(pkgs...)
	Verbosef("locating functions to have a logcall annotation removal...")
	if err := locator.Do(ctx); err != nil {
		return fmt.Errorf("failed to locate functions and/or interface implementations: %v", err)
	}

	commentMaps := locator.MakeCommentMaps()

	edits := map[string][]edit.Delta{}
	locator.WalkFunctions(func(fullname string,
		pkg *packages.Package,
		file *ast.File,
		fn *types.Func,
		decl *ast.FuncDecl,
		implements []string) {
		if locateutil.FunctionStatements(decl) == 0 {
			return
		}
		nodes := locateutil.FunctionCalls(decl, logcallRE, rc.Deferred)
		if len(nodes) == 0 {
			return
		}
		var start, end token.Pos
		cmap := commentMaps[file]
		for _, node := range nodes {
			start = node.Pos()
			end = node.End()
			cgs := cmap[node]
			if !locateutil.CommentGroupsContain(cgs, rc.Comment) {
				continue
			}
			cFirst, cLast := locateutil.CommentGroupBounds(cgs)
			if cFirst < start {
				start = cFirst
			}
			if cLast > end {
				end = cLast
			}
			from := pkg.Fset.PositionFor(start, false)
			to := pkg.Fset.PositionFor(end, false)
			delta := edit.Delete(from.Offset, to.Offset-from.Offset+1)
			edits[from.Filename] = append(edits[from.Filename], delta)
			Verbosef("delete: %v...%v\n", from, to)
		}
	})
	return applyEdits(ctx, computeOutputs(root, edits), edits)
}
