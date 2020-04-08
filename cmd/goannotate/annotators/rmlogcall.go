package annotators

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"cloudeng.io/go/locate"
	"cloudeng.io/go/locate/locateutil"
	"cloudeng.io/text/edit"
	"golang.org/x/tools/go/packages"
	"gopkg.in/yaml.v2"
)

type RmLogCall struct {
	Type        string   `annotator:"name of annotator type."`
	Name        string   `annotator:"name of annotator configuration."`
	Interfaces  []string `annotator:"list of interfaces whose implementations are to have logging function calls removed from."`
	Functions   []string `annotator:"list of functionms that are to have function calls removed from."`
	Logcall     string   `annotator:"the logging function call to be removed"`
	Comment     string   `annotator:"optional comment that must appear in the comments associated with the function call if it is to be removed."`
	Deferred    bool     `annotator:"if set requires that the function to be removed must be defered."`
	Concurrency int      `annotator:"the number of goroutines to use, zero for a sensible default."`
}

func init() {
	Register(&RmLogCall{})
}

// New implements annotators.T.
func (rc *RmLogCall) New(name string) T {
	return &RmLogCall{Name: name}
}

// Unmarshal implements annotators.T.
func (rc *RmLogCall) Unmarshal(buf []byte) error {
	return yaml.Unmarshal(buf, rc)
}

// Describe implements annotators.T.
func (rc *RmLogCall) Describe() string {
	return MustDescribe(rc, "an annotator that removes instances of calls to functions.")
}

// Do implements annotators.T.
func (rc *RmLogCall) Do(ctx context.Context, pkgs []string) error {
	locator := locate.New(
		concurrencyOpt(rc.Concurrency),
		locate.Trace(Verbosef),
		locate.IgnoreMissingFuctionsEtc(),
	)
	locator.AddInterfaces(rc.Interfaces...)
	locator.AddFunctions(rc.Functions...)
	locator.AddPackages(pkgs...)
	Verbosef("locating functions to have a logcall annotation removed...")
	if err := locator.Do(ctx); err != nil {
		return fmt.Errorf("failed to locate functions and/or interface implementations: %v\n", err)
	}

	edits := map[string][]edit.Delta{}
	locator.WalkFunctions(func(fullname string,
		pkg *packages.Package,
		file *ast.File,
		fn *types.Func,
		decl *ast.FuncDecl,
		implements []string) {
		if !locateutil.HasBody(decl) {
			return
		}
		cmap := ast.NewCommentMap(pkg.Fset, file, file.Comments)
		nodes := locateutil.FunctionCalls(decl, rc.Logcall, rc.Deferred)
		var start, end token.Pos
		for _, node := range nodes {
			start = node.Pos()
			end = node.End()
			cgs := cmap[node]
			if locateutil.CommentGroupsContain(cgs, rc.Comment) {

				cFirst, cLast := locateutil.CommentGroupBounds(cgs)
				if cFirst < start {
					start = cFirst
				}
				if cLast > end {
					end = cLast
				}
			}
			from := pkg.Fset.PositionFor(start, false)
			to := pkg.Fset.PositionFor(end, false)
			delta := edit.Delete(from.Offset, to.Offset-from.Offset+1)
			edits[from.Filename] = append(edits[from.Filename], delta)
			Verbosef("delete: %v...%v\n", from, to)
		}
	})
	return applyEdits(ctx, edits)
}
