package annotators_test

import (
	"context"
	"testing"

	"cloudeng.io/go/cmd/goannotate/annotators"
)

var expectedAddcall = []diffReport{
	{"empty.go", `2a3,4
> import "cloudeng.io/go/cmd/goannotate/annotators/testdata/apilog"
> 
5a8
> 	defer apilog.LogCallf(nil, "cloudeng.io/go/cmd/goannotate/annotators/testdata/impl.Write", "impl/empty.go:5", "buf[:%d]=...", len(buf))(nil, "_=?") // DO NOT EDIT, AUTO GENERATED BY cloudeng.io/go/cmd/goannotate/annotators.AddLogCall#add
9a13
> 	defer apilog.LogCallf(nil, "cloudeng.io/go/cmd/goannotate/annotators/testdata/impl.APIEmpty", "impl/empty.go:9", "n=%d", n)(nil, "_=?") // DO NOT EDIT, AUTO GENERATED BY cloudeng.io/go/cmd/goannotate/annotators.AddLogCall#add
`},
	{"existing.go", `17a18
> 	defer apilog.LogCallf(nil, "cloudeng.io/go/cmd/goannotate/annotators/testdata/impl.APINew", "impl/existing.go:17", "n=%d", n)(nil, "_=?") // DO NOT EDIT, AUTO GENERATED BY cloudeng.io/go/cmd/goannotate/annotators.AddLogCall#add
`},
	{"legacy.go", `8c8,9
< 	defer apilog.LogCallfLegacy(nil, "buf=%v...", buf)(nil, "") // gologcop: DO NOT EDIT, MUST BE FIRST STATEMENT
---
> 	defer apilog.LogCallf(nil, "cloudeng.io/go/cmd/goannotate/annotators/testdata/impl.Write", "impl/legacy.go:7", "buf[:%d]=...", len(buf))(nil, "_=?") // DO NOT EDIT, AUTO GENERATED BY cloudeng.io/go/cmd/goannotate/annotators.AddLogCall#add
> 	defer apilog.LogCallfLegacy(nil, "buf=%v...", buf)(nil, "")                                                                                          // gologcop: DO NOT EDIT, MUST BE FIRST STATEMENT
13c14,15
< 	defer apilog.LogCallfLegacy(nil, "n=%v...", n)(nil, "") // gologcop: DO NOT EDIT, MUST BE FIRST STATEMENT
---
> 	defer apilog.LogCallf(nil, "cloudeng.io/go/cmd/goannotate/annotators/testdata/impl.APILegacy", "impl/legacy.go:12", "n=%d", n)(nil, "_=?") // DO NOT EDIT, AUTO GENERATED BY cloudeng.io/go/cmd/goannotate/annotators.AddLogCall#add
> 	defer apilog.LogCallfLegacy(nil, "n=%v...", n)(nil, "")                                                                                    // gologcop: DO NOT EDIT, MUST BE FIRST STATEMENT
18c20,21
< 	apilog.LogCallfLegacy(nil, "n=%v...", n)(nil, "") // gologcop: DO NOT EDIT, MUST BE FIRST STATEMENT
---
> 	defer apilog.LogCallf(nil, "cloudeng.io/go/cmd/goannotate/annotators/testdata/impl.APILegacyNonDefer", "impl/legacy.go:17", "n=%d", n)(nil, "_=?") // DO NOT EDIT, AUTO GENERATED BY cloudeng.io/go/cmd/goannotate/annotators.AddLogCall#add
> 	apilog.LogCallfLegacy(nil, "n=%v...", n)(nil, "")                                                                                                  // gologcop: DO NOT EDIT, MUST BE FIRST STATEMENT
`},
}

func TestAddLogCall(t *testing.T) {
	ctx := context.Background()
	tmpdir, cleanup := setup(t)
	defer cleanup()
	err := annotators.Lookup("add").Do(ctx, tmpdir, []string{here + "impl"})
	if err != nil {
		t.Errorf("Do: %v", err)
	}
	original, copies := list(t, "testdata/impl/"), list(t, tmpdir)
	diffs := diffAll(t, original, copies)
	compare(t, diffs, expectedAddcall)
}