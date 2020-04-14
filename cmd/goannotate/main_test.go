package main_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"cloudeng.io/cmdutil"
)

func execit(t *testing.T, bin string, args ...string) string {
	cmd := exec.Command(bin, args...)
	out, err := cmd.Output()
	if err != nil {
		cl := strings.Join(cmd.Args, " ")
		var stderr string
		if execerr, ok := err.(*exec.ExitError); ok {
			stderr = string(execerr.Stderr)
		}
		t.Fatalf("failed to run: %v: %v: %v", cl, err, stderr)
	}
	return string(out)
}

var configFile = filepath.Join("annotators", "testdata", "config.yaml")

func runit(t *testing.T, name, packages string) (output string, tmpdir string) {
	td, err := ioutil.TempDir("", "goannotate")
	if err != nil {
		t.Fatalf("ioutil.TempDir: %v", err)
	}
	out := execit(t, "go", "run", ".", "--config="+configFile, "--annotation="+name, "--write-dir="+td, packages)
	return out, td
}

func list(t *testing.T, dir string) []string {
	cwd, _ := os.Getwd()
	paths, err := cmdutil.ListRegular(dir)
	if err != nil {
		t.Fatalf("ListRegular: %v", err)
	}
	absPaths := make([]string, len(paths))
	for i, p := range paths {
		absPaths[i] = filepath.Join(cwd, dir, p)
	}
	return absPaths
}

func TestCopyrightIsPresent(t *testing.T) {
	out, tmpdir := runit(t, "personal-apache", "cloudeng.io/go/cmd/goannotate/annotators/testdata/copyright")
	defer os.RemoveAll(tmpdir)
	original := list(t, filepath.Join("annotators", "testdata", "copyright"))
	lines := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
	sort.Strings(lines)
	if got, want := strings.Join(lines, "\n"), strings.Join(original, "\n"); got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestDescribe(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--config="+configFile, "--list")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	output := string(out)
	for _, expected := range []string{"AddLogCall", "EnsureCopyrightAndLicense", "RmLogCall"} {
		if !strings.Contains(output, expected) {
			t.Errorf("%v missing", expected)
		}
	}
}
