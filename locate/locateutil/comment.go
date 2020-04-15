// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package locateutil

import (
	"go/ast"
	"go/token"
	"strings"
)

// CommentGroupsContain returns if any of the supplied CommentGroups
// contain 'text'.
func CommentGroupsContain(comments []*ast.CommentGroup, text string) bool {
	if len(text) == 0 {
		return false
	}
	for _, cg := range comments {
		if strings.Contains(cg.Text(), text) {
			return true
		}
	}
	return false
}

// CommentGroupBounds returns the lowest and largest token.Pos of any
// of the supplied CommentGroups.
func CommentGroupBounds(comments []*ast.CommentGroup) (first, last token.Pos) {
	for _, cg := range comments {
		if first == token.NoPos {
			first = cg.Pos()
		}
		if last == token.NoPos {
			last = cg.End()
		}
		if cg.Pos() < first {
			first = cg.Pos()
		}
		if cg.End() > last {
			last = cg.End()
		}
	}
	return
}
