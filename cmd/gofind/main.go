package main

import (
	"context"
	"flag"
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"regexp"

	"cloudeng.io/cmdutil/flags"
	"cloudeng.io/go/locate"
	"cloudeng.io/go/locate/locateutil"
	"golang.org/x/tools/go/packages"
)

var (
	InterfaceFlag string
	CommentFlag   string
	FunctionFlag  string
)

func init() {
	flag.StringVar(&InterfaceFlag, "interfaces", "", "if set, find all implementations of these interfaces in the speficied packages. The package local component of the interface name is treated as a regular expression")
	flag.StringVar(&CommentFlag, "comments", "", "if set, find all comments that match this regular expression in the specified packages.")
	flag.StringVar(&FunctionFlag, "functions", "", "if set, find all functions whose name matches this regular expression.")
	flag.Usage = func() {
		usage()
		flag.PrintDefaults()
	}
}

func usage() {
	fmt.Printf(`
gofind {--comments,--interfaces,--functions} package-list...

gofind will find interface implementations, functions and comments in a set of
go packages by building those packages and traversing the resulting data
structures. 

The following will find all implementations of io.Reader.* in the current
and sub-packages. (Note, that the package local component is interpreted as
a regular expression, so use 'io.Reader$' to restrict to exactly that string).

  gofind --interface=io.Reader ./...

The following will find all 

`)
}

func exit(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func multiple(args ...string) bool {
	set := 0
	for _, a := range args {
		if len(a) > 0 {
			set++
		}
	}
	return set > 1
}

func main() {
	ctx := context.Background()
	flag.Parse()

	if !flags.ExactlyOneSet(CommentFlag, FunctionFlag, InterfaceFlag) {
		exit("only one of --comments, --functions or --interfaces can be set")
	}
	var err error
	if len(InterfaceFlag) > 0 {
		err = handleInterfaces(ctx, InterfaceFlag, flag.Args())
	}
	if len(CommentFlag) > 0 {
		err = handleComments(ctx, CommentFlag, flag.Args())
	}
	if len(FunctionFlag) > 0 {
		err = handleFunctions(ctx, FunctionFlag, flag.Args())
	}
	if err != nil {
		exit("error: %v", err)
	}
}

func handleInterfaces(ctx context.Context, ifcs string, pkgs []string) error {
	locator := locate.New()
	locator.AddPackages(pkgs...)
	locator.AddInterfaces(ifcs)
	if err := locator.Do(ctx); err != nil {
		exit("locator.Do failed: %v", err)
	}
	locator.WalkFunctions(func(fullname string, pkg *packages.Package, file *ast.File, fn *types.Func, decl *ast.FuncDecl, implements []string) {
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
		exit("locator.Do failed: %v", err)
	}
	locator.WalkComments(func(re, absoluteFilename string, node ast.Node, cg *ast.CommentGroup, pkg *packages.Package, file *ast.File) {
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
		exit("locator.Do failed: %v", err)
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
