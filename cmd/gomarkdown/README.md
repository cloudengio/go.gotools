# [cloudeng.io/go/cmd/gomarkdown](https://pkg.go.dev/cloudeng.io/go/cmd/gomarkdown?tab=doc)


# Usage of `gomarkdown`

`gomarkdown` is a utility for generating markdown from go source comments.
It mirrors the behaviour of go doc but producing markdown output instead. It
is most useful for packages hosted on github since github will automatically
render the README.md in each directory.

`gomarkdown` can be applied to multiple packages and will generate a README.md
for each package in that package's directory.

For example, the following will generate a README.md in every package and
command under ./...

    go run cloudeng.io/go/cmd/`gomarkdown` --overwrite ./...

In addition `gomarkdown` can be used to generate markdown for command line
packages. In doing so it employs simple heurestics to format package
comments.
  - lines with fewer than 5 words that end in a : are treated as headings.
  - all occurrences of the command's name are highlighted.

`gomarkdown` will also include any other existing .md files in the package's
directory under an 'External Markdown Files Included Here' section, using
subheadings that include each file's filename.

# Command line flags

    -circleci string
      	set to the circleci project to insert a circleci build badge
    -gopkg string
      	link to this site for full godoc and godoc examples (default "pkg.go.dev")
    -goreportcard
      	insert a link to goreportcard.com
    -markdown string
      	markdown style to use, currently only github is supported. (default "github")
    -md-output string
      	name of markdown output file. (default "README.md")
    -overwrite
      	overwrite existing file.


## External Markdown Files Included Here

### Code Review: External Markdown Files Feature in `gomarkdown` (code_review.md)

#### Overview

This code review evaluates the changes introduced to `gomarkdown` for including external `.md` files as subsections in generated package documentation (`README.md`). The review covers **logic correctness**, **performance efficiency**, and **security/robustness**.

---

#### 1. Logic Issues

