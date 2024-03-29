// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package locateutil

import (
	"go/ast"
	"go/token"
)

// ImportBlock returns the start and end positions of an import statement
// or import block for the supplied file.
func ImportBlock(file *ast.File) (start, end token.Pos) {
	for _, d := range file.Decls {
		d, ok := d.(*ast.GenDecl)
		if !ok || d.Tok != token.IMPORT {
			break
		}
		if start == token.NoPos {
			start = d.Pos()
		}
		end = d.End()
	}
	return
}

// IsImportedByFile returns true if the supplied path appears in the Imports
// section of an ast.File.
func IsImportedByFile(file *ast.File, path string) bool {
	path = `"` + path + `"`
	for _, im := range file.Imports {
		if im.Path.Value == path {
			return true
		}
	}
	return false
}
