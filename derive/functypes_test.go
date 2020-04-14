package derive_test

import (
	"context"
	"go/ast"
	"go/types"
	"strings"
	"testing"

	"cloudeng.io/errors"
	"cloudeng.io/go/derive"
	"cloudeng.io/go/locate"
	"golang.org/x/tools/go/packages"
)

const testdata = "cloudeng.io/go/derive/testdata/examples"

func TestFormatting(t *testing.T) {
	ctx := context.Background()
	locator := locate.New()
	locator.AddFunctions(testdata)
	if err := locator.Do(ctx); err != nil {
		t.Errorf("locate.Do: %v", err)
	}

	type formatted struct {
		spec, arg string
	}

	parameters, parametersContext, results := []formatted{}, []formatted{}, []formatted{}

	locator.WalkFunctions(func(name string, pkg *packages.Package, file *ast.File, fn *types.Func, decl *ast.FuncDecl, implemented []string) {
		signature := fn.Type().(*types.Signature)
		spec, args := derive.ArgsForParams(signature)
		parameters = append(parameters, formatted{spec, strings.Join(args, ", ")})
		if ctxArg, ok := derive.HasContext(signature); ok {
			// ignore context.Context at positiohn 0.
			spec, args = derive.ArgsForParams(signature, 0)
			args = append([]string{ctxArg}, args...)
		} else {
			spec, args = derive.ArgsForParams(signature)
		}
		parametersContext = append(parametersContext,
			formatted{spec, strings.Join(args, ", ")})
		spec, args = derive.ArgsForResults(signature)
		results = append(results, formatted{spec, strings.Join(args, ", ")})
	})

	cmp := func(got, want []formatted) {
		if got, want := len(got), len(want); got != want {
			t.Errorf("%v: got %v, want %v", errors.Caller(2, 1), got, want)
			return
		}
		for i := range got {
			if got, want := got[i], want[i]; got != want {
				t.Errorf("%v: %d: got %#v, want %#v", errors.Caller(2, 1), i, got, want)
			}
		}
	}

	j := func(spec, arg string) formatted {
		return formatted{spec, arg}
	}

	expectedParameters := []formatted{
		j("", ""),
		j("a=%d, b=%d, c=%d, d=%f, e=%f, f=%f, g=%f, h=%f", "a, b, c, d, e, f, g, h"),
		j("a=%t, b=%.10s..., c=%c, d=%d, e=%02x", "a, b, c, d, e"),
		j("a=%v", "a"),
		j("a=%p, b=%p, c=%p", "a, b, c"),
		j("a=?, b=?", ""),
		j("a=%s.10s..., b=%s.10s...", "a.String(), b.String()"),
		j("_=?, _=?", ""),
		j("a=%d, b=%.10s...", "a, b"),
		j("a[:%d]=..., b[:%d]=..., c[:%d]=...", "len(a), len(b), len(c)"),
		j("a=%d, ...b[:%d]=...", "a, len(b)"),
		j("a=%d, b=%d", "a, b"),
		j("ctx=?", ""),
		j("ctx=?, a=%d", "a"),
		j("ctx=?, _=?, c=%t", "c"),
	}
	expectedResults := []formatted{
		j("", ""),
		j("ar=%d, br=%d, cr=%d, dr=%f, er=%f, fr=%f, gr=%f, hr=%f", "ar, br, cr, dr, er, fr, gr, hr"),
		j("ar=%t, br=%.10s..., cr=%c, dr=%d, er=%02x", "ar, br, cr, dr, er"),
		j("err=%v", "err"),
		j("", ""),
		j("_=?, _=?", ""),
		j("", ""),
		j("", ""),
		j("", ""),
		j("", ""),
		j("", ""),
		j("ar=%.10s..., br=%.10s...", "ar, br"),
		j("", ""),
		j("", ""),
		j("", ""),
	}

	expectedParametersContext := make([]formatted, len(expectedParameters))
	copy(expectedParametersContext, expectedParameters)
	expectedParametersContext[12] = j("", "ctx")
	expectedParametersContext[13] = j("a=%d", "ctx, a")
	expectedParametersContext[14] = j("_=?, c=%t", "ctx, c")

	cmp(parameters, expectedParameters)
	cmp(parametersContext, expectedParametersContext)
	cmp(results, expectedResults)

}
