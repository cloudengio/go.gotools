package main

import (
	"context"
	"flag"
	"fmt"
	"go/ast"
	"go/doc"
	"io/ioutil"
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
	markdownFlag      string
	gopkgSiteFlag     string
	mdOutputFlag      string
	overwriteFlag     bool
	generateGoDocFlag bool
	goOutputFlag      string
)

func init() {
	flag.StringVar(&markdownFlag, "markdown", "github", "markdown style to use, currently only github is supported.")
	flag.StringVar(&gopkgSiteFlag, "gopkg", "pkg.go.dev", "link to this site for full godoc and godoc examples")
	flag.StringVar(&mdOutputFlag, "md-output", "README.md", "name of markdown output file.")
	flag.BoolVar(&overwriteFlag, "overwrite", false, "overwrite existing file.")
	flag.BoolVar(&generateGoDocFlag, "gocmd", false, "generate a go file with the --help output of the command packages")
	flag.StringVar(&goOutputFlag, "go-output", "cmdusage.go", "name of generated go file.")
}

func main() {
	ctx := context.Background()
	flag.Parse()
	pkgs := flag.Args()

	var docMode doc.Mode

	locator := locate.New(locate.IncludeTests())
	locator.AddPackages(pkgs...)
	if err := locator.Do(ctx); err != nil {
		cmdutil.Exit("failed to run locator: %v", err)
	}

	switch markdownFlag {
	case "github":
	default:
		cmdutil.Exit("unsupported mark down flavour: %v", markdownFlag)
	}

	switch gopkgSiteFlag {
	case "pkg.go.dev", "godoc.org":
	default:
		cmdutil.Exit("unsupported go pkg site: %v", gopkgSiteFlag)
	}

	// Merge the package and any associated test packages into a single
	// set of ast.Files for use with doc.NewFromFiles.
	merged := map[string][]*ast.File{}
	loaded := map[string]*packages.Package{}
	commands := map[string]bool{}

	errs := errors.M{}
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

	if err := errs.Err(); err != nil {
		cmdutil.Exit("%v", err)
	}

	for name, files := range merged {
		pkg := loaded[name]
		docPkg, err := doc.NewFromFiles(pkg.Fset, files, pkg.PkgPath, docMode)
		if err != nil {
			errs.Append(fmt.Errorf("failed to create ast.Package for %v: %v", pkg.PkgPath, err))
			return
		}
		// Merge all of the examples into the single package level examples
		// since the markdown will list all of the examples in one section.
		examples := docPkg.Examples
		for _, fneg := range docPkg.Funcs {
			examples = append(examples, fneg.Examples...)
		}
		for _, tyeg := range docPkg.Types {
			examples = append(examples, tyeg.Examples...)
		}
		docPkg.Examples = examples
		st := newOutputState(markdownFlag, gopkgSiteFlag, docPkg, pkg)
		dir := dirForPackage(pkg)
		var mdOutput string
		if commands[name] {
			help, err := helpText(ctx, name)
			if err != nil {
				errs.Append(err)
				continue
			}
			if generateGoDocFlag {
				out, err := st.outputCommand(filterExitStatus(help))
				if err != nil {
					errs.Append(err)
					continue
				}
				errs.Append(writeGo(filepath.Join(dir, goOutputFlag), out))
				continue
			}
			mdOutput, err = st.outputGodoc(filterExitStatus(help))
		} else {
			mdOutput, err = st.outputPackage()
		}
		if err != nil {
			errs.Append(err)
			continue
		}
		errs.Append(writeMarkdown(filepath.Join(dir, mdOutputFlag), mdOutput))
	}
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

func writeMarkdown(filename string, text string) error {
	if err := writeAllowed(filename); err != nil {
		return err
	}
	return ioutil.WriteFile(filename, []byte(text), 0622)
}

func writeGo(filename string, text string) error {
	if err := writeAllowed(filename); err != nil {
		return err
	}
	return ioutil.WriteFile(filename, []byte(text), 0622)
}