##### 1.1. False-Positive ATX H1 Heading Detection (`isH1`)
- **Location**: [`output.go`](file:///Users/cnicolaou/LocalOnly/dev/github.com/cloudengio/go.gotools/cmd/gomarkdown/output.go#L341-L349)
- **Issue**: `isH1` checks if a trimmed line starts with `#` and does not start with `##`. However, standard Markdown (CommonMark / GFM) requires a space or tab (or end-of-line) immediately following `#` for an ATX heading.
- **Impact**: Lines such as `#hashtag`, `#1_issue`, or C/C++ preprocessor directives like `#include <stdio.h>` will be incorrectly parsed as the document's H1 title.

##### 1.2. Lack of Support for Setext H1 Headings
- **Location**: [`output.go`](file:///Users/cnicolaou/LocalOnly/dev/github.com/cloudengio/go.gotools/cmd/gomarkdown/output.go#L341-L349)
- **Issue**: `isH1` only checks ATX headings (`# Title`). In Markdown, top-level headings can also be written using Setext underlines (`Title\n===`).
- **Impact**: External markdown files using Setext H1 headings will fail to extract the title, falling back to title-casing the filename while leaving the Setext heading unshifted in the body.

##### 1.3. Inaccurate Heading Level Shifting (`shiftHeading`)
- **Location**: [`output.go`](file:///Users/cnicolaou/LocalOnly/dev/github.com/cloudengio/go.gotools/cmd/gomarkdown/output.go#L365-L387)
- **Issue**: `shiftHeading` counts leading `#` characters without validating if a space/tab follows them. Additionally, lines with 7+ `#` characters (which are invalid ATX headings in Markdown) get shifted and capped to `######`.
- **Impact**: Non-heading text starting with `#` (e.g. `#comment`) will be mutated into `####comment`.

##### 1.4. Fragile Code Block Fence Tracking
- **Location**: [`output.go`](file:///Users/cnicolaou/LocalOnly/dev/github.com/cloudengio/go.gotools/cmd/gomarkdown/output.go#L389-L443)
- **Issue**: `inCodeBlock` toggles blindly on any line starting with ```` ``` ```` or `~~~`. In CommonMark:
  1. A code block fence opened with backticks can only be closed by backticks (not tildes), and vice-versa.
  2. A fence opened with $N$ characters can only be closed by a fence of $\ge N$ matching characters.
  3. The processing logic executes two separate passes over the line slice, duplicating fence state tracking.
- **Impact**: Nested code fences or code blocks containing literal ` ``` ` strings can cause `inCodeBlock` tracking to desync, corrupting or misinterpreting code comments as Markdown headings.

##### 1.5. Incorrect ATX Closing Sequence Trimming
- **Location**: [`output.go`](file:///Users/cnicolaou/LocalOnly/dev/github.com/cloudengio/go.gotools/cmd/gomarkdown/output.go#L408-L410)
- **Issue**: `strings.TrimRight(rawTitle, "#")` strips all trailing `#` characters from the extracted title string unconditionally.
- **Impact**: Titles ending in `#` (e.g. `# Introduction to C#`) will have their closing `#` stripped incorrectly, turning `C#` into `C`. Per GFM specs, closing `#` sequences must be preceded by spaces/tabs.

---

#### 2. Performance Issues

##### 2.1. Multiple Allocations & Intermediate String Copies
- **Location**: [`output.go`](file:///Users/cnicolaou/LocalOnly/dev/github.com/cloudengio/go.gotools/cmd/gomarkdown/output.go#L389-L443)
- **Issue**: `processExtraMarkdown` performs multiple memory-heavy transformations per file:
  1. `strings.Split(content, "\n")` allocates a slice of all lines.
  2. Appends lines to `bodyLines` slice.
  3. Appends lines to `shiftedLines` slice.
  4. `strings.Join(shiftedLines, "\n")` allocates a new combined string.
- **Impact**: Unnecessary garbage collection pressure and $O(N)$ memory duplication per external file. A single-pass scanner (`bufio.Scanner` or `strings.Builder`) can process file content in a single pass with minimal allocations.

##### 2.2. Redundant Double-Pass Processing
- **Location**: [`output.go`](file:///Users/cnicolaou/LocalOnly/dev/github.com/cloudengio/go.gotools/cmd/gomarkdown/output.go#L398-L436)
- **Issue**: The file content is iterated over twice—first to extract H1, and second to shift headings.
- **Impact**: Double work for line parsing and code block state checking.

---

#### 3. Security & Robustness Issues

##### 3.1. Unbounded File Reading (Memory Exhaustion / OOM Risk)
- **Location**: [`output.go`](file:///Users/cnicolaou/LocalOnly/dev/github.com/cloudengio/go.gotools/cmd/gomarkdown/output.go#L330)
- **Issue**: `os.ReadFile` reads the entire contents of any matching `.md` file into memory without size constraints.
- **Impact**: If a huge `.md` file (or a special device/FIFO file) is present in the package directory, it could lead to memory exhaustion (OOM) and crash the tool.

##### 3.2. Symlink Traversal
- **Location**: [`output.go`](file:///Users/cnicolaou/LocalOnly/dev/github.com/cloudengio/go.gotools/cmd/gomarkdown/output.go#L300-L330)
- **Issue**: `os.ReadDir` returns directory entries including symlinks. `os.ReadFile` will follow symlinks pointing outside the repository root.
- **Impact**: A symlink pointing to sensitive external markdown files could accidentally leak or embed external files into the generated `README.md`.

---

#### 4. Step-by-Step Fix Plan

> **Note**: Do not apply code fixes until instructed.

1. **Fix ATX & Setext Heading Parsing (`isH1` / `processExtraMarkdown`)**:
   - Update heading detection to require space/tab/EOL after `#` (`# Heading`).
   - Add support for Setext H1 headings (`Underline\n===`).
   - Only trim closing `#` sequences if preceded by a space/tab.

2. **Fix Heading Shifting (`shiftHeading`)**:
   - Verify space/tab follows `#` before shifting heading levels.
   - Leave non-heading lines starting with `#` (such as `#hashtag` or C comments) unmodified.

3. **Improve Code Block Fence Matching & Single-Pass Processing**:
   - Refactor `processExtraMarkdown` to use a single pass over file lines.
   - Track code block fence type (``` vs ~~~) and minimum length so fence closing is robust.

4. **Add Security Bounds & Symlink Verification (`loadExtraMarkdown`)**:
   - Enforce a maximum file size limit (e.g. 5 MB) per external `.md` file using `io.LimitReader`.
   - Skip symlinks (`entry.Type()&os.ModeSymlink != 0`) or verify canonical paths stay within the package directory.

5. **Optimize Memory Allocations**:
   - Replace intermediate slice allocations (`bodyLines`, `shiftedLines`, `strings.Split`) with a `bufio.Scanner` streaming pipeline writing directly into a `strings.Builder`.

6. **Update Unit Tests**:
   - Add test cases in `output_test.go` for:
     - Non-heading `#` strings (`#include`, `#hashtag`).
     - Setext headings (`Title\n===`).
     - Code block fence edge cases (nested fences).
     - Oversized files and symlinks.


