// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
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

	got, err := loadExtraMarkdown(openTestRoot(t, dir), "README.md")
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

	got, err := loadExtraMarkdown(openTestRoot(t, dir), "README.md")
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

func testDocPackage(t *testing.T, pkgPath, src string) (*packages.Package, *token.FileSet, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "p.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	pkg := &packages.Package{PkgPath: pkgPath, Fset: fset}
	return pkg, fset, file
}

func openTestRoot(t *testing.T, dir string) *os.Root {
	t.Helper()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Close() })
	return root
}

func writeTestFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRenderPackageMarkdown_WithTemplate(t *testing.T) {
	pkg, _, file := testDocPackage(t, "example.com/p", "// Package p provides helpers.\npackage p\n\n// Hello greets.\nfunc Hello() {}\n")
	docPkg, docFset, err := packageDoc(pkg, []*ast.File{file}, 0)
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	intro := "# Custom Intro\n\nLiteral braces: {{not-a-template}}.\n"
	design := "## Design Notes\n\nDetails from subdir.\n"
	ignored := "# Ignored\n\nShould not appear.\n"
	tpl := "{{include \"intro.md\"}}\n{{godoc}}\n{{include \"docs/design.md\"}}"

	writeTestFiles(t, dir, map[string]string{
		"intro.md":       intro,
		"docs/design.md": design,
		"ignored.md":     ignored,
		templateFilename: tpl,
	})

	got, err := renderPackageMarkdown(dir, renderConfig{mdOutput: "README.md"}, docPkg, pkg, docFset,
		markdownFlavour("github"),
		goPkgSite("pkg.go.dev"),
	)
	if err != nil {
		t.Fatalf("renderPackageMarkdown failed: %v", err)
	}

	for _, want := range []string{
		intro,
		design,
		"# Package [example.com/p](https://pkg.go.dev/example.com/p?tab=doc)",
		"func Hello()",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{
		"External Markdown Files Included Here",
		"Should not appear",
	} {
		if strings.Contains(got, unwanted) {
			t.Errorf("did not expect %q in output, got:\n%s", unwanted, got)
		}
	}

	introIdx := strings.Index(got, "# Custom Intro")
	godocIdx := strings.Index(got, "# Package [example.com/p]")
	designIdx := strings.Index(got, "## Design Notes")
	if introIdx >= godocIdx || godocIdx >= designIdx {
		t.Errorf("unexpected section order: intro=%d, godoc=%d, design=%d\ngot:\n%s", introIdx, godocIdx, designIdx, got)
	}
}

func TestRenderPackageMarkdown_CommandWithTemplate(t *testing.T) {
	pkg, _, file := testDocPackage(t, "example.com/cmd/mytool", "// Usage of mytool:\n//\n// Overview:\n// mytool runs tasks.\npackage main\n")
	docPkg, docFset, err := packageDoc(pkg, []*ast.File{file}, 0)
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	writeTestFiles(t, dir, map[string]string{
		"extra.md":       "## Extra\n\nExtra text.\n",
		templateFilename: "{{godoc}}\n{{include \"extra.md\"}}",
	})

	got, err := renderPackageMarkdown(dir, renderConfig{mdOutput: "README.md"}, docPkg, pkg, docFset,
		markdownFlavour("github"),
		goPkgSite("pkg.go.dev"),
		asCommand(true),
	)
	if err != nil {
		t.Fatalf("renderPackageMarkdown failed: %v", err)
	}

	if !strings.Contains(got, "# [example.com/cmd/mytool](https://pkg.go.dev/example.com/cmd/mytool?tab=doc)") {
		t.Errorf("expected command godoc header, got:\n%s", got)
	}
	if !strings.Contains(got, "# Usage of `mytool`") || !strings.Contains(got, "`mytool` runs tasks.") {
		t.Errorf("expected command formatting in godoc output, got:\n%s", got)
	}
	if !strings.Contains(got, "## Extra\n\nExtra text.\n") {
		t.Errorf("expected included extra.md, got:\n%s", got)
	}
}

func TestRenderPackageMarkdown_WithoutTemplate(t *testing.T) {
	pkg, _, file := testDocPackage(t, "example.com/p", "// Package p provides helpers.\npackage p\n")
	docPkg, docFset, err := packageDoc(pkg, []*ast.File{file}, 0)
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	writeTestFiles(t, dir, map[string]string{"notes.md": "# My Notes\n\nBody.\n"})

	got, err := renderPackageMarkdown(dir, renderConfig{mdOutput: "README.md"}, docPkg, pkg, docFset,
		markdownFlavour("github"),
		goPkgSite("pkg.go.dev"),
	)
	if err != nil {
		t.Fatalf("renderPackageMarkdown failed: %v", err)
	}

	if !strings.Contains(got, "## External Markdown Files Included Here") || !strings.Contains(got, "### My Notes (notes.md)") {
		t.Errorf("expected default external markdown inclusion without template, got:\n%s", got)
	}
}

func TestRenderPackageMarkdown_TemplateErrors(t *testing.T) {
	pkg, _, file := testDocPackage(t, "example.com/p", "package p\n")
	docPkg, docFset, err := packageDoc(pkg, []*ast.File{file}, 0)
	if err != nil {
		t.Fatal(err)
	}

	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "secret.md")
	if err := os.WriteFile(outsideFile, []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		setup func(t *testing.T, dir string)
	}{
		{
			name: "missing include file",
			setup: func(t *testing.T, dir string) {
				writeTestFiles(t, dir, map[string]string{templateFilename: `{{include "missing.md"}}`})
			},
		},
		{
			name: "invalid template syntax",
			setup: func(t *testing.T, dir string) {
				writeTestFiles(t, dir, map[string]string{templateFilename: `{{include `})
			},
		},
		{
			name: "path traversal blocked",
			setup: func(t *testing.T, dir string) {
				writeTestFiles(t, dir, map[string]string{templateFilename: `{{include "../secret.md"}}`})
			},
		},
		{
			name: "absolute path blocked",
			setup: func(t *testing.T, dir string) {
				tpl := `{{include "` + filepath.ToSlash(outsideFile) + `"}}`
				writeTestFiles(t, dir, map[string]string{templateFilename: tpl})
			},
		},
		{
			name: "include directory rejected",
			setup: func(t *testing.T, dir string) {
				if err := os.Mkdir(filepath.Join(dir, "subdir"), 0700); err != nil {
					t.Fatal(err)
				}
				writeTestFiles(t, dir, map[string]string{templateFilename: `{{include "subdir"}}`})
			},
		},
		{
			name: "include output file rejected",
			setup: func(t *testing.T, dir string) {
				writeTestFiles(t, dir, map[string]string{
					"README.md":      "# Previous output\n",
					templateFilename: `{{include "./readme.md"}}`,
				})
			},
		},
		{
			name: "include symlink to output file rejected",
			setup: func(t *testing.T, dir string) {
				writeTestFiles(t, dir, map[string]string{
					"README.md":      "# Previous output\n",
					templateFilename: `{{include "alias.md"}}`,
				})
				if err := os.Symlink("README.md", filepath.Join(dir, "alias.md")); err != nil {
					t.Skipf("symlinks not supported: %v", err)
				}
			},
		},
		{
			name: "include hard link to output file rejected",
			setup: func(t *testing.T, dir string) {
				writeTestFiles(t, dir, map[string]string{
					"README.md":      "# Previous output\n",
					templateFilename: `{{include "alias.md"}}`,
				})
				if err := os.Link(filepath.Join(dir, "README.md"), filepath.Join(dir, "alias.md")); err != nil {
					t.Skipf("hard links not supported: %v", err)
				}
			},
		},
		{
			name: "template is directory rejected",
			setup: func(t *testing.T, dir string) {
				if err := os.Mkdir(filepath.Join(dir, templateFilename), 0700); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "symlink escaping root blocked",
			setup: func(t *testing.T, dir string) {
				if err := os.Symlink(outsideFile, filepath.Join(dir, "escape.md")); err != nil {
					t.Skipf("symlinks not supported: %v", err)
				}
				writeTestFiles(t, dir, map[string]string{templateFilename: `{{include "escape.md"}}`})
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			tc.setup(t, dir)
			if _, err := renderPackageMarkdown(dir, renderConfig{mdOutput: "README.md"}, docPkg, pkg, docFset,
				markdownFlavour("github"),
				goPkgSite("pkg.go.dev"),
			); err == nil {
				t.Errorf("expected error for %s, got nil", tc.name)
			}
		})
	}
}

func TestReadBoundedFile_SizeLimit(t *testing.T) {
	dir := t.TempDir()
	writeTestFiles(t, dir, map[string]string{"small.md": "hello", "large.md": "hello world!"})
	root := openTestRoot(t, dir)

	got, err := readBoundedFile(root, "small.md", 5)
	if err != nil {
		t.Fatalf("unexpected error for exact-limit file: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("got %q, want %q", string(got), "hello")
	}

	if _, err := readBoundedFile(root, "large.md", 5); err == nil {
		t.Errorf("expected size-limit error for large.md, got nil")
	}
}

func TestExcludeFlagParsing(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "default",
			args: nil,
			want: []string{defaultExclude},
		},
		{
			name: "comma separated",
			args: []string{"--exclude=a.md,b.md"},
			want: []string{"a.md", "b.md"},
		},
		{
			name: "repeated",
			args: []string{"--exclude=a.md", "--exclude=b.md"},
			want: []string{"a.md", "b.md"},
		},
		{
			name: "empty clears default",
			args: []string{"--exclude="},
			want: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			var cf cmdlineFlags
			if err := registerAndParseFlags(fs, tc.args, &cf); err != nil {
				t.Fatalf("registerAndParseFlags failed: %v", err)
			}
			if got, want := fs.Lookup("exclude").DefValue, strconv.Quote(defaultExclude); got != want {
				t.Errorf("DefValue = %q, want %q", got, want)
			}
			if !slices.Equal(cf.Exclude.Values, tc.want) {
				t.Errorf("exclusions = %v, want %v", cf.Exclude.Values, tc.want)
			}
		})
	}

	for _, bad := range []string{"--exclude=docs/a.md", "--exclude=..", "--exclude=."} {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		var cf cmdlineFlags
		if err := registerAndParseFlags(fs, []string{bad}, &cf); err == nil {
			t.Errorf("expected error for %q, got nil", bad)
		}
	}
}

