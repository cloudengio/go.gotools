// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"cloudeng.io/cmdutil"
	"cloudeng.io/cmdutil/flags"
	"cloudeng.io/errors"
	"cloudeng.io/go/locate"
	"golang.org/x/tools/go/packages"
)

type cmdlineFlags struct {
	Markdown     string `cmdline:"markdown,github,'markdown style to use, currently only github is supported.'"`
	GoPkgSite    string `cmdline:"gopkg,pkg.go.dev,link to this site for full godoc and godoc examples"`
	CircleCI     string `cmdline:"circleci,,set to the circleci project to insert a circleci build badge"`
	GoReportCard bool   `cmdline:"goreportcard,false,insert a link to goreportcard.com"`
	MDOutput     string `cmdline:"md-output,README.md,name of markdown output file."`
	Overwrite    bool   `cmdline:"overwrite,false,overwrite existing file."`
}

// validate reports whether the flags that accept one of a fixed set of
// values were given one of them.
func (c cmdlineFlags) validate() error {
	if err := flags.OneOf(c.Markdown).Validate("github"); err != nil {
		return err
	}
	return flags.OneOf(c.GoPkgSite).Validate("pkg.go.dev", "godoc.org")
}

var cmdFlags cmdlineFlags

func main() {
	ctx := context.Background()
	if err := flags.RegisterFlagsInStruct(flag.CommandLine, "cmdline", &cmdFlags, nil, nil); err != nil {
		cmdutil.Exitf("failed to register flags: %v", err)
	}
	flag.Parse()
	if err := cmdFlags.validate(); err != nil {
		cmdutil.Exitf("%s", err)
	}

	locator := locate.New(locate.IncludeTests())
	locator.AddPackages(flag.Args()...)
	if err := locator.Do(ctx); err != nil {
		cmdutil.Exitf("failed to run locator: %v", err)
	}

	if err := generateMarkdown(locator); err != nil {
		cmdutil.Exitf("%v", err)
	}
}

// generateMarkdown writes the markdown for every package that locator found
// to the markdown output file in that package's directory. Errors are
// accumulated so that a failure for one package does not prevent the
// markdown for the others from being generated.
func generateMarkdown(locator *locate.T) error {
	merged, loaded, commands := mergePackages(locator)
	errs := errors.M{}
	for name, files := range merged {
		pkg := loaded[name]
		text, err := markdownForPackage(pkg, files, commands[name])
		if err != nil {
			errs.Append(err)
			continue
		}
		errs.Append(writeMarkdown(filepath.Join(dirForPackage(pkg), cmdFlags.MDOutput), text))
	}
	return errs.Err()
}

// mergePackages merges each package and any associated test package into a
// single set of ast.Files for use with doc.NewFromFiles. It returns those
// files, the loaded package and whether the package is a command, each keyed
// by the package path with any _test suffix removed.
func mergePackages(locator *locate.T) (merged map[string][]*ast.File, loaded map[string]*packages.Package, commands map[string]bool) {
	merged = map[string][]*ast.File{}
	loaded = map[string]*packages.Package{}
	commands = map[string]bool{}
	locator.WalkPackages(func(pkg *packages.Package) {
		if pkg.Name == "main" {
			commands[pkg.PkgPath] = true
		}
		if strings.HasSuffix(pkg.PkgPath, ".test") {
			// ignore compiled test code (ie. the generated main).
			return
		}
		loaded[pkg.PkgPath] = pkg
		baseName := strings.TrimSuffix(pkg.PkgPath, "_test")
		merged[baseName] = append(merged[baseName], pkg.Syntax...)
	})
	return merged, loaded, commands
}

