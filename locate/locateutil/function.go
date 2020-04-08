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

type funcVisitor struct {
	callname string
	deferred bool
	nodes    []ast.Node
}

func callMatches(callexpr *ast.CallExpr, callname string) bool {
	switch id := callexpr.Fun.(type) {
	case *ast.Ident:
		return id.String() == callname
	case *ast.SelectorExpr:
		if sel, ok := id.X.(*ast.Ident); ok {
			return (sel.String() + "." + id.Sel.String()) == callname
		}
	case *ast.CallExpr:
		r := callMatches(id, callname)
		return r
	}
	return false
}

func (v *funcVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *ast.DeferStmt:
		if !v.deferred {
			return nil
		}
		if callMatches(n.Call, v.callname) {
			v.nodes = append(v.nodes, node)
		}
	case *ast.ExprStmt:
		if v.deferred {
			return nil
		}
		call, ok := n.X.(*ast.CallExpr)
		if !ok {
			return nil
		}
		if callMatches(call, v.callname) {
			v.nodes = append(v.nodes, node)
		}
	}
	return v
}

// FunctionCalls determines if the supplied function declaration contains a call
// 'callname' where callname is either a function name or a selector (eg. foo.bar).
// If deferred is true the function call must be defer'ed.
func FunctionCalls(decl *ast.FuncDecl, callname string, deferred bool) []ast.Node {
	if !HasBody(decl) {
		return nil
	}
	v := &funcVisitor{
		callname: callname,
		deferred: deferred,
	}
	ast.Walk(v, decl)
	return v.nodes
}

// HasBody returns true of the function has a body.
func HasBody(decl *ast.FuncDecl) bool {
	return len(decl.Body.List) > 0
}
