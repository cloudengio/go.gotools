// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package functions_test

import (
	"go/ast"
	"go/types"
	"reflect"
	"testing"

	"cloudeng.io/errors"
	"cloudeng.io/go/cmd/goannotate/annotators/functions"
	"cloudeng.io/go/cmd/goannotate/annotators/internal/testutil"
	"golang.org/x/net/context"
	"golang.org/x/tools/go/packages"
)

func execute(t *testing.T, typeName string) (string, []string) {
	const here = "cloudeng.io/go/cmd/goannotate/annotators/functions"
	ctx := context.Background()
	testutil.SetupFunctions(t)
	locator := testutil.LocatePackages(ctx, t, here+"/testdata/sample")
	generator := functions.Lookup(here + typeName)
	var calls []string
	errs := &errors.M{}
	locator.WalkFunctions(func(fullname string, pkg *packages.Package, file *ast.File, fn *types.Func, decl *ast.FuncDecl, implements []string) {
		call, err := generator.Generate(pkg.Fset, fn, decl)
		calls = append(calls, call)
		errs.Append(err)
	})
	if err := errs.Err(); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	return generator.Import(), calls
}

func TestCtxLogCall(t *testing.T) {
	importPath, calls := execute(t, ".LogCallWithContext")
	if got, want := importPath, "cloudeng.io/go/cmd/goannotate/annotators/testdata/apilog"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	expectedCalls := []string{
		`defer apilog.LogCallf(ctx, "cloudeng.io/go/cmd/goannotate/annotators/functions/testdata/sample.ExampleCtx", "a=%d", a)(ctx, "err=%v", err)`,
		`defer apilog.LogCallf(nil, "cloudeng.io/go/cmd/goannotate/annotators/functions/testdata/sample.Example", "a=%d", a)(nil, "_=?")`,
	}
	if got, want := calls, expectedCalls; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