// markdownForPackage returns the markdown for a single package, which is
// formatted as a command if isCommand is set.
func markdownForPackage(pkg *packages.Package, files []*ast.File, isCommand bool) (string, error) {
	var docMode doc.Mode
	docPkg, err := packageDoc(pkg, files, docMode)
	if err != nil {
		return "", err
	}
	// Merge all of the examples into the single package level Examples
	// since the markdown will list all of the examples in one section.
	docPkg.Examples = allExamples(docPkg)

	dir := dirForPackage(pkg)
	extraMD, err := loadExtraMarkdown(dir, cmdFlags.MDOutput)
	if err != nil {
		return "", err
	}
	st := newOutputState(docPkg, pkg,
		extraMarkdown(extraMD),
		markdownFlavour(cmdFlags.Markdown),
		goPkgSite(cmdFlags.GoPkgSite),
		goreportcard(cmdFlags.GoReportCard),
		circleciProject(cmdFlags.CircleCI),
	)
	if isCommand {
		return st.outputCommand()
	}
	return st.outputPackage()
}

// packageDoc returns the documentation for pkg as parsed into files, falling
// back to fallbackPackageDoc if that cannot be determined.
func packageDoc(pkg *packages.Package, files []*ast.File, mode doc.Mode) (*doc.Package, error) {
	docPkg, err := doc.NewFromFiles(pkg.Fset, files, pkg.PkgPath, mode)
	if err != nil {
		if fallback, ferr := fallbackPackageDoc(pkg, mode); fallback != nil || ferr != nil {
			docPkg, err = fallback, ferr
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create ast.Package for %v: %v", pkg.PkgPath, err)
	}
	return docPkg, nil
}

// fallbackPackageDoc determines a package's documentation by parsing its
// original Go source files directly. It is used when doc.NewFromFiles fails
// for the files the package was loaded from, eg. for cgo files in the go
// build cache, which have no .go extension. It returns nil, nil if there are
// no source files to fall back to.
func fallbackPackageDoc(pkg *packages.Package, mode doc.Mode) (*doc.Package, error) {
	if len(pkg.GoFiles) == 0 {
		return nil, nil
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, file := range pkg.GoFiles {
		// Files that cannot be parsed are skipped rather than reported, the
		// aim being to recover as much documentation as possible.
		if parsed, err := parser.ParseFile(fset, file, nil, parser.ParseComments); err == nil {
			files = append(files, parsed)
		}
	}
	if len(files) == 0 {
		return nil, nil
	}
	return doc.NewFromFiles(fset, files, pkg.PkgPath, mode)
}

// allExamples returns every example in docPkg: those at package level and
// those associated with its functions and types.
func allExamples(docPkg *doc.Package) []*doc.Example {
	examples := docPkg.Examples
	for _, fn := range docPkg.Funcs {
		examples = append(examples, fn.Examples...)
	}
	for _, typ := range docPkg.Types {
		examples = append(examples, typ.Examples...)
	}
	return examples
}

func dirForPackage(pkg *packages.Package) string {
	files := pkg.GoFiles
	if len(files) == 0 {
		files = pkg.CompiledGoFiles
	}
	if len(files) == 0 {
		panic(fmt.Sprintf("no source files for %v\n", pkg.PkgPath))
	}
	return filepath.Dir(files[0])
}

// writeMarkdown writes text to filename, but only if it differs from that
// file's current contents, and prints the name of each file it writes. The
// markdown for every package is regenerated on every run and most of it is
// typically unchanged, so this reports just those files that were updated.
// An existing file whose contents would change is overwritten only if
// --overwrite is specified.
func writeMarkdown(filename, text string) error {
	existing, err := os.ReadFile(filename)
	switch {
	case err == nil:
		if string(existing) == text {
			return nil
		}
		if !cmdFlags.Overwrite {
			return fmt.Errorf("cannot overwrite existing file: %v", filename)
		}
	case !os.IsNotExist(err):
		return fmt.Errorf("unexpected error for %v: %v", filename, err)
	}
	if err := os.WriteFile(filename, []byte(text), 0600); err != nil {
		return err
	}
	fmt.Printf("updated: %v\n", filename)
	return nil
}
