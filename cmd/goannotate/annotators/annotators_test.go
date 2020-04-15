package annotators_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"cloudeng.io/cmdutil"
	"cloudeng.io/errors"
	"cloudeng.io/go/cmd/goannotate/annotators"
	"gopkg.in/yaml.v2"
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

func diff(t *testing.T, a, b string) string {
	cmd := exec.Command("diff", a, b)
	// Ignore return code since differences are expected.
	out, _ := cmd.CombinedOutput()
	return string(out)
}

type diffReport struct {
	name string
	diff string
}

func diffAll(t *testing.T, a, b []string) []diffReport {
	if got, want := len(a), len(b); got != want {
		t.Errorf("%v: got %v, want %v", errors.Caller(2, 1), got, want)
		return nil
	}
	var diffs []diffReport
	for i := range a {
		diffs = append(diffs, diffReport{
			name: filepath.Base(a[i]),
			diff: diff(t, a[i], b[i]),
		})
	}
	return diffs
}

func compare(t *testing.T, a, b []diffReport) {
	if got, want := len(a), len(b); got != want {
		t.Errorf("%v: got %v, want %v", errors.Caller(2, 1), got, want)
		return
	}
	for i := range a {
		if got, want := a[i].name, b[i].name; got != want {
			t.Errorf("%v: got %v, want %v", errors.Caller(2, 1), got, want)
			return
		}
		if got, want := a[i].diff, b[i].diff; got != want {
			t.Errorf("%v: %v: got %v, want %v", errors.Caller(2, 1), a[i].name, got, want)
			return
		}
	}
}

var initConfigOnce sync.Once

func initConfig(t *testing.T) {
	config := &struct {
		Annotations []annotators.Spec `yam:"annotations"`
	}{}
	initConfigOnce.Do(func() {
		buf, err := ioutil.ReadFile(filepath.Join("testdata", "config.yaml"))
		if err != nil {
			t.Fatalf("failed to read config file: %v", err)
		}
		err = yaml.Unmarshal(buf, &config)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
	})
}

func setup(t *testing.T) (string, func()) {
	initConfig(t)
	td, err := ioutil.TempDir("", "goannotate")
	if err != nil {
		t.Fatalf("ioutil.TempDir: %v", err)
	}
	t.Logf("tempdir: %v", td)
	return td, func() {
		if !t.Failed() {
			os.RemoveAll(td)
		}
	}
}
