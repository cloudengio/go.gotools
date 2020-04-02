package locate

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
	"sync"

	"cloudeng.io/errors"
	"golang.org/x/tools/go/packages"
)

type fileDesc struct {
	name     string
	ast      *ast.File
	pkg      *packages.Package
	comments ast.CommentMap
}

type loader struct {
	sync.Mutex
	// Indexed by packatge path.
	packages map[string]*packages.Package
	// Indexed by absolute filename.
	files map[string]fileDesc
	trace traceFunc
}

func newLoader(trace traceFunc) *loader {
	return &loader{
		packages: make(map[string]*packages.Package),
		files:    make(map[string]fileDesc),
		trace:    trace,
	}
}

func (ld *loader) loadPaths(paths []string) error {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedCompiledGoFiles,
		Tests:      false,
		BuildFlags: nil, // TODO: provide an option for buildflags.
	}
	if len(paths) == 0 {
		return nil
	}
	pkgs, err := packages.Load(cfg, paths...)
	if err != nil {
		return err
	}
	errs := &errors.M{}
	for _, pkg := range pkgs {
		if len(pkg.Name) == 0 {
			errs.Append(fmt.Errorf("failed to find: %v", pkg))
			continue
		}
		if pkg.IllTyped {
			errs.Append(fmt.Errorf("failed to type check: %v", pkg))
		}
	}
	if err := errs.Err(); err != nil {
		return err
	}
	ld.Lock()
	defer ld.Unlock()
	for _, pkg := range pkgs {
		ld.packages[pkg.PkgPath] = pkg
		for i, filename := range pkg.CompiledGoFiles {
			file := pkg.Syntax[i]
			ld.files[filename] = fileDesc{
				name:     filename,
				ast:      file,
				pkg:      pkg,
				comments: ast.NewCommentMap(pkg.Fset, file, file.Comments),
			}
			ld.trace("load: file: %v\n", filename)
		}
		ld.trace("load: package: %v\n", pkg.PkgPath)
	}
	return nil
}

func (ld *loader) lookupPackage(path string) *packages.Package {
	ld.Lock()
	defer ld.Unlock()
	if pkg := ld.packages[path]; pkg != nil {
		ld.trace("load: cached: %v\n", path)
		return pkg
	}
	return nil
}

func (ld *loader) lookupFile(filename string) (*ast.File, ast.CommentMap, *packages.Package) {
	ld.Lock()
	defer ld.Unlock()
	d := ld.files[filename]
	return d.ast, d.comments, d.pkg
}

func (ld *loader) astForPosition(path string, pos token.Pos) *ast.File {
	pkg := ld.lookupPackage(path)
	if pkg == nil {
		return nil
	}
	position := pkg.Fset.PositionFor(pos, false)
	ast, _, _ := ld.lookupFile(position.Filename)
	return ast
}

func (ld *loader) position(path string, pos token.Pos) token.Position {
	pkg := ld.lookupPackage(path)
	if pkg == nil {
		return token.Position{}
	}
	return pkg.Fset.PositionFor(pos, false)
}

func (ld *loader) walkFiles(fn func(filename string, pkg *packages.Package, cmap ast.CommentMap, file *ast.File)) {
	ld.Lock()
	files := make([]fileDesc, len(ld.files))
	i := 0
	for _, v := range ld.files {
		files[i] = v
		i++
	}
	ld.Unlock()
	sort.Slice(files, func(i, j int) bool {
		return files[i].name < files[j].name
	})
	for _, file := range files {
		fn(file.name, file.pkg, file.comments, file.ast)
	}
}
