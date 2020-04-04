package locateutil

import 	"go/types"

// IsAbstract returns true if the functional declaration is abstract.
func IsAbstract(fn *types.Func) bool {
	if fn == nil {
		return false
	}
	sig, _ := fn.Type().(*types.Signature)
	if sig.Recv() == nil {
		return false
	}
	return isAbstract(sig.Recv().Type())
}

func isAbstract(typ types.Type) bool{
	switch v := typ.(type) {
	case *types.Interface:
		return true
	case *types.Named:
		return isAbstract(v.Underlying())
	}
	return false
}