func TestRenderPackageMarkdown_ExclusionsAndConfig(t *testing.T) {
	pkg, _, file := testDocPackage(t, "example.com/p", "// Package p provides helpers.\npackage p\n")
	docPkg, docFset, err := packageDoc(pkg, []*ast.File{file}, 0)
	if err != nil {
		t.Fatal(err)
	}

	baseFiles := map[string]string{
		"code_review.md": "# Code Review\n\nReview content.\n",
		"notes.md":       "# Notes\n\nNotes content.\n",
		"design.md":      "# Design\n\nDesign content.\n",
	}

	tests := []struct {
		name        string
		configYAML  string
		writeConfig bool
		cliExcludes []string
		want        []string
		unwanted    []string
	}{
		{
			name:        "cli exclude filters code_review.md",
			cliExcludes: []string{"code_review.md"},
			want:        []string{"Notes content.", "Design content."},
			unwanted:    []string{"Review content."},
		},
		{
			name:        "config exclusions override cli exclude",
			writeConfig: true,
			configYAML:  "exclusions:\n  - notes.md\n",
			cliExcludes: []string{"code_review.md"},
			want:        []string{"Review content.", "Design content."},
			unwanted:    []string{"Notes content."},
		},
		{
			name:        "config empty exclusions overrides cli exclude",
			writeConfig: true,
			configYAML:  "exclusions: []\n",
			cliExcludes: []string{"code_review.md"},
			want:        []string{"Review content.", "Notes content.", "Design content."},
		},
		{
			name:        "config without exclusions keeps cli exclude",
			writeConfig: true,
			configYAML:  "{}\n",
			cliExcludes: []string{"code_review.md"},
			want:        []string{"Notes content.", "Design content."},
			unwanted:    []string{"Review content."},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeTestFiles(t, dir, baseFiles)
			if tc.writeConfig {
				writeTestFiles(t, dir, map[string]string{configFilename: tc.configYAML})
			}
			rc := renderConfig{mdOutput: "README.md", exclusions: tc.cliExcludes}
			got, err := renderPackageMarkdown(dir, rc, docPkg, pkg, docFset,
				markdownFlavour("github"),
				goPkgSite("pkg.go.dev"),
			)
			if err != nil {
				t.Fatalf("renderPackageMarkdown failed: %v", err)
			}
			for _, w := range tc.want {
				if !strings.Contains(got, w) {
					t.Errorf("expected %q in output, got:\n%s", w, got)
				}
			}
			for _, u := range tc.unwanted {
				if strings.Contains(got, u) {
					t.Errorf("did not expect %q in output, got:\n%s", u, got)
				}
			}
		})
	}
}

