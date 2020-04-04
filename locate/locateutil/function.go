// Package locateutil provides utility routines for use with its parent
// locate package.
package locateutil

import (
	"go/ast"
	"go/token"
	"go/types"

	"regexp"
	"sort"

	"golang.org/x/tools/go/packages"
)

// FuncDesc represents a function definition, declaration and the file
// and position within that file. Decl will be nil if Abstract is true.
type FuncDesc struct {
	Type     *types.Func
	Abstract bool
	Decl     *ast.FuncDecl
	File     *ast.File
	Position token.Position
	Package  *packages.Package
}

func Functions(pkg *packages.Package, re *regexp.Regexp, noMethods bool) []FuncDesc {
	descs := []FuncDesc{}
	for k, obj := range pkg.TypesInfo.Defs {
		if obj == nil || !k.IsExported() || !re.MatchString(k.Name) {
			continue
		}
		fn, ok := obj.(*types.Func)
		if !ok {
			continue
		}
		recv := fn.Type().(*types.Signature).Recv()
		abstract := false
		if recv != nil {
			if noMethods {
				continue
			}
			abstract = IsAbstract(fn)
		}
		pos := pkg.Fset.PositionFor(k.Pos(), false)
		var file *ast.File
		for i, v := range pkg.CompiledGoFiles {
			if v == pos.Filename {
				file = pkg.Syntax[i]
			}
		}
		descs = append(descs, FuncDesc{
			Package:  pkg,
			Type:     fn,
			Abstract: abstract,
			Decl:     findFuncOrMethodDecl(fn, file),
			Position: pos,
			File:     file,
		})
	}
	sort.SliceStable(descs, func(i, j int) bool {
		if descs[i].Position.Filename == descs[j].Position.Filename {
			return descs[i].Position.Offset < descs[j].Position.Offset
		}
		return descs[i].Position.Filename < descs[j].Position.Filename
	})
	return descs
}

func findFuncOrMethodDecl(fn *types.Func, file *ast.File) *ast.FuncDecl {
	for _, d := range file.Decls {
		d, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if d.Name.NamePos == fn.Pos() {
			return d
		}
	}
	return nil
}
