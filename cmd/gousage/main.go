// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"go/doc"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"cloudeng.io/cmdutil"
	"cloudeng.io/errors"
	"cloudeng.io/go/locate"
	"golang.org/x/tools/go/packages"
)

var (
	overwriteFlag bool
	goOutputFlag  string
)

func init() {
	flag.BoolVar(&overwriteFlag, "overwrite", false, "overwrite existing file.")
	flag.StringVar(&goOutputFlag, "go-output", "cmdusage.go", "name of generated go file.")
}

func main() {
	ctx := context.Background()
	flag.Parse()
	pkgs := flag.Args()

	var docMode doc.Mode

	locator := locate.New()
	locator.AddPackages(pkgs...)
	if err := locator.Do(ctx); err != nil {
		cmdutil.Exit("failed to run locator: %v", err)
	}

	errs := errors.M{}
	locator.WalkPackages(func(pkg *packages.Package) {
		if pkg.Name != "main" {
			return
		}
		docPkg, err := doc.NewFromFiles(pkg.Fset, pkg.Syntax, pkg.PkgPath, docMode)
		if err != nil {
			errs.Append(fmt.Errorf("failed to create ast.Package for %v: %v", pkg.PkgPath, err))
			return
		}
		st := newOutputState(docPkg, pkg)
		dir := dirForPackage(pkg)

		help, err := helpText(ctx, pkg.PkgPath)
		if err != nil {
			errs.Append(err)
			return
		}
		out, err := st.outputGodoc(filterUsage(help))
		if err != nil {
			errs.Append(err)
			return
		}
		errs.Append(writeGo(filepath.Join(dir, goOutputFlag), out))

	})
	if err := errs.Err(); err != nil {
		cmdutil.Exit("%v", err)
	}
}

func dirForPackage(pkg *packages.Package) string {
	if len(pkg.CompiledGoFiles) == 0 {
		panic(fmt.Sprintf("no source files for %v\n", pkg.PkgPath))
	}
	return filepath.Dir(pkg.CompiledGoFiles[0])
}

func helpText(ctx context.Context, pkg string) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "run", pkg, "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// ignore exit errors.
			return string(out), nil
		}
		return "", fmt.Errorf("failed to run %v: %v", strings.Join(cmd.Args, " "), err)
	}
	return string(out), nil
}

func writeAllowed(filename string) error {
	if overwriteFlag {
		return nil
	}
	_, err := os.Stat(filename)
	if err == nil {
		return fmt.Errorf("cannot overwite existing file: %v", filename)
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("unexpected error for %v: %v", filename, err)
	}
	return nil
}

func writeGo(filename string, text string) error {
	cmd := exec.Command("gofmt")
	cmd.Stdin = strings.NewReader(text)
	formatted, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Printf("writing: %v\n", filename)
	if err := writeAllowed(filename); err != nil {
		return err
	}
	return os.WriteFile(filename, formatted, 0622)
}