func TestRenderPackageMarkdown_ConfigErrors(t *testing.T) {
	pkg, _, file := testDocPackage(t, "example.com/p", "package p\n")
	docPkg, docFset, err := packageDoc(pkg, []*ast.File{file}, 0)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		setup func(t *testing.T, dir string)
	}{
		{
			name: "invalid yaml",
			setup: func(t *testing.T, dir string) {
				writeTestFiles(t, dir, map[string]string{configFilename: "exclusions: [unclosed"})
			},
		},
		{
			name: "unknown key",
			setup: func(t *testing.T, dir string) {
				writeTestFiles(t, dir, map[string]string{configFilename: "unknown_key: [a.md]\n"})
			},
		},
		{
			name: "exclusion is a path",
			setup: func(t *testing.T, dir string) {
				writeTestFiles(t, dir, map[string]string{configFilename: "exclusions: [docs/notes.md]\n"})
			},
		},
		{
			name: "config is directory",
			setup: func(t *testing.T, dir string) {
				if err := os.Mkdir(filepath.Join(dir, configFilename), 0700); err != nil {
					t.Fatal(err)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			tc.setup(t, dir)
			if _, err := renderPackageMarkdown(dir, renderConfig{mdOutput: "README.md"}, docPkg, pkg, docFset,
				markdownFlavour("github"),
				goPkgSite("pkg.go.dev"),
			); err == nil {
				t.Errorf("expected error for %s, got nil", tc.name)
			}
		})
	}
}

