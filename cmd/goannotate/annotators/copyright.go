package annotators

import (
	"context"
	"fmt"
	"go/ast"
	"strings"

	"cloudeng.io/go/locate"
	"cloudeng.io/text/edit"
	"golang.org/x/tools/go/packages"
	"gopkg.in/yaml.v2"
)

type EnsureCopyrightAndLicense struct {
	Type        string   `annotator:"name of annotator type."`
	Name        string   `annotator:"name of annotation."`
	Packages    []string `annotator:"packages to be annotated"`
	Copyright   string   `annotator:"desired copyright notice."`
	License     string   `annotator:"desired license notice."`
	Concurrency int      `annotator:"the number of goroutines to use, zero for a sensible default."`
}

func init() {
	Register(&EnsureCopyrightAndLicense{})
}

// New implements annotators.Annotators.
func (ec *EnsureCopyrightAndLicense) New(name string) Annotation {
	return &EnsureCopyrightAndLicense{Name: name}
}

// UnmarshalYAML implements annotators.Annotations.
func (ec *EnsureCopyrightAndLicense) UnmarshalYAML(buf []byte) error {
	return yaml.Unmarshal(buf, ec)
}

// Describe implements annotators.Annotations.
func (ec *EnsureCopyrightAndLicense) Describe() string {
	return MustDescribe(ec,
		`an annotator that ensures that a copyright and license notice is 
present at the top of all files. It will not remove existing notices.`,
	)
}

// Do implements annotators.Annotations.
func (ec *EnsureCopyrightAndLicense) Do(ctx context.Context, root string, pkgs []string) error {
	if len(ec.Copyright) == 0 {
		return fmt.Errorf("missing or empty copyright specified in the configuration file")
	}
	locator := locate.New(
		concurrencyOpt(ec.Concurrency),
		locate.Trace(Verbosef),
		locate.IgnoreMissingFuctionsEtc(),
	)
	locator.AddPackages(pkgs...)
	Verbosef("locating functions to have a copyright/license annotation...")
	if err := locator.Do(ctx); err != nil {
		return fmt.Errorf("failed to locate functions and/or interface implementations: %v\n", err)
	}

	newCopyright := strings.TrimSuffix(ec.Copyright, "\n") + "\n"
	newLicense := strings.TrimSuffix(ec.License, "\n") + "\n\n"

	dirty := map[string]bool{}
	edits := map[string][]edit.Delta{}
	locator.WalkFiles(func(filename string,
		pkg *packages.Package,
		comments ast.CommentMap,
		file *ast.File,
		mask locate.HitMask) {

		tokenFile := pkg.Fset.File(file.Pos())

		var copyright *ast.Comment
		var licenseStart, licenseLen int
		for _, cg := range file.Comments {
			if tokenFile.Offset(cg.Pos()) == 0 && cg != file.Doc {
				// Find comment block at top of file that is not a 'doc' comment.
				comments := cg.List
				sanitized := strings.TrimSpace(strings.ToLower(cg.Text()))
				if strings.HasPrefix(sanitized, "copyright") {
					copyright = comments[0]
					if len(comments) > 1 {
						at := comments[1].Pos()
						licenseStart = tokenFile.Offset(at)
						licenseLen = tokenFile.Offset(cg.End()) - licenseStart
					}
				}
			}
		}
		var deltas []edit.Delta
		if copyright != nil {
			deltas = append(deltas, edit.ReplaceString(0, len(copyright.Text)+1, newCopyright))
			if licenseStart != 0 && len(ec.License) > 0 {
				deltas = append(deltas, edit.ReplaceString(licenseStart, licenseLen, newLicense))
			}
		} else {
			// New copy right and license.
			deltas = append(deltas, edit.InsertString(0, newCopyright))
			if len(ec.License) > 0 {
				deltas = append(deltas, edit.InsertString(0, newLicense))
			}
		}
		edits[filename] = append(edits[filename], deltas...)
		dirty[filename] = true
	})
	return applyEdits(ctx, computeOutputs(root, edits), edits)
}
