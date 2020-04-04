package locate

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"regexp"

	"cloudeng.io/go/locate/locateutil"
	"cloudeng.io/sync/errgroup"
	"golang.org/x/tools/go/packages"
)

func (t *T) findInterfaces(ctx context.Context, interfaces []string) error {
	group, ctx := errgroup.WithContext(ctx)
	group = errgroup.WithConcurrency(group, t.options.concurrency)
	for _, ifc := range interfaces {
		pkgPath, ifcRE, err := getPathAndRegexp(ifc)
		if err != nil {
			return err
		}
		group.GoContext(ctx, func() error {
			return t.findInterfacesInPackage(ctx, pkgPath, ifcRE)
		})
	}
	return group.Wait()
}

func (t *T) findInterfacesInPackage(ctx context.Context, pkgPath string, ifcRE *regexp.Regexp) error {
	pkg := t.loader.lookupPackage(pkgPath)
	if pkg == nil {
		return fmt.Errorf("locating interfaces: failed to lookup: %v", pkgPath)
	}
	found := 0
	checked := pkg.TypesInfo
	// Look in info.Defs for defined interfaces.
	for k, obj := range checked.Defs {
		if obj == nil || !k.IsExported() || !ifcRE.MatchString(k.Name) {
			continue
		}
		if _, ok := obj.(*types.TypeName); !ok {
			continue
		}
		ifcType := locateutil.IsInterfaceDefinition(pkg, obj)
		if ifcType == nil {
			continue
		}
		if el := ifcType.NumEmbeddeds(); el > 0 {
			// Make sure to include embedded interfaces. To do so, gather
			// the names of the embedded interfaces and iterate over the
			// typed checked definitions to locate them.
			names := map[string]bool{}
			for i := 0; i < el; i++ {
				et := ifcType.EmbeddedType(i)
				named, ok := et.(*types.Named)
				if !ok {
					continue
				}
				obj := named.Obj()
				epkg := obj.Pkg()
				if epath := epkg.Path(); epath != pkgPath {
					// Treat the external embedded interface as if it was
					// directly requested.
					re, _ := regexp.Compile(obj.Name() + "$")
					if err := t.findInterfacesInPackage(ctx, epath, re); err != nil {
						return err
					}
					continue
				}
				// Record the name of the locally defined embedded interfaces
				// and then look for them in the typed checked Defs.
				names[named.Obj().Name()] = true
			}
			for ek, eobj := range checked.Defs {
				if names[ek.Name] {
					ifcType := locateutil.IsInterfaceDefinition(pkg, eobj)
					if ifcType == nil {
						continue
					}
					t.addInterface(pkgPath, ek.Name, ek.Pos(), ifcType)
				}
			}
		}
		found++
		t.addInterface(pkgPath, k.Name, k.Pos(), ifcType)
	}
	if !t.options.ignoreMissingFunctionsEtc && found == 0 {
		return fmt.Errorf("failed to find any exported interfaces in %v for %s", pkgPath, ifcRE)
	}
	return nil
}

func (t *T) addInterface(path, name string, pos token.Pos, ifcType *types.Interface) {
	t.mu.Lock()
	defer t.mu.Unlock()
	position := t.loader.position(path, pos)
	fqn := path + "." + name
	filename := position.Filename
	ast, _, _ := t.loader.lookupFile(filename)
	t.interfaces[fqn] = interfaceDesc{
		path:     path,
		ifc:      ifcType,
		decl:     findInterfaceDecl(name, ast),
		position: position,
	}
	if t.interfaces[fqn].decl == nil {
		fmt.Printf("Failed to locate source code location for package %v, name %v, interface %s @ %v\n", path, name, ifcType.String(), position)
		panic("internal error")
	}
	t.dirty[filename] |= HasInterface
	t.trace("interface: %v @ %v\n", fqn, position)
}

func findInterfaceDecl(name string, file *ast.File) *ast.TypeSpec {
	for _, d := range file.Decls {
		d, ok := d.(*ast.GenDecl)
		if !ok || d.Tok != token.TYPE {
			continue
		}
		for _, spec := range d.Specs {
			typSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if typSpec.Name.Name == name {
				return typSpec
			}
		}
	}
	return nil
}

type interfaceDesc struct {
	path     string
	ifc      *types.Interface
	decl     *ast.TypeSpec
	position token.Position
}

// WalkInterfaces calls the supplied function for each interface location,
// ordered by filename and then position within file.
// The function is called with the packages.Package and ast for the file
// that contains the interface, as well as the type and declaration of the
// interface.
func (t *T) WalkInterfaces(fn func(
	fullname string,
	pkg *packages.Package,
	file *ast.File,
	decl *ast.TypeSpec,
	ifc *types.Interface)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	sorted := make([]sortByPos, len(t.interfaces))
	i := 0
	for k, v := range t.interfaces {
		sorted[i] = sortByPos{
			name:    k,
			pos:     v.position,
			payload: v,
		}
		i++
	}
	sorter(sorted)
	for _, loc := range sorted {
		ifc := loc.payload.(interfaceDesc)
		file, _, pkg := t.loader.lookupFile(ifc.position.Filename)
		fn(loc.name, pkg, file, ifc.decl, ifc.ifc)
	}
}
