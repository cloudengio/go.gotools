package locateutil_test

import (
	"go/ast"
	"regexp"
	"strings"
	"testing"

	"cloudeng.io/go/locate/locateutil"
	"golang.org/x/tools/go/packages"
)

func TestFunction(t *testing.T) {
	pkgs, err := packages.Load(packagesConfig,
		"cloudeng.io/go/locate/testdata/data",
		"cloudeng.io/go/locate/testdata/impl",
	)
	if err != nil {
		t.Errorf("pkg.Load: %v", err)
	}
	var fns []locateutil.FuncDesc
	// Top-level functions only.
	for _, pkg := range pkgs {
		fns = append(fns, locateutil.Functions(pkg, regexp.MustCompile(".*"), true)...)
	}
	if got, want := len(fns), 2; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	for _, fn := range fns {
		if got, want := locateutil.IsAbstract(fn.Type), false; got != want {
			t.Errorf("%v: got %v, want %v", fn.Type.Name(), got, want)
		}
	}

}

func TestFunctiosAndMethods(t *testing.T) {
	pkgs, err := packages.Load(packagesConfig,
		"cloudeng.io/go/locate/testdata/data",
		"cloudeng.io/go/locate/testdata/impl",
	)
	if err != nil {
		t.Errorf("pkg.Load: %v", err)
	}

	var fns []locateutil.FuncDesc
	// Top-level functions and methods.
	fns = nil
	for _, pkg := range pkgs {
		fns = append(fns, locateutil.Functions(pkg, regexp.MustCompile(".*"), false)...)
	}
	if got, want := len(fns), 16; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	abstract := []bool{
		false, false, false, true,
		true, true, true, false,
		false, false, false, false,
		false, false, true, true,
	}
	for i, fn := range fns {
		if got, want := locateutil.IsAbstract(fn.Type), abstract[i]; got != want {
			t.Errorf("%v: %v: got %v, want %v", i, fn.Type.FullName(), got, want)
		}
	}
	for i, tc := range []struct {
		name, pos string
	}{
		{"Fn1", "data/functions.go:7:6"},
		{"Fn1", "data/functions.go:13:15"},
		{"Fn2", "data/functions_more.go:3:6"},
		{"M1", "interfaces.go:4:2"},
		{"M2", "interfaces.go:5:2"},
		{"M1", "interfaces.go:9:2"},
		{"M3", "interfaces.go:13:2"},
		{"M1", "impl/impls.go:5:17"},
		{"M2", "impl/impls.go:9:17"},
		{"M3", "impl/impls.go:15:17"},
		{"M1", "impl/impls.go:22:18"},
		{"M2", "impl/impls.go:26:18"},
		{"M3", "impl/impls.go:30:18"},
		{"M1", "impl/impls.go:37:17"},
		{"M1", "impl/impls.go:44:3"},
		{"M2", "impl/impls.go:45:3"},
	} {
		if got, want := fns[i].Type.Name(), tc.name; got != want {
			t.Errorf("%v: got %v, want %v", i, got, want)
		}
		if got, want := fns[i].Position.String(), tc.pos; !strings.HasSuffix(got, want) {
			t.Errorf("%v: got %v doesn't suffix %v", i, got, want)
		}
		if fns[i].File == nil {
			t.Errorf("missing ast for %v", i)
		}
		if fns[i].Decl == nil && !fns[i].Abstract {
			t.Errorf("missing declaration for %v", i)
		}
	}
}

func TestFunctionCalls(t *testing.T) {
	pkgs, err := packages.Load(packagesConfig,
		"cloudeng.io/go/locate/testdata/functions",
	)
	if err != nil {
		t.Errorf("pkg.Load: %v", err)
	}
	pkg := pkgs[0]
	fns := locateutil.Functions(pkg, regexp.MustCompile(".*"), true)

	expected := []struct {
		statements                int
		hasFuncCall, hasDeferCall int
		hasComment                bool
	}{
		{0, 0, 0, false}, // functions.Empty
		{1, 1, 0, true},  // functions.Hascall
		{1, 0, 1, true},  // functions.HasDefer
		{1, 0, 0, false}, // functions.HasOther
		{1, 0, 0, false}, // functions.HasOtherdefer
		{4, 0, 0, false}, // functions.Expressions
	}
	for i, fn := range fns {
		if got, want := locateutil.FunctionStatements(fn.Decl), expected[i].statements; got != want {
			t.Errorf("%v: got %v, want %v", i, got, want)
		}
		nodes := locateutil.FunctionCalls(fn.Decl, regexp.MustCompile("ioutil.ReadFile"), false)
		if got, want := len(nodes), expected[i].hasFuncCall; got != want {
			t.Errorf("%v: got %v, want %v", i, got, want)
		}
		nodes = locateutil.FunctionCalls(fn.Decl, regexp.MustCompile("ioutil.Read.*"), true)
		if got, want := len(nodes), expected[i].hasDeferCall; got != want {
			t.Errorf("%v: got %v, want %v", i, got, want)
		}
		cmap := ast.NewCommentMap(fn.Package.Fset, fn.File, fn.File.Comments)
		if got, want := locateutil.FunctionHasComment(fn.Decl, cmap, "nologcall:"), expected[i].hasComment; got != want {
			t.Errorf("%v: got %v, want %v", i, got, want)
		}
	}
}
