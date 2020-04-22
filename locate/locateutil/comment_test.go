// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package locateutil_test

import (
	"go/ast"
	"regexp"
	"strings"
	"testing"

	"cloudeng.io/go/locate/locateutil"
	"golang.org/x/tools/go/packages"
)

func TestComments(t *testing.T) {
	pkgs, err := packages.Load(packagesConfig,
		"cloudeng.io/go/locate/testdata/functions",
	)
	if err != nil {
		t.Errorf("pkg.Load: %v", err)
	}
	pkg := pkgs[0]
	fns := locateutil.Functions(pkg, regexp.MustCompile("HasCall$|HasDefer$"), true)
	if got, want := len(fns), 2; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	cases := []struct {
		deferred    bool
		first, last string
	}{
		{false, "functions.go:12:2", "functions.go:16:19"}, // functinns.HasCall
		{true, "functions.go:20:2", "functions.go:23:15"},  // functions.HasDefer
	}

	for i, fn := range fns {
		cm := ast.NewCommentMap(fn.Package.Fset, fn.File, fn.File.Comments)
		calls := locateutil.FunctionCalls(fn.Decl, regexp.MustCompile("ioutil.ReadFile"), cases[i].deferred)
		if got, want := len(calls), 1; got != want {
			t.Errorf("got %v, want %v", got, want)
			continue
		}
		cg := cm[calls[0]]
		if cg == nil {
			t.Errorf("%v: no comments found", i)
			continue
		}
		for _, text := range []string{"", "not there"} {
			if got, want := locateutil.CommentGroupsContain(cg, text), false; got != want {
				t.Errorf("got %v, want %v", got, want)
			}
		}
		for _, text := range []string{"before", "same line", "after"} {
			if got, want := locateutil.CommentGroupsContain(cg, text), true; got != want {
				t.Errorf("got %v, want %v", got, want)
			}
		}
		first, last := locateutil.CommentGroupBounds(cg)
		if got, want := fn.Package.Fset.Position(first).String(), cases[i].first; !strings.HasSuffix(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := fn.Package.Fset.Position(last).String(), cases[i].last; !strings.HasSuffix(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}
