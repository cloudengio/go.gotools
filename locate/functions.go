package locate

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"regexp"
	"sort"
	"strings"

	"cloudeng.io/sync/errgroup"
	"golang.org/x/tools/go/packages"
)

func (t *T) findFunctions(ctx context.Context, functions []string) error {
	group, ctx := errgroup.WithContext(ctx)
	group = errgroup.WithConcurrency(group, t.options.concurrency)
	for _, name := range functions {
		pkgPath, nameRE, err := getPathAndRegexp(name)
		if err != nil {
			return err
		}
		group.GoContext(ctx, func() error {
			return t.findFunctionsInPackage(ctx, pkgPath, nameRE)
		})
	}
	return group.Wait()
}

func (t *T) findFunctionsInPackage(ctx context.Context, pkgPath string, fnRE *regexp.Regexp) error {
	pkg := t.loader.lookupPackage(pkgPath)
	if pkg == nil {
		return fmt.Errorf("locating functions: failed to lookup: %v", pkgPath)
	}
	found := 0
	checked := pkg.TypesInfo
	// Look in info.Defs for functions.
	for k, obj := range checked.Defs {
		if obj == nil || !k.IsExported() || !fnRE.MatchString(k.Name) {
			continue
		}
		fn, ok := obj.(*types.Func)
		if !ok {
			continue
		}
		if fn.Type().(*types.Signature).Recv() != nil {
			// either a method or an abstract function.
			continue
		}
		found++
		t.addFunction(pkgPath, k.Pos(), fn, "")
	}
	if !t.options.ignoreMissingFunctionsEtc && found == 0 {
		return fmt.Errorf("failed to find any exported functions in %v for %s", pkgPath, fnRE)
	}
	return nil
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

func (t *T) addFunction(path string, pos token.Pos, fnType *types.Func, implements string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.addFunctionLocked(path, pos, fnType, implements)
}

func (t *T) addFunctionLocked(path string, pos token.Pos, fnType *types.Func, implements string) {
	position := t.loader.position(path, pos)
	filename := position.Filename
	ast, _, _ := t.loader.lookupFile(filename)
	decl := findFuncOrMethodDecl(fnType, ast)
	if decl == nil {
		// Ignore method definitions which do not have a body. For example,
		// the following definition will have a method Fn1 which does not
		// have a body.
		// type S struct { A interface { Fn1() } }
		return
	}
	fqn := fnType.FullName()
	desc := t.functions[fqn]
	var ifcs []string
	if len(implements) > 0 {
		ifcs = append(desc.implements, implements)
		t.trace("method: %v implementing %v @ %v\n", fqn, implements, position)
	} else {
		t.trace("function: %v @ %v\n", fqn, position)
	}
	t.functions[fqn] = funcDesc{
		path:       path,
		fn:         fnType,
		decl:       decl,
		position:   position,
		implements: ifcs,
	}
	t.dirty[filename] |= HasFunction
}

type funcDesc struct {
	path       string
	fn         *types.Func
	decl       *ast.FuncDecl
	position   token.Position
	implements []string
}

// WalkFunctions calls the supplied function for each function location,
// ordered by filename and then position within file.
// The function is called with the packages.Package and ast for the file
// that contains the function, as well as the type and declaration of the
// function and the list of interfaces that implements.
func (t *T) WalkFunctions(fn func(
	fullname string,
	pkg *packages.Package,
	file *ast.File,
	fn *types.Func,
	decl *ast.FuncDecl,
	implements []string)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	sorted := make([]sortByPos, len(t.functions))
	i := 0
	for k, v := range t.functions {
		sorted[i] = sortByPos{
			name:    k,
			pos:     v.position,
			payload: v,
		}
		i++
	}
	sorter(sorted)
	for _, loc := range sorted {
		fnd := loc.payload.(funcDesc)
		file, _, pkg := t.loader.lookupFile(fnd.position.Filename)
		fn(loc.name, pkg, file, fnd.fn, fnd.decl, fnd.implements)
	}
}

// Functions returns a string representation of all function locations.
func (t *T) Functions() string {
	out := strings.Builder{}
	t.WalkFunctions(func(name string, pkg *packages.Package, file *ast.File, fn *types.Func, decl *ast.FuncDecl, implements []string) {
		out.WriteString(name)
		if len(implements) > 0 {
			out.WriteString(" implements ")
			sort.Strings(implements)
			out.WriteString(strings.Join(implements, ", "))
		}
		out.WriteString(" @ ")
		out.WriteString(pkg.Fset.PositionFor(decl.Type.Func, false).String())
		out.WriteString("\n")
	})
	return out.String()
}
