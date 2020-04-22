// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package annotators_test

import (
	"context"
	"path/filepath"
	"testing"

	"cloudeng.io/go/cmd/goannotate/annotators"
	"cloudeng.io/go/cmd/goannotate/annotators/internal/testutil"
)

var expectedRmLegacycall = []testutil.DiffReport{
	{Name: "legacy.go", Diff: `8d7
< 	defer apilog.LogCallfLegacy(nil, "buf=%v...", buf)(nil, "") // gologcop: DO NOT EDIT, MUST BE FIRST STATEMENT
13d11
< 	defer apilog.LogCallfLegacy(nil, "n=%v...", n)(nil, "") // gologcop: DO NOT EDIT, MUST BE FIRST STATEMENT
`},
}

var expectedRmNoDeferLegacycall = []testutil.DiffReport{
	{Name: "legacy.go", Diff: `18d17
< 	apilog.LogCallfLegacy(nil, "n=%v...", n)(nil, "") // gologcop: DO NOT EDIT, MUST BE FIRST STATEMENT
`},
}

func TestRmLogCall(t *testing.T) {
	ctx := context.Background()
	tmpdir, cleanup := testutil.SetupAnnotators(t)
	defer cleanup()
	err := annotators.Lookup("rmlegacy").Do(ctx, tmpdir, []string{here + "impl"})
	if err != nil {
		t.Errorf("Do: %v", err)
	}
	original := []string{filepath.Join("testdata", "impl", "legacy.go")}
	copies := list(t, tmpdir)
	diffs := testutil.DiffMultipleFiles(t, original, copies)
	testutil.CompareDiffReports(t, diffs, expectedRmLegacycall)

	tmpdir, cleanup = testutil.SetupAnnotators(t)
	defer cleanup()
	err = annotators.Lookup("rmlegacy-nodefer").Do(ctx, tmpdir, []string{here + "impl"})
	if err != nil {
		t.Errorf("Do: %v", err)
	}
	copies = list(t, tmpdir)
	diffs = testutil.DiffMultipleFiles(t, original, copies)
	testutil.CompareDiffReports(t, diffs, expectedRmNoDeferLegacycall)
}
