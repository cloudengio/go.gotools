package locate_test

import (
	"context"
	"sort"
	"strings"
	"testing"

	"cloudeng.io/errors"
	"cloudeng.io/go/locate"
)

const here = "cloudeng.io/go/locate/testdata/"

func compareLocations(t *testing.T, value string, prefixes, suffixes []string) {
	loc := errors.Caller(2, 1)
	locations := strings.Split(value, "\n")
	if got, want := len(locations), len(suffixes)+1; got != want {
		t.Errorf("%v: got %v, want %v", loc, got, want)
		return
	}
	sort.Strings(locations)
	for i, l := range locations {
		if i == 0 {
			// empty line comes first after sorting.
			if got, want := len(l), 0; got != want {
				t.Errorf("%v: got %v, want %v", loc, got, want)
			}
			continue
		}
		if got, want := l, prefixes[i-1]; !strings.HasPrefix(got, want) {
			t.Errorf("%v: %v doesn't start with %v", loc, got, want)
		}
		if got, want := l, suffixes[i-1]; !strings.HasSuffix(got, want) {
			t.Errorf("%v: got %v doesn't have suffix %v", loc, got, want)
		}
	}
}

func compareFiles(t *testing.T, files string, expected ...string) {
	loc := errors.Caller(2, 1)
	found := strings.Split(files, "\n")
	sort.Strings(found)
	if got, want := len(found), len(expected)+1; got != want {
		t.Errorf("%v: got %v, want %v", loc, got, want)
		return
	}
	for i, f := range found {
		if i == 0 {
			// empty line comes first after sorting.
			if got, want := len(f), 0; got != want {
				t.Errorf("%v: got %v, want %v", loc, got, want)
			}
			continue
		}
		if got, want := f, expected[i-1]; !strings.Contains(got, want) {
			t.Errorf("%v: got %v doesn't have suffix %v", loc, got, want)
		}
	}
}

func compareSlices(t *testing.T, got, want []string) {
	if got, want := len(got), len(want); got != want {
		t.Errorf("%v: got %v, want %v", errors.Caller(2, 1), got, want)
		return
	}
	for i := range got {
		if got, want := got[i], want[i]; !strings.HasSuffix(got, want) {
			t.Errorf("%v: got %v does not end with %v", errors.Caller(2, 1), got, want)
			return
		}
	}
}

func TestMultiPackageError(t *testing.T) {
	ctx := context.Background()

	locator := locate.New()
	locator.AddFunctions(here+"data.nomatch", "notapackage")
	err := locator.Do(ctx)
	if err == nil || !strings.Contains(err.Error(), "failed to find: notapackage") {
		t.Fatalf("expected a specific error, but got: %v", err)
	}

	locator = locate.New()
	locator.AddInterfaces(here + "multipackage")
	err = locator.Do(ctx)
	if err == nil || !strings.Contains(err.Error(), "failed to type check: cloudeng.io/go/locate/testdata/multipackage") {
		t.Fatalf("expected a specific error, but got: %v", err)
	}

	locator = locate.New()
	locator.AddInterfaces(here + "parseerror")
	err = locator.Do(ctx)
	if err == nil || !strings.Contains(err.Error(), "failed to type check: cloudeng.io/go/locate/testdata/parseerror") {
		t.Fatalf("expected a specific error, but got: %v", err)
	}

	locator = locate.New()
	locator.AddInterfaces(here + "typeerror")
	err = locator.Do(ctx)
	if err == nil || !strings.Contains(err.Error(), "failed to type check: cloudeng.io/go/locate/testdata/typeerror") {
		t.Fatalf("expected a specific error, but got: %v", err)
	}

	locator = locate.New()
	locator.AddInterfaces(here + "data.(")
	err = locator.Do(ctx)
	if err == nil || !strings.Contains(err.Error(), "failed to compile regexp") {
		t.Fatalf("expected a specific error, but got: %v", err)
	}
}
