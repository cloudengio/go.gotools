// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

// TestFallbackPackageDocFormatting tests that the declarations returned by
// fallbackPackageDoc are formatted using the FileSet that they were parsed
// into, rather than that of the package, in which their positions refer to
// unrelated locations.
func TestFallbackPackageDocFormatting(t *testing.T) {
	src := `package p

// R is a struct.
type R struct {
	ID  uint64
	Seq uint64
	// From is a doc comment
	// on two lines.
	From  uint64
	Frame []byte
}

// Q is a struct.
type Q struct {
	// Wait is a doc comment.
	Wait int
}

const (
	// A is a constant.
	A = 1
	// B is a constant.
	B = 2
)

// F is a function.
func F(a int) int {
	return a
}
`
	filename := filepath.Join(t.TempDir(), "p.go")
	if err := os.WriteFile(filename, []byte(src), 0600); err != nil {
		t.Fatal(err)
	}

	// As for a cgo file in the go build cache, the file that the package was
	// loaded from has no .go extension, so that doc.NewFromFiles fails for it.
	// It is a single line, so any position that is incorrectly resolved using
	// the package's FileSet is on that line, which reliably breaks the
	// placement of comments.
	fset := token.NewFileSet()
	loadedSrc := "package p // " + strings.Repeat("x", len(src)) + "\n"
	loaded, err := parser.ParseFile(fset, "generated", loadedSrc, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	pkg := &packages.Package{
		PkgPath: "example.com/p",
		Fset:    fset,
		GoFiles: []string{filename},
	}

	docPkg, docFset, err := packageDoc(pkg, []*ast.File{loaded}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if docFset == fset {
		t.Errorf("expected the fallback's FileSet, not that of the package")
	}

	got, err := newOutputState(docPkg, pkg, docFset, goPkgSite("pkg.go.dev")).outputPackage()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"type R struct {\n\tID  uint64\n\tSeq uint64\n\t// From is a doc comment\n\t// on two lines.\n\tFrom  uint64\n\tFrame []byte\n}\n",
		"type Q struct {\n\t// Wait is a doc comment.\n\tWait int\n}\n",
		"const (\n\t// A is a constant.\n\tA = 1\n\t// B is a constant.\n\tB = 2\n)\n",
		"func F(a int) int\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output does not contain:\n%s\ngot:\n%s", want, got)
		}
	}
}

func TestValueDeclConsts(t *testing.T) {
	for _, tc := range []struct {
		name, src, want string
	}{
		{"line", "// A is a constant.\nconst A = 1\n",
			"const A = 1"},
		{"multiple names on a line", "const A, B = 1, 2\n",
			"const A, B = 1, 2"},
		{"group of one", "const (\n\t// A is a constant.\n\tA = 1 // trailing\n)\n",
			"// A is a constant.\nconst A = 1 // trailing"},
		{"group of two", "const (\n\t// A is a constant.\n\tA = 1\n\tB = 2\n)\n",
			"const (\n\t// A is a constant.\n\tA = 1\n\tB = 2\n)"},
		{"group of three", "const (\n\tA = 1\n\tB = 2\n\tC = 3\n)\n",
			"const (\n\tA = 1\n\tB = 2\n\tC = 3\n)"},
		{"unexported removed leaving one", "const (\n\t// A is a constant.\n\tA = 1\n\t// b is hidden.\n\tb = 2\n)\n",
			"// A is a constant.\nconst A = 1"},
		{"unexported removed leaving two", "const (\n\tA = 1\n\tb = 2\n\tC = 3\n)\n",
			"const (\n\tA = 1\n\n\tC = 3\n)"},
		{"iota", "const (\n\tA = iota\n\tB\n)\n",
			"const (\n\tA = iota\n\tB\n)"},
		{"iota without a value left", "const (\n\ta = iota\n\tB\n)\n",
			"const (\n\tB\n)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "p.go", "package p\n\n"+tc.src, parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			pkg := &packages.Package{PkgPath: "example.com/p", Fset: fset}
			docPkg, docFset, err := packageDoc(pkg, []*ast.File{file}, 0)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := len(docPkg.Consts), 1; got != want {
				t.Fatalf("got %v const declarations, want %v", got, want)
			}
			st := newOutputState(docPkg, pkg, docFset, goPkgSite("pkg.go.dev"))
			if got := st.valueDecl(docPkg.Consts[0].Decl); got != tc.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tc.want)
			}
		})
	}
}

