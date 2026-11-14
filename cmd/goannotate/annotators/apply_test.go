// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package annotators

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cloudeng.io/text/edit"
)

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = orig
	}()

	runErr := fn()
	_ = w.Close()
	out, readErr := io.ReadAll(r)
	_ = r.Close()
	if readErr != nil {
		t.Fatal(readErr)
	}
	return string(out), runErr
}

func TestApplyEdits_OnlyOutputsChangedFiles(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	unchanged := filepath.Join(dir, "unchanged.go")
	noopEdit := filepath.Join(dir, "noop.go")
	changed := filepath.Join(dir, "changed.go")

	src := "package p\n\nfunc F() {}\n"
	for _, f := range []string{unchanged, noopEdit, changed} {
		if err := os.WriteFile(f, []byte(src), 0600); err != nil {
			t.Fatal(err)
		}
	}

	edits := map[string][]edit.Delta{
		unchanged: nil,
		noopEdit:  {edit.ReplaceString(0, len("package p"), "package p")},
		changed:   {edit.InsertString(0, "// Copyright 2026.\n\n")},
	}

	// In-place: only changed should be written and printed.
	out, err := captureStdout(t, func() error {
		return applyEdits(ctx, computeOutputs("", edits), edits)
	})
	if err != nil {
		t.Fatalf("applyEdits in-place failed: %v", err)
	}
	if got := strings.TrimSpace(out); got != changed {
		t.Errorf("in-place output = %q, want %q", got, changed)
	}

	// Writing to a new output dir creates all files, so all are reported.
	outDir := filepath.Join(t.TempDir(), "out")
	out, err = captureStdout(t, func() error {
		return applyEdits(ctx, computeOutputs(outDir, edits), edits)
	})
	if err != nil {
		t.Fatalf("applyEdits to new dir failed: %v", err)
	}
	for _, f := range []string{unchanged, noopEdit, changed} {
		if !strings.Contains(out, f) {
			t.Errorf("expected %q in first --write-dir output, got %q", f, out)
		}
	}

	// Re-running with the same --write-dir writes nothing and prints nothing.
	out, err = captureStdout(t, func() error {
		return applyEdits(ctx, computeOutputs(outDir, edits), edits)
	})
	if err != nil {
		t.Fatalf("applyEdits second run failed: %v", err)
	}
	if out != "" {
		t.Errorf("expected no output on second --write-dir run, got %q", out)
	}
}
