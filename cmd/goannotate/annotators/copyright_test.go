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

var expectedPersonalApache = []testutil.DiffReport{
	{Name: "cloudeng.go", Diff: `1c1
< // Copyright 2020 cloudeng llc. All rights reserved.
---
> // Copyright 2020 Cosmos Nicolaou. All rights reserved.
`},
	{Name: "empty.go", Diff: `0a1,4
> // Copyright 2020 Cosmos Nicolaou. All rights reserved.
> // Use of this source code is governed by the Apache-2.0
> // license that can be found in the LICENSE file.
> 
`},
	{Name: "packagecomment.go", Diff: `0a1,4
> // Copyright 2020 Cosmos Nicolaou. All rights reserved.
> // Use of this source code is governed by the Apache-2.0
> // license that can be found in the LICENSE file.
> 
`},
	{Name: "personal.go", Diff: ""},
}

func TestCopyright(t *testing.T) {
	ctx := context.Background()
	tmpdir, cleanup := testutil.SetupAnnotators(t)
	defer cleanup()
	err := annotators.Lookup("personal-apache").Do(ctx, tmpdir, []string{here + "copyright"})
	if err != nil {
		t.Errorf("Do: %v", err)
	}
	original, copies := list(t, filepath.Join("testdata", "copyright")), list(t, tmpdir)
	diffs := testutil.DiffMultipleFiles(t, original, copies)
	testutil.CompareDiffReports(t, diffs, expectedPersonalApache)
}
