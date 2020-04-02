package locate_test

import (
	"context"
	"go/ast"
	"testing"

	"cloudeng.io/go/locate"
	"golang.org/x/tools/go/packages"
)

func TestFilesAndImports(t *testing.T) {
	ctx := context.Background()
	locator := locate.New()
	locator.AddInterfaces(here + "data")
	locator.AddFunctions(here+"imports", here+"data")
	err := locator.Do(ctx)
	if err != nil {
		t.Fatalf("locate.Do: %v", err)
	}
	start, stop, masks := []string{}, []string{}, []string{}
	locator.WalkFiles(func(filename string, pkg *packages.Package, comments ast.CommentMap, file *ast.File, has locate.HitMask) {
		begin, end := locate.ImportBlock(file)
		start = append(start, pkg.Fset.Position(begin).String())
		stop = append(stop, pkg.Fset.Position(end).String())
		masks = append(masks, has.String())
	})
	startAt := []string{
		"-",
		"-",
		"-",
		"blocks.go:3:1",
		"import.go:3:1",
		"imports.go:3:1",
	}
	stopAt := []string{
		"-",
		"-",
		"-",
		"blocks.go:8:2",
		"import.go:3:13",
		"imports.go:6:2",
	}
	contains := []string{
		"function",
		"function",
		"interface",
		"function",
		"function",
		"function",
	}
	compareSlices(t, start, startAt)
	compareSlices(t, stop, stopAt)
	compareSlices(t, masks, contains)
}
