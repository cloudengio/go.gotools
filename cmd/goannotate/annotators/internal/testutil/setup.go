package testutil

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"cloudeng.io/go/cmd/goannotate/annotators"
	"cloudeng.io/go/cmd/goannotate/annotators/functions"
	"cloudeng.io/go/locate"
	"gopkg.in/yaml.v2"
)

var (
	initConfigOnce sync.Once
	initErr        error
)

func initConfig(t *testing.T, v interface{}) {
	initConfigOnce.Do(func() {
		buf, err := ioutil.ReadFile(filepath.Join("testdata", "config.yaml"))
		if err != nil {
			initErr = fmt.Errorf("failed to read config file: %v", err)
			return
		}
		err = yaml.Unmarshal(buf, v)
		if err != nil {
			initErr = fmt.Errorf("error unmarshaling config: %v", err)
		}
	})
	if initErr != nil {
		// make sure error is reported
		t.Fatalf("error: %v", initErr)
	}

}

// SetupAnnotators reads  ./testdata/config.yaml and initializes
// the annototators package, creates a temp directory and a cleanup function
// to remove the test directory on test failures.
func SetupAnnotators(t *testing.T) (string, func()) {
	config := &struct {
		Annotations []annotators.Spec `yam:"annotations"`
	}{}
	initConfig(t, config)
	td, err := ioutil.TempDir("", "goannotate")
	if err != nil {
		t.Fatalf("ioutil.TempDir: %v", err)
	}
	if len(annotators.Available()) == 0 {
		t.Fatalf("no annotations found")
	}
	t.Logf("tempdir: %v", td)
	return td, func() {
		if !t.Failed() {
			os.RemoveAll(td)
		}
	}
}

func SetupFunctions(t *testing.T) {
	config := &struct {
		Generators []functions.Spec `yam:"generators"`
	}{}
	initConfig(t, config)
	if len(functions.CallGenerators()) == 0 {
		t.Fatalf("no call generators found")
	}
}

func LocatePackages(t *testing.T, ctx context.Context, pkgs ...string) *locate.T {
	locator := locate.New()
	locator.AddFunctions(pkgs...)
	locator.AddPackages(pkgs...)
	if err := locator.Do(ctx); err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	return locator
}
