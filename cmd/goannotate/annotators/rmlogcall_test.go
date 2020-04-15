package annotators_test

import (
	"context"
	"path/filepath"
	"testing"

	"cloudeng.io/go/cmd/goannotate/annotators"
)

var expectedRmLegacycall = []diffReport{
	{"legacy.go", `8d7
< 	defer apilog.LogCallfLegacy(nil, "buf=%v...", buf)(nil, "") // gologcop: DO NOT EDIT, MUST BE FIRST STATEMENT
13d11
< 	defer apilog.LogCallfLegacy(nil, "n=%v...", n)(nil, "") // gologcop: DO NOT EDIT, MUST BE FIRST STATEMENT
`},
}

var expectedRmNoDeferLegacycall = []diffReport{
	{"legacy.go", `18d17
< 	apilog.LogCallfLegacy(nil, "n=%v...", n)(nil, "") // gologcop: DO NOT EDIT, MUST BE FIRST STATEMENT
`},
}

func TestRmLogCall(t *testing.T) {
	ctx := context.Background()
	tmpdir, cleanup := setup(t)
	defer cleanup()
	err := annotators.Lookup("rmlegacy").Do(ctx, tmpdir, []string{here + "impl"})
	if err != nil {
		t.Errorf("Do: %v", err)
	}
	original := []string{filepath.Join("testdata", "impl", "legacy.go")}
	copies := list(t, tmpdir)
	diffs := diffAll(t, original, copies)
	compare(t, diffs, expectedRmLegacycall)

	tmpdir, cleanup = setup(t)
	defer cleanup()
	err = annotators.Lookup("rmlegacy-nodefer").Do(ctx, tmpdir, []string{here + "impl"})
	if err != nil {
		t.Errorf("Do: %v", err)
	}
	copies = list(t, tmpdir)
	diffs = diffAll(t, original, copies)
	compare(t, diffs, expectedRmNoDeferLegacycall)
}
