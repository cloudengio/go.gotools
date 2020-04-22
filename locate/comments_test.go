// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package locate_test

import (
	"context"
	"go/ast"
	"testing"

	"cloudeng.io/go/locate"
	"golang.org/x/tools/go/packages"
)

func TestComments(t *testing.T) {
	ctx := context.Background()
	locator := locate.New()
	locator.AddComments(".*")
	locator.AddPackages(here+"data", here+"data/embedded", here+"comments")
	err := locator.Do(ctx)
	if err != nil {
		t.Fatalf("locate.Do: %v", err)
	}

	positions := []string{}
	locator.WalkComments(func(
		absoluteFilename string,
		re string,
		node ast.Node,
		cg *ast.CommentGroup,
		pkg *packages.Package,
		file *ast.File,
	) {
		pos := pkg.Fset.PositionFor(cg.Pos(), false)
		positions = append(positions, pos.String())
	})
	commentsAt := []string{
		"comments/doc.go:1:1",
		"comments/doc.go:4:11",
		"comments/doc.go:6:1",
		"comments/funcs.go:4:20",
		"data/embedded/embedded.go:17:1",
	}
	compareSlices(t, positions, commentsAt)
}
