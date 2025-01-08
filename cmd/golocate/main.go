// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"go/ast"
	"go/types"
	"regexp"

	"cloudeng.io/cmdutil"
	"cloudeng.io/cmdutil/flags"
	"cloudeng.io/go/locate"
	"cloudeng.io/go/locate/locateutil"
	"golang.org/x/tools/go/packages"
)

var (
	interfaceFlag string
	commentFlag   string
	functionFlag  string
)

func init() {
	flag.StringVar(&interfaceFlag, "interfaces", "", "if set, find all implementations of these interfaces in the speficied packages. The package local component of the interface name is treated as a regular expression")
	flag.StringVar(&commentFlag, "comments", "", "if set, find all comments that match this regular expression in the specified packages.")
	flag.StringVar(&functionFlag, "functions", "", "if set, find all functions whose name matches this regular expression.")
}

func main() {
	ctx := context.Background()
	flag.Parse()

	if !flags.ExactlyOneSet(commentFlag, functionFlag, interfaceFlag) {
		cmdutil.Exit("only one of --comments, --functions or --interfaces can be set")
	}
	var err error
	if len(interfaceFlag) > 0 {
		err = handleInterfaces(ctx, interfaceFlag, flag.Args())
	}
	if len(commentFlag) > 0 {
		err = handleComments(ctx, commentFlag, flag.Args())
	}
	if len(functionFlag) > 0 {
		err = handleFunctions(ctx, functionFlag, flag.Args())
	}
	if err != nil {
		cmdutil.Exit("error: %v", err)
	}
}

func handleInterfaces(ctx context.Context, ifcs string, pkgs []string) error {
	locator := locate.New()
	locator.AddPackages(pkgs...)
	locator.AddInterfaces(ifcs)
	if err := locator.Do(ctx); err != nil {
		cmdutil.Exit("locator.Do failed: %v", err)
	}
	locator.WalkFunctions(func(_ string, pkg *packages.Package, _ *ast.File, fn *types.Func, _ *ast.FuncDecl, implements []string) {
		for _, ifc := range implements {
			pos := pkg.Fset.PositionFor(fn.Pos(), false)
			fmt.Printf("%v[%s]: %s\n", fn, ifc, pos)
		}
	})
	return nil
}

func handleComments(ctx context.Context, comments string, pkgs []string) error {
	locator := locate.New()
	locator.AddPackages(pkgs...)
	locator.AddComments(comments)
	if err := locator.Do(ctx); err != nil {
		cmdutil.Exit("locator.Do failed: %v", err)
	}
	locator.WalkComments(func(_, absoluteFilename string, node ast.Node, cg *ast.CommentGroup, pkg *packages.Package, _ *ast.File) {
		pos := pkg.Fset.PositionFor(cg.Pos(), false)
		fmt.Printf("%s: %T %s\n", absoluteFilename, node, pos)
	})
	return nil
}

func handleFunctions(ctx context.Context, functions string, pkgs []string) error {
	re, err := regexp.Compile(functions)
	if err != nil {
		return err
	}
	locator := locate.New()
	locator.AddPackages(pkgs...)
	if err := locator.Do(ctx); err != nil {
		cmdutil.Exit("locator.Do failed: %v", err)
	}
	// option for methods/functions only.
	locator.WalkPackages(func(pkg *packages.Package) {
		funcs := locateutil.Functions(pkg, re, false)
		for _, fn := range funcs {
			fmt.Printf("%v: %v\n", fn.Type.FullName(), fn.Position)
		}
	})
	return nil
}
