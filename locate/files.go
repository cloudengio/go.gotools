package locate

import (
	"go/ast"

	"golang.org/x/tools/go/packages"
)

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
		fn(filename, pkg, comments, file, has)
	})
}
