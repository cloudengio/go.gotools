package locate

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/packages"
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

// Files returns a string representation of all files that contain interfaces
// or functions that were located..
func (t *T) Files() string {
	out := strings.Builder{}
	t.WalkFiles(func(name string, pkg *packages.Package, comments ast.CommentMap, file *ast.File, hitMask HitMask) {
		out.WriteString(name)
		out.WriteString(": ")
		out.WriteString(file.Name.String())
		out.WriteString(" (" + hitMask.String() + ")")
		out.WriteString("\n")
	})
	return out.String()
}

// WalkFiles calls the supplied function for each file that contains
// a located comment, interface or function, ordered by filename. The function
// is called with the absolute file name of the file, the packages.Package
// to which it belongs and its ast. The function is called in order of
// filename and then position within filename.
func (t *T) WalkFiles(fn func(
	absoluteFilename string,
	pkg *packages.Package,
	comments ast.CommentMap,
	file *ast.File,
	has HitMask,
)) {
	t.loader.walkFiles(func(
		filename string,
		pkg *packages.Package,
		comments ast.CommentMap,
		file *ast.File) {
		t.mu.Lock()
		has := t.dirty[filename]
		t.mu.Unlock()
		if has == 0 {
			return
		}
		fn(filename, pkg, comments, file, has)
	})
}
