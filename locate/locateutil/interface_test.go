// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package locateutil_test

import (
	"go/types"
	"reflect"
	"sort"
	"testing"

	"cloudeng.io/go/locate/locateutil"
	"golang.org/x/tools/go/packages"
)

func TestInterfaceType(t *testing.T) {
	pkgs, err := packages.Load(packagesConfig,
		"cloudeng.io/go/locate/testdata/data",
	)
	if err != nil {
		t.Errorf("pkg.Load: %v", err)
	}
	pkg := pkgs[0]
	names := []string{}
	defs := []string{}
	for k, v := range pkg.TypesInfo.Defs {
		if v == nil {
			continue
		}
		if locateutil.InterfaceType(v.Type()) != nil {
			names = append(names, k.Name)
		}
		if locateutil.IsInterfaceDefinition(pkg, v) != nil {
			defs = append(defs, k.Name)
		}
	}
	sort.Strings(names)
	sort.Strings(defs)
	want := []string{
		"Field", "Ifc1", "Ifc2", "Ifc3", "IgnoredVariable", "hidden",
	}
	if got, want := names, want; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	want = []string{
		"Ifc1", "Ifc2", "Ifc3", "hidden",
	}
	if got, want := defs, want; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if got, want := locateutil.IsAbstract(nil), false; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := locateutil.IsInterfaceDefinition(pkg, nil), (*types.Interface)(nil); got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
