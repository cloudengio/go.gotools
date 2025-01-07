// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package locate

import (
	"context"
	"fmt"
	"go/types"
	"regexp"

	"cloudeng.io/go/locate/locateutil"
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

var allfuncs = regexp.MustCompile(".*")

func (t *T) findImplementationInPackage(_ context.Context, pkgPath string) error {
	pkg := t.loader.lookupPackage(pkgPath)
	if pkg == nil {
		return fmt.Errorf("locating interface implementations: failed to lookup: %v", pkgPath)
	}
	funcs := locateutil.Functions(pkg, allfuncs, false)
	for _, fd := range funcs {
		if !fd.Type.Exported() || fd.Decl == nil {
			// Ignore non-exported functions and interface function definitions
			// which do not have a declaration.
			continue
		}
		sig := fd.Type.Type().(*types.Signature)
		rcv := sig.Recv()
		if rcv == nil || locateutil.IsAbstract(fd.Type) {
			// Ignore functions and abstract methods.
			continue
		}
		// This is concrete method, check it against all interfaces.
		t.mu.Lock()
		for ifcPath, ifcType := range t.interfaces {
			if types.Implements(rcv.Type(), ifcType.ifc) {
				t.addFunctionLocked(fd, pkgPath, ifcPath)
			}
		}
		t.mu.Unlock()
	}
	return nil
}
