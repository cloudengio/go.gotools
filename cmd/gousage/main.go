// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"go/doc"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"cloudeng.io/cmdutil"
	"cloudeng.io/cmdutil/flags"
	"cloudeng.io/errors"
	"cloudeng.io/go/locate"
	"golang.org/x/tools/go/packages"
)

type cmdlineFlags struct {
	GoOutput  string `cmdline:"go-output,cmdusage.go,name of generated go file."`
	Overwrite bool   `cmdline:"overwrite,false,overwrite existing file."`
}

var cmdFlags cmdlineFlags

func main() {
	ctx := context.Background()
	if err := flags.RegisterFlagsInStruct(flag.CommandLine, "cmdline", &cmdFlags, nil, nil); err != nil {
		cmdutil.Exitf("failed to register flags: %v", err)
	}
	flag.Parse()
	pkgs := flag.Args()

	var docMode doc.Mode

	locator := locate.New()
	locator.AddPackages(pkgs...)
	if err := locator.Do(ctx); err != nil {
		cmdutil.Exitf("failed to run locator: %v", err)
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
		errs.Append(writeGo(filepath.Join(dir, cmdFlags.GoOutput), out))

	})
	if err := errs.Err(); err != nil {
		cmdutil.Exitf("%v", err)
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

// writeGo formats text with gofmt and writes it to filename, but only if it
// differs from that file's current contents, and prints the name of each file
// it writes. The usage is regenerated for every command on every run and most
// of it is typically unchanged, so this reports just those files that were
// updated. An existing file whose contents would change is overwritten only
// if --overwrite is specified.
func writeGo(filename string, text string) error {
	cmd := exec.Command("gofmt")
	cmd.Stdin = strings.NewReader(text)
	formatted, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	existing, err := os.ReadFile(filename)
	switch {
	case err == nil:
		if bytes.Equal(existing, formatted) {
			return nil
		}
		if !cmdFlags.Overwrite {
			return fmt.Errorf("cannot overwrite existing file: %v", filename)
		}
	case !os.IsNotExist(err):
		return fmt.Errorf("unexpected error for %v: %v", filename, err)
	}
	if err := os.WriteFile(filename, formatted, 0600); err != nil {
		return err
	}
	fmt.Printf("updated: %v\n", filename)
	return nil
}
