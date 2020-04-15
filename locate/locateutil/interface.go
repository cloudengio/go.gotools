package locateutil

import (
	"go/types"

	"golang.org/x/tools/go/packages"
)

// IsAbstract returns true if the function declaration is abstract.
func IsAbstract(fn *types.Func) bool {
	if fn == nil {
		return false
	}
	sig, _ := fn.Type().(*types.Signature)
	if sig.Recv() == nil {
		return false
	}
	return InterfaceType(sig.Recv().Type()) != nil
}

// InterfaceType returns the underlying *types.Interface if typ represents
// an interface or nil otherwise.
func InterfaceType(typ types.Type) *types.Interface {
	switch v := typ.(type) {
	case *types.Interface:
		return v
	case *types.Named:
		return InterfaceType(v.Underlying())
	}
	return nil
}

// IsInterfaceDefinition returns the interface type that the suplied object
// defines in the specified package, if any. This specifically excludes
// embedded types which are defined in other packages and anonymous interfaces.
func IsInterfaceDefinition(pkg *packages.Package, obj types.Object) *types.Interface {
	if obj == nil {
		return nil
	}
	if _, ok := obj.(*types.TypeName); !ok {
		return nil
	}
	typ := obj.Type()
	ifc := InterfaceType(typ)
	if ifc == nil {
		return nil
	}
	if named, ok := typ.(*types.Named); ok {
		obj := named.Obj()
		if obj == nil || obj.Pkg() == nil {
			return nil
		}
		if obj.Pkg().Path() == pkg.PkgPath {
			return ifc
		}
	}
	return nil
}
