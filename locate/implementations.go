package locate

import (
	"context"
	"fmt"
	"go/types"

	"cloudeng.io/sync/errgroup"
)

func (t *T) findImplementations(ctx context.Context, packages []string) error {
	group, ctx := errgroup.WithContext(ctx)
	group = errgroup.WithConcurrency(group, t.options.concurrency)
	for _, pkg := range packages {
		pkg := pkg
		group.GoContext(ctx, func() error {
			return t.findImplementationInPackage(ctx, pkg)
		})
	}
	return group.Wait()
}

func (t *T) findImplementationInPackage(ctx context.Context, pkgPath string) error {
	pkg := t.loader.lookupPackage(pkgPath)
	if pkg == nil {
		return fmt.Errorf("locating interface implementations: failed to lookup: %v", pkgPath)
	}
	checked := pkg.TypesInfo
	// Look in info.Defs for functions.
	for k, obj := range checked.Defs {
		if obj == nil || !obj.Exported() || obj.Parent() != nil {
			continue
		}
		fn, ok := obj.(*types.Func)
		if !ok {
			continue
		}
		sig := fn.Type().(*types.Signature)
		rcv := sig.Recv()
		if rcv == nil || isInterfaceType(pkg.PkgPath, rcv.Type()) != nil {
			// ignore functions and abstract methods
			continue
		}
		// a concrete method
		t.mu.Lock()
		for ifcPath, ifcType := range t.interfaces {
			if types.Implements(rcv.Type(), ifcType.ifc) {
				t.addFunctionLocked(pkgPath, k.Pos(), fn, ifcPath)
			}
		}
		t.mu.Unlock()
	}
	return nil
}