func TestPackageDocFileSet(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "p.go", "package p\n", parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	pkg := &packages.Package{PkgPath: "example.com/p", Fset: fset}
	_, docFset, err := packageDoc(pkg, []*ast.File{file}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if docFset != fset {
		t.Errorf("expected the package's FileSet when no fallback is needed")
	}
}

func TestTitleFromFilename(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"claude-summary.md", "Claude Summary"},
		{"claude_summary.md", "Claude Summary"},
		{"notes.md", "Notes"},
		{"api-v2-spec.md", "Api V2 Spec"},
	}

	for _, tt := range tests {
		got := titleFromFilename(tt.filename)
		if got != tt.expected {
			t.Errorf("titleFromFilename(%q) = %q, want %q", tt.filename, got, tt.expected)
		}
	}
}

func TestShiftHeading(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"# Heading", "#### Heading"},
		{"## SubHeading", "#### SubHeading"},
		{"### Detail", "##### Detail"},
		{"#### DeepDetail", "###### DeepDetail"},
		{"##### MaxMinus1", "###### MaxMinus1"},
		{"###### MaxDepth", "###### MaxDepth"},
		{"  ## Indented", "  #### Indented"},
		{"Not a heading", "Not a heading"},
		{"#hashtag", "#hashtag"},
		{"#include <stdio.h>", "#include <stdio.h>"},
		{"####### TooManyHashes", "####### TooManyHashes"},
	}

	for _, tt := range tests {
		got := shiftHeading(tt.input)
		if got != tt.expected {
			t.Errorf("shiftHeading(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestProcessExtraMarkdown_WithH1(t *testing.T) {
	content := "# Claude Analysis Summary\n\nThis package does X.\n\n## Architecture\nIt has two modules.\n\n### Submodule A\nDetails here.\n"
	got := processExtraMarkdown("claude-summary.md", content)
	expected := "### Claude Analysis Summary (claude-summary.md)\n\nThis package does X.\n\n#### Architecture\nIt has two modules.\n\n##### Submodule A\nDetails here."

	if strings.TrimSpace(got) != strings.TrimSpace(expected) {
		t.Errorf("processExtraMarkdown with H1 failed.\nGot:\n%s\nWant:\n%s", got, expected)
	}
}

func TestProcessExtraMarkdown_SetextH1(t *testing.T) {
	content := "My Custom Document Title\n========================\n\nIntroductory description.\n\n## Section 1\nDetails 1.\n"
	got := processExtraMarkdown("setext.md", content)
	expected := "### My Custom Document Title (setext.md)\n\nIntroductory description.\n\n#### Section 1\nDetails 1."

	if strings.TrimSpace(got) != strings.TrimSpace(expected) {
		t.Errorf("processExtraMarkdown with Setext H1 failed.\nGot:\n%s\nWant:\n%s", got, expected)
	}
}

func TestProcessExtraMarkdown_ClosingHashes(t *testing.T) {
	content := "# Introduction to C# ###\n\nThis guide covers C# programming.\n"
	got := processExtraMarkdown("csharp.md", content)
	expected := "### Introduction to C# (csharp.md)\n\nThis guide covers C# programming."

	if strings.TrimSpace(got) != strings.TrimSpace(expected) {
		t.Errorf("processExtraMarkdown with closing hashes failed.\nGot:\n%s\nWant:\n%s", got, expected)
	}
}

func TestProcessExtraMarkdown_WithoutH1(t *testing.T) {
	content := "This package does Y without an H1 header.\n\n## Features\n- Feature 1\n- Feature 2\n"
	got := processExtraMarkdown("claude-summary.md", content)
	expected := "### Claude Summary (claude-summary.md)\n\nThis package does Y without an H1 header.\n\n#### Features\n- Feature 1\n- Feature 2"

	if strings.TrimSpace(got) != strings.TrimSpace(expected) {
		t.Errorf("processExtraMarkdown without H1 failed.\nGot:\n%s\nWant:\n%s", got, expected)
	}
}

func TestProcessExtraMarkdown_CodeBlockIgnored(t *testing.T) {
	content := "# Code Example\n\nHere is code:\n\n```bash\n# This is a bash comment, not H1\necho \"hello\"\n```\n\n## Next Section\nMore text.\n"
	got := processExtraMarkdown("example.md", content)
	expected := "### Code Example (example.md)\n\nHere is code:\n\n```bash\n# This is a bash comment, not H1\necho \"hello\"\n```\n\n#### Next Section\nMore text."

	if strings.TrimSpace(got) != strings.TrimSpace(expected) {
		t.Errorf("processExtraMarkdown with code block comments failed.\nGot:\n%s\nWant:\n%s", got, expected)
	}
}

func TestLoadExtraMarkdown(t *testing.T) {
	dir := t.TempDir()

	readmeContent := "# README Title"
	claudeContent := "# Claude Summary\n\nClaude details here."
	archContent := "Architecture details."

	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(readmeContent), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "claude-summary.md"), []byte(claudeContent), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "architecture.md"), []byte(archContent), 0600); err != nil {
		t.Fatal(err)
	}

	got, err := loadExtraMarkdown(dir, "README.md")
	if err != nil {
		t.Fatalf("loadExtraMarkdown returned error: %v", err)
	}

	// Should contain parent header
	if !strings.Contains(got, "## External Markdown Files Included Here") {
		t.Errorf("Expected output to contain '## External Markdown Files Included Here', got:\n%s", got)
	}

	// Should include architecture.md first, then claude-summary.md, with filenames in headers
	if !strings.Contains(got, "### Architecture (architecture.md)") {
		t.Errorf("Expected output to contain '### Architecture (architecture.md)', got:\n%s", got)
	}
	if !strings.Contains(got, "### Claude Summary (claude-summary.md)") {
		t.Errorf("Expected output to contain '### Claude Summary (claude-summary.md)', got:\n%s", got)
	}
	if strings.Contains(got, "README Title") {
		t.Errorf("Output should not contain ignored README.md content, got:\n%s", got)
	}

	// Verify order: architecture before claude-summary
	archIdx := strings.Index(got, "### Architecture (architecture.md)")
	claudeIdx := strings.Index(got, "### Claude Summary (claude-summary.md)")
	if archIdx > claudeIdx {
		t.Errorf("Expected architecture section before claude section, got indices %d, %d", archIdx, claudeIdx)
	}
}

func TestLoadExtraMarkdown_Security(t *testing.T) {
	dir := t.TempDir()

	// 1. Create a symlink pointing to an external file
	targetFile := filepath.Join(t.TempDir(), "external-secret.md")
	if err := os.WriteFile(targetFile, []byte("# Secret Content"), 0600); err != nil {
		t.Fatal(err)
	}
	symlinkPath := filepath.Join(dir, "symlink.md")
	if err := os.Symlink(targetFile, symlinkPath); err != nil {
		t.Logf("Symlinks not supported on this OS/env, skipping symlink test: %v", err)
	}

	// 2. Create a normal markdown file
	if err := os.WriteFile(filepath.Join(dir, "normal.md"), []byte("# Normal Content"), 0600); err != nil {
		t.Fatal(err)
	}

	got, err := loadExtraMarkdown(dir, "README.md")
	if err != nil {
		t.Fatalf("loadExtraMarkdown failed: %v", err)
	}

	if strings.Contains(got, "Secret Content") {
		t.Errorf("loadExtraMarkdown should skip symlinked files, but included secret content:\n%s", got)
	}
	if !strings.Contains(got, "Normal Content") {
		t.Errorf("loadExtraMarkdown should include normal regular files:\n%s", got)
	}
}
