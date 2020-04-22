package annotators_test

import (
	"path/filepath"
	"testing"

	"cloudeng.io/cmdutil"
)

const here = "cloudeng.io/go/cmd/goannotate/annotators/testdata/"

func list(t *testing.T, dir string) []string {
	paths, err := cmdutil.ListRegular(dir)
	if err != nil {
		t.Fatalf("ListRegular: %v", err)
	}
	absPaths := make([]string, len(paths))
	for i, p := range paths {
		absPaths[i] = filepath.Join(dir, p)
	}
	return absPaths
}