func TestRenderPackageMarkdown_TemplateIgnoresConfig(t *testing.T) {
	pkg, _, file := testDocPackage(t, "example.com/p", "// Package p provides helpers.\npackage p\n")
	docPkg, docFset, err := packageDoc(pkg, []*ast.File{file}, 0)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	writeTestFiles(t, dir, map[string]string{
		configFilename:   "unknown_key: [a.md]\n",
		"notes.md":       "# Notes\n\nNotes content.\n",
		templateFilename: `{{include "notes.md"}}`,
	})
	rc := renderConfig{mdOutput: "README.md", exclusions: []string{"notes.md"}}
	got, err := renderPackageMarkdown(dir, rc, docPkg, pkg, docFset,
		markdownFlavour("github"),
		goPkgSite("pkg.go.dev"),
	)
	if err != nil {
		t.Fatalf("renderPackageMarkdown failed: %v", err)
	}
	if !strings.Contains(got, "Notes content.") {
		t.Errorf("expected included notes.md, got:\n%s", got)
	}
}

func TestCleanExclusions(t *testing.T) {
	got, err := cleanExclusions([]string{" a.md ", "", "./b.md"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"a.md", "b.md"}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	for _, bad := range []string{"docs/a.md", "../a.md", ".", ".."} {
		if _, err := cleanExclusions([]string{bad}); err == nil {
			t.Errorf("expected error for %q", bad)
		}
	}
}

func TestRenderPackageMarkdown_IncludeSubdirectory(t *testing.T) {
	pkg, _, file := testDocPackage(t, "example.com/p", "package p\n")
	docPkg, docFset, err := packageDoc(pkg, []*ast.File{file}, 0)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	writeTestFiles(t, dir, map[string]string{
		"docs/extra.md":  "## Nested\n",
		templateFilename: `{{include "docs/extra.md"}}`,
	})
	got, err := renderPackageMarkdown(dir, renderConfig{mdOutput: "README.md"}, docPkg, pkg, docFset,
		goPkgSite("pkg.go.dev"),
	)
	if err != nil {
		t.Fatalf("renderPackageMarkdown failed: %v", err)
	}
	if got != "## Nested\n" {
		t.Errorf("got %q, want %q", got, "## Nested\n")
	}
}

func TestLookupPackage(t *testing.T) {
	pkg := &packages.Package{PkgPath: "example.com/p"}
	xtest := &packages.Package{PkgPath: "example.com/p_test"}

	if got := lookupPackage(map[string]*packages.Package{
		"example.com/p":      pkg,
		"example.com/p_test": xtest,
	}, "example.com/p"); got != pkg {
		t.Errorf("expected the non-test package to be preferred, got %v", got)
	}

	got := lookupPackage(map[string]*packages.Package{"example.com/p_test": xtest}, "example.com/p")
	if got == nil {
		t.Fatal("expected fallback to the external test package, got nil")
	}
	if got.PkgPath != "example.com/p" {
		t.Errorf("PkgPath = %q, want %q", got.PkgPath, "example.com/p")
	}
	if xtest.PkgPath != "example.com/p_test" {
		t.Errorf("loaded test package was modified: PkgPath = %q", xtest.PkgPath)
	}

	if got := lookupPackage(map[string]*packages.Package{}, "example.com/p"); got != nil {
		t.Errorf("expected nil for a missing package, got %v", got)
	}
}

func TestWriteMarkdown_OnlyOutputsChangedFiles(t *testing.T) {
	origOverwrite := cmdFlags.Overwrite
	t.Cleanup(func() {
		cmdFlags.Overwrite = origOverwrite
	})

	capture := func(fn func() error) (string, error) {
		orig := os.Stdout
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout = w
		defer func() { os.Stdout = orig }()
		runErr := fn()
		_ = w.Close()
		out, readErr := io.ReadAll(r)
		_ = r.Close()
		if readErr != nil {
			t.Fatal(readErr)
		}
		return string(out), runErr
	}

	file := filepath.Join(t.TempDir(), "README.md")

	cmdFlags.Overwrite = false
	out, err := capture(func() error { return writeMarkdown(file, "# Initial\n") })
	if err != nil {
		t.Fatalf("new writeMarkdown failed: %v", err)
	}
	if strings.TrimSpace(out) != "updated: "+file {
		t.Errorf("got output %q, want %q", out, "updated: "+file)
	}

	out, err = capture(func() error { return writeMarkdown(file, "# Initial\n") })
	if err != nil {
		t.Fatalf("unchanged writeMarkdown failed: %v", err)
	}
	if out != "" {
		t.Errorf("expected no output for unchanged file, got %q", out)
	}

	cmdFlags.Overwrite = true
	out, err = capture(func() error { return writeMarkdown(file, "# Updated\n") })
	if err != nil {
		t.Fatalf("updated writeMarkdown failed: %v", err)
	}
	if strings.TrimSpace(out) != "updated: "+file {
		t.Errorf("got output %q, want %q", out, "updated: "+file)
	}
}
