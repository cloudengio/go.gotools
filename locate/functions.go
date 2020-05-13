// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package locate

import (
	"context"
	"fmt"
	"go/ast"
	"go/types"
	"regexp"
	"sort"

	"cloudeng.io/go/locate/locateutil"
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
	funcs := locateutil.Functions(pkg, fnRE, !t.options.includeMethods)
	for _, fd := range funcs {
		t.addFunction(fd, pkgPath, "")
	}
	if !t.options.ignoreMissingFunctionsEtc && len(funcs) == 0 {
		return fmt.Errorf("failed to find any exported functions in %v for %s", pkgPath, fnRE)
	}
	return nil
}

func (t *T) addFunction(desc locateutil.FuncDesc, path string, implements string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.addFunctionLocked(desc, path, implements)
}

func (t *T) addFunctionLocked(desc locateutil.FuncDesc, path string, implements string) {
	fqn := desc.Type.FullName()
	var ifcs []string
	if len(implements) > 0 {
		//nolint:gocritic
		ifcs = append(t.functions[fqn].implements, implements)
		sort.Strings(ifcs)
		t.trace("method: %v implementing %v @ %v\n", fqn, implements, desc.Position)
	} else {
		t.trace("function: %v @ %v\n", fqn, desc.Position)
	}
	t.functions[fqn] = funcDesc{
		FuncDesc:   desc,
		path:       path,
		implements: ifcs,
	}
	t.dirty[desc.Position.Filename] |= HasFunction
}

type funcDesc struct {
	locateutil.FuncDesc
	path       string
	implements []string
}

// WalkFunctions calls the supplied function for each function location,
// ordered by filename and then position within file.
// The function is called with the packages.Package and ast for the file
// that contains the function, as well as the type and declaration of the
// function and the list of interfaces that implements. The function is called
// in order of filename and then position within filename.
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
			pos:     v.Position,
			payload: v,
		}
		i++
	}
	sorter(sorted)
	for _, loc := range sorted {
		fnd := loc.payload.(funcDesc)
		fn(loc.name, fnd.Package, fnd.File, fnd.Type, fnd.Decl, fnd.implements)
	}
}
