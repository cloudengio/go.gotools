// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package derive

import (
	"go/types"
	"strings"
)

// ContextType is the standard go context type.
const ContextType = "context.Context"

func implementsStringer(typ types.Type) bool {
	ms := types.NewMethodSet(typ)
	if ms.Len() == 0 {
		if _, ok := typ.(*types.Pointer); !ok {
			ms = types.NewMethodSet(types.NewPointer(typ))
		}
	}
	for i := 0; i < ms.Len(); i++ {
		obj := ms.At(i).Obj()
		if obj.Name() != "String" || !obj.Exported() {
			continue
		}
		sig := obj.Type().(*types.Signature)
		if sig.Params().Len() != 0 || sig.Results().Len() != 1 {
			continue
		}
		return sig.Results().At(0).Type().String() == "string"
	}
	return false
}

func specForBasicType(id string, bt *types.Basic) (string, string) {
	switch bt.Name() {
	case "rune":
		return id + "=%c", id
	case "byte":
		return id + "=%02x", id
	}
	if bt.Kind() == types.UnsafePointer {
		return id + "=%p", id
	}
	info := bt.Info()
	switch {
	case (info & types.IsString) != 0:
		return id + "=%.10s...", id
	case (info & types.IsBoolean) != 0:
		return id + "=%t", id
	case (info & types.IsInteger) != 0:
		return id + "=%d", id
	case (info&types.IsFloat != 0) || (info&types.IsComplex != 0):
		return id + "=%f", id
	}
	return id + "=?", ""
}

func formatForVar(name string, typ types.Type) (string, string) {
	if implementsStringer(typ) {
		return name + "=%s.10s...", name + ".String()"
	}
	switch vt := typ.(type) {
	case *types.Basic:
		return specForBasicType(name, vt)
	case *types.Pointer:
		return name + "=%p", name
	case *types.Named:
		underlying := vt.Underlying()
		if vt.Obj().Name() == "error" {
			if _, ok := underlying.(*types.Interface); ok {
				return name + "=%v", name
			}
		}
		return formatForVar(name, vt.Underlying())
	case *types.Slice, *types.Map:
		return name + "[:%d]=...", "len(" + name + ")"
	}
	return name + "=?", ""
}

// FormatForVar determines an appropriate format spec and argument for a single
// function argument or result. The format spec is intended to be passed
// to a fmt style logging function. It takes care to ensure that the log
// output is bounded as follows:
//  1. strings and types that implement stringer are printed as %.10s
//  2. slices and maps have only their length printed
//  3. errors are printed as %v with no other restrictions
//  4. runes are printed as %c, bytes as %02x and pointers are as %02x
//  5. for all other types, only the name of the variable is printed
func FormatForVar(v *types.Var) (string, string) {
	name := v.Name()
	if len(name) == 0 || name == "_" {
		return "_=?", ""
	}
	return formatForVar(name, v.Type())
}

// formatAndArgs returns the format string and arguments for the supplied
// tuple. The number of arguments may be less than the length of the
// tuple when it contains parameters that are named as '_' or when
// results are unamed.
func formatAndArgs(tuple *types.Tuple, variadic bool, ignore map[int]bool) (format string, arguments []string) {
	for i := 0; i < tuple.Len(); i++ {
		if ignore != nil && ignore[i] {
			continue
		}
		spec, arg := FormatForVar(tuple.At(i))
		if variadic && (i == tuple.Len()-1) {
			spec = "..." + spec
		}
		format += spec + ", "
		if len(arg) > 0 {
			arguments = append(arguments, arg)
		}
	}
	format = strings.TrimSuffix(format, ", ")
	return
}

// ArgsForParams returns the format and arguments to use to log the
// function's arguments. The option ignoreAtPosition arguments specify
// that those positions should be ignored altogether. This is useful for
// handling context.Context like arguments which need often need to be
// handled separately.
func ArgsForParams(signature *types.Signature, ignoreAtPosition ...int) (format string, arguments []string) {
	pt := map[int]bool{}
	for _, v := range ignoreAtPosition {
		pt[v] = true
	}
	return formatAndArgs(signature.Params(), signature.Variadic(), pt)
}

// ArgsForResults returns the format and arguments to use to log the
// function's results.
func ArgsForResults(signature *types.Signature) (format string, arguments []string) {
	return formatAndArgs(signature.Results(), false, nil)
}

// ParamAt returns the name and type of the parameter at pos. It returns
// false if no such parameter exists.
func ParamAt(signature *types.Signature, pos int) (varName, typeName string, ok bool) {
	params := signature.Params()
	if params.Len() == 0 || params.Len() < pos {
		return
	}
	v := params.At(pos)
	return v.Name(), v.Type().String(), true
}

// HasContext returns true and the name of the first parameter to the function
// if that first parameter is context.Context.
func HasContext(signature *types.Signature) (string, bool) {
	return HasCustomContext(signature, ContextType)
}

// HasCustomContext returns true and the name of the first parameter to the function
// if that first parameter is the specified customContext.
func HasCustomContext(signature *types.Signature, customContext string) (string, bool) {
	v, t, ok := ParamAt(signature, 0)
	if !ok || len(t) == 0 {
		return "", false
	}
	if t == customContext || (t[0] == '*' && t[1:] == customContext) {
		return v, true
	}
	return "", false
}
