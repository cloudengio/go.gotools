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
subheadings that include each file's filename. Files matching --exclude
(default code_review.md) are ignored. Exclusions are filenames, not paths,
and are matched case-insensitively. If a .gomarkdown.yaml file is present in
the package's directory, its 'exclusions' list overrides --exclude.

If a .gomarkdown.tpl file is present in the package's directory,
`gomarkdown` uses it as a Go text/template to generate the output instead
of appending external .md files automatically. In this case --exclude and
.gomarkdown.yaml are not used. Two template functions are available:
  - include "<path>" includes the contents of the specified file. The path
    is relative to the package's directory and may refer to a file in a
    subdirectory, but it may not refer to a file outside of the package's
    directory nor to the markdown output file (including via a symlink or
    hard link).
  - godoc generates the godoc markdown output for the package or command.

# Command line flags

    -circleci string
      	set to the circleci project to insert a circleci build badge
    -exclude value
      	comma-separated list of markdown files to exclude. (default "code_review.md")
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


