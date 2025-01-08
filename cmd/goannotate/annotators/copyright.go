// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package annotators

import (
	"context"
	"fmt"
	"go/ast"
	"regexp"
	"strings"

	"cloudeng.io/errors"
	"cloudeng.io/go/cmd/goannotate/annotators/internal"
	"cloudeng.io/go/locate"
	"cloudeng.io/text/edit"
	"golang.org/x/tools/go/packages"
	"gopkg.in/yaml.v2"
)

// EnsureCopyrightAndLicense represents an annotator that can insert or replace
// copyright and license headers from go source code files.
type EnsureCopyrightAndLicense struct {
	EssentialOptions `yaml:",inline"`

	Copyright       string   `yaml:"copyright" annotator:"desired copyright notice."`
	Exclusions      []string `yaml:"exclusions" annotator:"regular expressions for files to be excluded."`
	License         string   `yaml:"license" annotator:"desired license notice."`
	UpdateCopyright bool     `yaml:"updateCopyright" annotator:"set to true to update existing copyright notice"`
	UpdateLicense   bool     `yaml:"updateLicense" annotator:"set to true to update existing license notice"`
}

func init() {
	Register(&EnsureCopyrightAndLicense{})
}

// New implements annotators.Annotators.
func (ec *EnsureCopyrightAndLicense) New(name string) Annotation {
	n := &EnsureCopyrightAndLicense{}
	n.Name = name
	return n
}

// UnmarshalYAML implements annotators.Annotations.
func (ec *EnsureCopyrightAndLicense) UnmarshalYAML(buf []byte) error {
	return yaml.Unmarshal(buf, ec)
}

// Describe implements annotators.Annotations.
func (ec *EnsureCopyrightAndLicense) Describe() string {
	return internal.MustDescribe(ec,
		`an annotator that ensures that a copyright and license notice is 
present at the top of all files. It will not remove existing notices.`,
	)
}

func compileREs(exprs []string) ([]*regexp.Regexp, error) {
	errs := errors.M{}
	exclusionREs := make([]*regexp.Regexp, len(exprs))
	for i, expr := range exprs {
		re, err := regexp.Compile(expr)
		if err != nil {
			errs.Append(fmt.Errorf("exclusion %v: failed to compile:%v", expr, err))
			continue
		}
		exclusionREs[i] = re
	}
	return exclusionREs, errs.Err()
}

// Do implements annotators.Annotations.
func (ec *EnsureCopyrightAndLicense) Do(ctx context.Context, root string, pkgs []string) error {
	if len(ec.Copyright) == 0 {
		return fmt.Errorf("missing or empty copyright specified in the configuration file")
	}
	exclusionREs, err := compileREs(ec.Exclusions)
	if err != nil {
		return err
	}
	locator := locate.New(
		concurrencyOpt(ec.Concurrency),
		locate.Trace(Verbosef),
		locate.IgnoreMissingFuctionsEtc(),
		locate.IncludeTests(),
	)
	if len(pkgs) == 0 {
		pkgs = ec.Packages
	}
	locator.AddPackages(pkgs...)
	Verbosef("locating functions to have a copyright/license annotation...")
	if err := locator.Do(ctx); err != nil {
		return fmt.Errorf("failed to locate functions and/or interface implementations: %v", err)
	}

	state := walkerState{
		EnsureCopyrightAndLicense: ec,
		dirty:                     map[string]bool{},
		edits:                     map[string][]edit.Delta{},
		exclusionREs:              exclusionREs,
		newCopyright:              strings.TrimSuffix(ec.Copyright, "\n") + "\n",
		newLicense:                strings.TrimSuffix(ec.License, "\n") + "\n\n",
	}
	locator.WalkFiles(state.determineEdits)
	return applyEdits(ctx, computeOutputs(root, state.edits), state.edits)
}

type walkerState struct {
	*EnsureCopyrightAndLicense
	dirty        map[string]bool
	edits        map[string][]edit.Delta
	exclusionREs []*regexp.Regexp
	newCopyright string
	newLicense   string
}

func (ws *walkerState) determineEdits(filename string,
	pkg *packages.Package,
	_ ast.CommentMap,
	file *ast.File,
	_ locate.HitMask) {

	for _, re := range ws.exclusionREs {
		if re.MatchString(filename) {
			ws.edits[filename] = nil
			ws.dirty[filename] = true
			return
		}
	}

	tokenFile := pkg.Fset.File(file.Pos())

	var copyright *ast.Comment
	var licenseStart, licenseLen int
	for _, cg := range file.Comments {
		if tokenFile.Offset(cg.Pos()) == 0 && cg != file.Doc {
			// Find comment block at top of file that is not a 'doc' comment.
			comments := cg.List
			sanitized := strings.TrimSpace(strings.ToLower(cg.Text()))
			if strings.HasPrefix(sanitized, "copyright") {
				copyright = comments[0]
				if len(comments) > 1 {
					at := comments[1].Pos()
					licenseStart = tokenFile.Offset(at)
					licenseLen = tokenFile.Offset(cg.End()) - licenseStart
				}
			}
		}
	}
	var deltas []edit.Delta
	if copyright != nil {
		if ws.UpdateCopyright {
			deltas = append(deltas, edit.ReplaceString(0, len(copyright.Text)+1, ws.newCopyright))
		}
		if licenseStart != 0 && len(ws.License) > 0 && ws.UpdateLicense {
			deltas = append(deltas, edit.ReplaceString(licenseStart, licenseLen, ws.newLicense))
		}
	} else {
		// New copy right and license.
		deltas = append(deltas, edit.InsertString(0, ws.newCopyright))
		if len(ws.License) > 0 {
			deltas = append(deltas, edit.InsertString(0, ws.newLicense))
		}
	}
	ws.edits[filename] = append(ws.edits[filename], deltas...)
	ws.dirty[filename] = true
}
