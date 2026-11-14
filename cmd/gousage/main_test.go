// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestExtractCopyrightHeader(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "copyright and license block",
			src: "// Copyright 2020 cloudeng llc. All rights reserved.\n" +
				"// Use of this source code is governed by the Apache-2.0\n" +
				"// license that can be found in the LICENSE file.\n\n" +
				"// Usage of foo:\npackage main\n",
			want: "// Copyright 2020 cloudeng llc. All rights reserved.\n" +
				"// Use of this source code is governed by the Apache-2.0\n" +
				"// license that can be found in the LICENSE file.",
		},
		{
			name: "separate copyright and license blocks",
			src: "// Copyright 2020 cloudeng llc.\n\n" +
				"// Licensed under the Apache License.\n\n" +
				"// Usage of foo:\npackage main\n",
			want: "// Copyright 2020 cloudeng llc.\n\n" +
				"// Licensed under the Apache License.",
		},
		{
			name: "spdx header",
			src:  "// SPDX-License-Identifier: Apache-2.0\n\n// Usage of foo:\npackage main\n",
			want: "// SPDX-License-Identifier: Apache-2.0",
		},
		{
			name: "no header",
			src:  "// Usage of foo:\npackage main\n",
			want: "",
		},
		{
			name: "usage comment with blank line is not header",
			src:  "// Usage of foo:\n\npackage main\n",
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractCopyrightHeader([]byte(tc.src)); got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q", got, tc.want)
			}
		})
	}
}

func TestWriteGo_PreservesHeaderAndReportsOnlyChanged(t *testing.T) {
	origOverwrite := cmdFlags.Overwrite
	t.Cleanup(func() {
		cmdFlags.Overwrite = origOverwrite
	})

	dir := t.TempDir()
	file := filepath.Join(dir, "cmdusage.go")
	header := "// Copyright 2020 cloudeng llc. All rights reserved.\n" +
		"// Use of this source code is governed by the Apache-2.0\n" +
		"// license that can be found in the LICENSE file."
	usage1 := "// Usage of tool:\n//\n//\t-a\npackage main\n"
	usage2 := "// Usage of tool:\n//\n//\t-b\npackage main\n"

	// 1. New file is written and reported.
	cmdFlags.Overwrite = false
	out, err := captureStdout(t, func() error { return writeGo(file, usage1) })
	if err != nil {
		t.Fatalf("new writeGo failed: %v", err)
	}
	if !strings.Contains(out, "updated: "+file) {
		t.Errorf("expected updated output, got %q", out)
	}

	// 2. Seed file with a copyright header and matching usage.
	withHeader := header + "\n\n" + usage1
	if err := os.WriteFile(file, []byte(withHeader), 0600); err != nil {
		t.Fatal(err)
	}

	// 3. Unchanged usage with existing header succeeds without --overwrite and prints nothing.
	cmdFlags.Overwrite = false
	out, err = captureStdout(t, func() error { return writeGo(file, usage1) })
	if err != nil {
		t.Fatalf("unchanged writeGo failed: %v", err)
	}
	if out != "" {
		t.Errorf("expected no output for unchanged file, got %q", out)
	}
	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != withHeader {
		t.Errorf("file modified unexpectedly:\ngot:\n%s\nwant:\n%s", got, withHeader)
	}

	// 4. Changed usage without --overwrite returns an error and preserves existing content.
	out, err = captureStdout(t, func() error { return writeGo(file, usage2) })
	if err == nil {
		t.Fatalf("expected overwrite error, got nil")
	}
	if out != "" {
		t.Errorf("expected no output on error, got %q", out)
	}

	// 5. Changed usage with --overwrite preserves the copyright header and reports the update.
	cmdFlags.Overwrite = true
	out, err = captureStdout(t, func() error { return writeGo(file, usage2) })
	if err != nil {
		t.Fatalf("overwrite writeGo failed: %v", err)
	}
	if strings.TrimSpace(out) != "updated: "+file {
		t.Errorf("got output %q, want %q", out, "updated: "+file)
	}
	got, err = os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	want := header + "\n\n" + usage2
	if string(got) != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}
