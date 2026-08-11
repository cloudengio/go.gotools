// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
