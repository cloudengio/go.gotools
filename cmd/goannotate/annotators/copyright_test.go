package annotators_test

import (
	"path/filepath"
	"testing"

	"cloudeng.io/go/cmd/goannotate/annotators"
	"golang.org/x/net/context"
)

var expectedPersonalApache = []diffReport{
	{"cloudeng.go",
		`1c1
< // Copyright 2020 cloudeng llc. All rights reserved.
---
> // Copyright 2020 Cosmos Nicolaou. All rights reserved.
`},
	{"empty.go", `0a1,4
> // Copyright 2020 Cosmos Nicolaou. All rights reserved.
> // Use of this source code is governed by the Apache-2.0
> // license that can be found in the LICENSE file.
> 
`},
	{"packagecomment.go", `0a1,4
> // Copyright 2020 Cosmos Nicolaou. All rights reserved.
> // Use of this source code is governed by the Apache-2.0
> // license that can be found in the LICENSE file.
> 
`},
	{"personal.go", ""},
}

func TestCopyright(t *testing.T) {
	ctx := context.Background()
	tmpdir, cleanup := setup(t)
	defer cleanup()
	err := annotators.Lookup("personal-apache").Do(ctx, tmpdir, []string{here + "copyright"})
	if err != nil {
		t.Errorf("Do: %v", err)
	}
	original, copies := list(t, filepath.Join("testdata", "copyright")), list(t, tmpdir)
	diffs := diffAll(t, original, copies)
	compare(t, diffs, expectedPersonalApache)
}
