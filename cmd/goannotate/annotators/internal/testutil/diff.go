// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package testutil

import (
	"os/exec"
	"path/filepath"
	"testing"

	"cloudeng.io/errors"
)

func DiffOneFile(a, b string) string {
	cmd := exec.Command("diff", a, b)
	// Ignore return code since differences are expected.
	out, _ := cmd.CombinedOutput()
	return string(out)
}

type DiffReport struct {
	Name string
	Diff string
}

func DiffMultipleFiles(t *testing.T, a, b []string) []DiffReport {
	if got, want := len(a), len(b); got != want {
		t.Errorf("%v: got %v, want %v", errors.Caller(2, 1), got, want)
		t.Logf("got: %v\n", a)
		t.Logf("want: %v\n", b)
		return nil
	}
	var diffs []DiffReport
	for i := range a {
		diffs = append(diffs, DiffReport{
			Name: filepath.Base(a[i]),
			Diff: DiffOneFile(a[i], b[i]),
		})
	}
	return diffs
}

func CompareDiffReports(t *testing.T, a, b []DiffReport) {
	if got, want := len(a), len(b); got != want {
		t.Errorf("%v: got %v, want %v", errors.Caller(2, 1), got, want)
		t.Logf("got: %v\n", a)
		t.Logf("want: %v\n", b)
		return
	}
	for i := range a {
		if got, want := a[i].Name, b[i].Name; got != want {
			t.Errorf("%v: got %v, want %v", errors.Caller(2, 1), got, want)
			return
		}
		if got, want := a[i].Diff, b[i].Diff; got != want {
			t.Errorf("%v: %v: got %v, want %v", errors.Caller(2, 1), a[i].Name, got, want)
			return
		}
	}
}
