// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package locate_test

import (
	"context"
	"go/ast"
	"go/types"
	"path/filepath"
	"testing"

	"cloudeng.io/go/locate"
	"golang.org/x/tools/go/packages"
)

func TestFunctions(t *testing.T) {
	ctx := context.Background()
	locator := locate.New(locate.IgnoreMissingFuctionsEtc())
	locator.AddFunctions(here + "data.Fn2$")
	if err := locator.Do(ctx); err != nil {
		t.Fatalf("locator.Do: %v", err)
	}

	compareLocations(t, listFunctions(locator), []string{
		here + "data.Fn2",
	}, []string{
		"data/functions_more.go:3:1",
	})

	locator = locate.New()
	locator.AddFunctions(here + "data.Fn2$")
	locator.AddFunctions(here + "data")
	if err := locator.Do(ctx); err != nil {
		t.Fatalf("locator.Do: %v", err)
	}
	compareLocations(t, listFunctions(locator), []string{
		here + "data.Fn1",
		here + "data.Fn2",
	}, []string{
		"data/functions.go:7:1",
		"data/functions_more.go:3:1",
	})
	compareLocations(t, listFunctions(locator), []string{
		here + "data.Fn1",
		here + "data.Fn2",
	}, []string{
		"data/functions.go:7:1",
		"data/functions_more.go:3:1",
	})
	compareFiles(t, listFiles(locator),
		filepath.Join("data", "functions.go")+": data",
		filepath.Join("data", "functions_more.go")+": data",
	)

	compareSlices(t, listPackages(locator),
		[]string{"cloudeng.io/go/locate/testdata/data"})

}

func TestMethods(t *testing.T) {
	ctx := context.Background()
	locator := locate.New(locate.IncludeMethods(true))
	locator.AddFunctions(here + "data.Fn1$")
	if err := locator.Do(ctx); err != nil {
		t.Fatalf("locator.Do: %v", err)
	}
	compareLocations(t, listFunctions(locator), []string{
		"(*cloudeng.io/go/locate/testdata/data.rcv).Fn1",
		here + "data.Fn1",
	}, []string{
		"data/functions.go:13:1",
		"data/functions.go:7:1",
	})
}

func TestMainPackage(t *testing.T) {
	ctx := context.Background()
	locator := locate.New()
	locator.AddFunctions(here + "cmd")
	if err := locator.Do(ctx); err != nil {
		t.Fatalf("locator.Do: %v", err)
	}
	compareLocations(t, listFunctions(locator), []string{
		here + "cmd.InMain",
	}, []string{
		"cmd/main.go:6:1",
	})

	locator = locate.New(locate.IncludeMethods(true))
	locator.AddFunctions(here + "cmd.InMain")
	if err := locator.Do(ctx); err != nil {
		t.Fatalf("locator.Do: %v", err)
	}
	compareLocations(t, listFunctions(locator), []string{
		"(*cloudeng.io/go/locate/testdata/cmd.rcvr).InMain",
		here + "cmd.InMain",
	}, []string{
		"cmd/main.go:12:1",
		"cmd/main.go:6:1",
	})
}

func TestFunctionsAndInterfaces(t *testing.T) {
	ctx := context.Background()
	locator := locate.New()
	locator.AddFunctions(here + "data.Fn2$")
	locator.AddInterfaces(here + "data.Ifc2$")
	if err := locator.Do(ctx); err != nil {
		t.Fatalf("locator.Do: %v", err)
	}
	compareLocations(t, listInterfaces(locator), []string{
		here + "data.Ifc2 interface",
	}, []string{
		filepath.Join("data", "interfaces.go") + ":12:6",
	})
	compareLocations(t, listFunctions(locator), []string{
		here + "data.Fn2",
	}, []string{
		filepath.Join("data", "functions_more.go") + ":3:1",
	})
}

func TestFunctionDecls(t *testing.T) {
	ctx := context.Background()
	locator := locate.New()
	locator.AddFunctions(here + "data")
	locator.AddInterfaces(here + "data")
	locator.AddPackages(here+"data", here+"impl")
	if err := locator.Do(ctx); err != nil {
		t.Fatalf("locate.Do: %v", err)
	}
	start, stop := []string{}, []string{}
	locator.WalkFunctions(func(_ string, pkg *packages.Package, _ *ast.File, _ *types.Func, decl *ast.FuncDecl, _ []string) {
		begin := decl.Body.Pos()
		end := decl.Body.End()
		start = append(start, pkg.Fset.Position(begin).String())
		stop = append(stop, pkg.Fset.Position(end).String())
	})
	startAt := []string{
		"functions.go:7:18",
		"functions_more.go:3:23",
		"impls.go:5:22",
		"impls.go:9:28",
		"impls.go:15:31",
		"impls.go:22:23",
		"impls.go:26:29",
		"impls.go:30:32",
	}
	stopAt := []string{
		"functions.go:9:2",
		"functions_more.go:5:2",
		"impls.go:7:2",
		"impls.go:11:2",
		"impls.go:18:2",
		"impls.go:24:2",
		"impls.go:28:2",
		"impls.go:33:2",
	}
	compareSlices(t, start, startAt)
	compareSlices(t, stop, stopAt)

	compareSlices(t, listPackages(locator),
		[]string{
			"cloudeng.io/go/locate/testdata/data",
			"cloudeng.io/go/locate/testdata/impl",
		})
}
