package locate_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"cloudeng.io/go/locate"
)

func implements(ifcs ...string) string {
	out := strings.Builder{}
	for i, ifc := range ifcs {
		out.WriteString(here)
		out.WriteString("data.")
		out.WriteString(ifc)
		if i < len(ifcs)-1 {
			out.WriteString(", ")
		}
	}
	return out.String()
}

func TestInterfaces(t *testing.T) {
	ctx := context.Background()
	locator := locate.New(locate.IgnoreMissingFuctionsEtc())
	locator.AddInterfaces(here+"data.xxxx", here+"data.Ifc2")
	locator.AddInterfaces(here+"data.Ifc1", here+"data.Ifc2")
	if err := locator.Do(ctx); err != nil {
		t.Fatalf("locator.Do: %v", err)
	}
	compareLocations(t, listInterfaces(locator), []string{
		here + "data.Ifc1",
		here + "data.Ifc2",
	}, []string{
		filepath.Join("data", "interfaces.go") + ":3:6",
		filepath.Join("data", "interfaces.go") + ":12:6",
	})

	locator = locate.New(locate.IgnoreMissingFuctionsEtc())
	locator.AddInterfaces(here+"data.xxxx", here+"data.Ifc2")
	locator.AddInterfaces(here+"data.Ifc1", here+"data.Ifc2")
	locator.AddInterfaces(here + "data")
	if err := locator.Do(ctx); err != nil {
		t.Fatalf("locator.Do: %v", err)
	}

	compareLocations(t, listInterfaces(locator), []string{
		here + "data.Ifc1",
		here + "data.Ifc2",
		here + "data.Ifc3",
	}, []string{
		filepath.Join("data", "interfaces.go") + ":3:6",
		filepath.Join("data", "interfaces.go") + ":12:6",
		filepath.Join("data", "interfaces.go") + ":16:6",
	})
	compareFiles(t, listFiles(locator), "data/interfaces.go: data")
}

func TestEmbeddedInterfaces(t *testing.T) {
	ctx := context.Background()

	locator := locate.New(locate.IgnoreMissingFuctionsEtc())
	locator.AddInterfaces(here + "data/embedded.StructEmbed")
	if err := locator.Do(ctx); err != nil {
		t.Fatalf("locator.Do: %v", err)
	}
	if got, want := len(listInterfaces(locator)), 0; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	locator = locate.New()
	locator.AddInterfaces(here + "data/embedded.IfcE$")
	locator.AddPackages(here + "data/embedded/pkg")
	if err := locator.Do(ctx); err != nil {
		t.Fatalf("locator.Do: %v", err)
	}
	compareLocations(t, listInterfaces(locator), []string{
		here + "data/embedded.IfcE",
		here + "data/embedded.IfcE1",
		here + "data/embedded.IfcE2",
		here + "data/embedded.ifcE3",
		here + "data/embedded/pkg.Pkg",
	}, []string{
		filepath.Join("data", "embedded", "embedded.go") + ":18:6",
		filepath.Join("data", "embedded", "embedded.go") + ":5:6",
		filepath.Join("data", "embedded", "embedded.go") + ":9:6",
		filepath.Join("data", "embedded", "embedded.go") + ":13:6",
		filepath.Join("data", "embedded", "pkg", "interface.go") + ":3:6",
	})
	compareFiles(t, listFiles(locator),
		filepath.Join("data", "embedded", "embedded.go")+": embedded",
		filepath.Join("data", "embedded", "pkg", "interface.go")+": pkg",
	)
}
func TestFindImplementations(t *testing.T) {
	ctx := context.Background()
	locator := locate.New()
	locator.AddInterfaces(here + "data")
	locator.AddPackages(here+"data", here+"impl")
	if err := locator.Do(ctx); err != nil {
		t.Fatalf("locator.Do: %v", err)
	}

	compareLocations(t, listFunctions(locator), []string{
		"(*" + here + "impl.Impl1).M1 implements " + implements("Ifc1"),
		"(*" + here + "impl.Impl1).M2 implements " + implements("Ifc1"),
		"(*" + here + "impl.Impl12).M1 implements " + implements("Ifc1", "Ifc2", "Ifc3"),
		"(*" + here + "impl.Impl12).M2 implements " + implements("Ifc1", "Ifc2", "Ifc3"),
		"(*" + here + "impl.Impl12).M3 implements " + implements("Ifc1", "Ifc2", "Ifc3"),
		"(*" + here + "impl.impl2).M3 implements " + implements("Ifc2"),
	}, []string{
		filepath.Join("impl", "impls.go") + ":5:1",
		filepath.Join("impl", "impls.go") + ":9:1",
		filepath.Join("impl", "impls.go:") + "22:1",
		filepath.Join("impl", "impls.go:") + "26:1",
		filepath.Join("impl", "impls.go:") + "30:1",
		filepath.Join("impl", "impls.go:") + "15:1",
	})
}
