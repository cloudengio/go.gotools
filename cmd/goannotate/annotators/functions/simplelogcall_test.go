// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package functions_test

import (
	"reflect"
	"testing"
)

func TestSimpleLogCall(t *testing.T) {
	importPath, calls := execute(t, ".SimpleLogCall")
	if got, want := importPath, "log"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	expectedCalls := []string{
		`log.Logf("a=%d", a)`,
		`log.Logf("a=%d", a)`,
	}
	if got, want := calls, expectedCalls; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
