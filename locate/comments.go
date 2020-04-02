package locate

import (
	"context"
	"go/ast"
	"regexp"

	"cloudeng.io/errors"
	"golang.org/x/tools/go/packages"
)

type commentDesc struct {
	re       string
	filename string
	cg       *ast.CommentGroup
	node     ast.Node
	pkg      *packages.Package
	file     *ast.File
}

func (t *T) findComments(ctx context.Context, exprs []string) error {
	regexps := make([]*regexp.Regexp, len(exprs))
	errs := &errors.M{}
	for i, expr := range exprs {
		re, err := regexp.Compile(expr)
		errs.Append(err)
		regexps[i] = re
	}
	if err := errs.Err(); err != nil {
		return err
	}
	t.loader.walkFiles(func(filename string, pkg *packages.Package, cmap ast.CommentMap, file *ast.File) {
		for k, v := range cmap {
			for _, cg := range v {
				for _, re := range regexps {
					if re.MatchString(cg.Text()) {
						t.addComment(re.String(), filename, k, cg, pkg, file)
					}
				}
			}
		}
	})
	return nil
}

func (t *T) addComment(re string, filename string, node ast.Node, cg *ast.CommentGroup, pkg *packages.Package, file *ast.File) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.dirty[filename] |= HasComment
	t.comments[re] = append(t.comments[re], commentDesc{
		filename: filename,
		re:       re,
		node:     node,
		cg:       cg,
		pkg:      pkg,
		file:     file,
	})
}

// WalkComments calls the supplied function for each comment that was matched
// by the specified regular expressions. The function is called with the
// absolute filename, the node that the comment is associated with, the comment
// and the packates.Package to which the file belongs and its ast. The function
// is called in order of filename and then position within filename.
func (t *T) WalkComments(fn func(
	re string,
	absoluteFilename string,
	node ast.Node,
	cg *ast.CommentGroup,
	pkg *packages.Package,
	file *ast.File,
)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	sorted := make([]sortByPos, 0, len(t.functions))
	i := 0
	for _, v := range t.comments {
		for _, c := range v {
			sorted = append(sorted, sortByPos{
				name:    c.filename,
				pos:     c.pkg.Fset.PositionFor(c.cg.Pos(), false),
				payload: c,
			})
			i++
		}
	}
	sorter(sorted)
	for _, loc := range sorted {
		fnd := loc.payload.(commentDesc)
		fn(fnd.filename, fnd.re, fnd.node, fnd.cg, fnd.pkg, fnd.file)
	}
}
