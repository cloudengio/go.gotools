// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// This tool identifies transitive dependencies matching a specified import path
// prefix and suggests 'go mod edit --replace' directives for any that belong to
// workspace modules. It is intended to be used in a repository with a go.work
// file and multiple go modules to replace dependencies with local versions
// during CI testing to avoid multiple rounds go get updates and commits.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// Package represents the structure returned by 'go list -json'
type Package struct {
	Root       string   `json:"Root"`
	Dir        string   `json:"Dir"`
	ImportPath string   `json:"ImportPath"`
	Deps       []string `json:"Deps"` // Contains all transitive dependencies
}

func handleFlags() (string, string, bool) {
	prefix := flag.String("prefix", "", "import path prefix to match (required)")
	targetPkg := flag.String("pkg", "./...", "package path passed to go list")
	apply := flag.Bool("apply", false, "apply the go mod edit commands")

	flag.Parse()

	if *prefix == "" {
		fmt.Fprintln(os.Stderr, "Usage: go run replace-deps.go -prefix <prefix> [-pkg <package_path>]")
		fmt.Fprintln(os.Stderr, "Example: go run replace-deps.go -prefix github.com/gin-gonic -pkg ./...")
		os.Exit(1)
	}
	return *prefix, *targetPkg, *apply
}

func main() {
	prefix, targetPkg, apply := handleFlags()

	// Parse the data into a structured format
	modFile, err := readGomod()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing go.mod: %v\n", err)
		os.Exit(1)
	}

	// Decode the JSON stream.
	// (Using a decoder because 'go list ./...' can output multiple JSON objects)
	matchedDeps, moduleRoot, err := readMatchedDeps(targetPkg, prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading matched dependencies: %v\n", err)
		os.Exit(1)
	}

	// Find and parse the go.work file, then report which workspace modules
	// contain the matched dependencies.
	workModules, workDir, err := readGowork()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing workspace: %v\n", err)
		os.Exit(1)
	}

	if len(matchedDeps) == 0 {
		fmt.Printf("No transitively included packages found matching prefix: %s\n", prefix)
		return
	}

	fmt.Printf("Transitive dependencies with prefix '%s':\n", prefix)
	for dep := range matchedDeps {
		fmt.Printf("  - %s\n", dep)
	}

	// For each matched dep, find which workspace module it belongs to.
	inWorkspace := make(map[string]string) // module path → local directory
	for dep := range matchedDeps {
		for modPath, dir := range workModules {
			if dep == modPath || strings.HasPrefix(dep, modPath+"/") {
				inWorkspace[modPath] = dir
			}
		}
	}

	if len(inWorkspace) == 0 {
		fmt.Println("\nNo matched dependencies belong to workspace modules.")
		return
	}

	if err := processDependencies(inWorkspace, workDir, modFile, moduleRoot, apply); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing dependencies: %v\n", err)
		os.Exit(1)
	}
}

func processDependencies(inWorkspace map[string]string, workDir string, modFile *modfile.File, moduleRoot string, apply bool) error {
	var errs []error
	for modPath, dir := range inWorkspace {
		isInRepo, err := isSubdir(workDir, dir)
		if err != nil {
			errs = append(errs, fmt.Errorf("error checking paths for %s: %w", dir, err))
			continue
		}
		if isInRepo && modFile.Module.Mod.Path != modPath {
			rp, err := filepath.Rel(moduleRoot, dir)
			if err != nil {
				errs = append(errs, fmt.Errorf("error computing relative path for %s: %w", dir, err))
				continue
			}

			if apply {
				output, err := exec.Command("go", "mod", "edit", "--replace", fmt.Sprintf("%s=%s", modPath, rp)).CombinedOutput()
				if err != nil {
					errs = append(errs, fmt.Errorf("error applying replace directive for %s: %w (output: %s)", modPath, err, string(output)))
				} else {
					fmt.Printf("Applied: go mod edit --replace %s=%s\n", modPath, rp)
				}
			} else {
				fmt.Printf("go mod edit --replace %s=%s\n", modPath, rp)
			}
		}
	}
	return errors.Join(errs...)
}

func readGomod() (*modfile.File, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current directory: %w", err)
	}
	data, err := os.ReadFile(filepath.Join(cwd, "go.mod"))
	if err != nil {
		return nil, fmt.Errorf("error reading go.mod: in %s: %w", cwd, err)
	}

	// Parse the data into a structured format
	modfile, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return nil, fmt.Errorf("error parsing go.mod: %w", err)
	}
	return modfile, nil
}

func readGowork() (map[string]string, string, error) {
	workFile, err := findGoWork()
	if err != nil {
		return nil, "", fmt.Errorf("no go.work file found: %w", err)
	}

	data, err := os.ReadFile(workFile)
	if err != nil {
		return nil, "", fmt.Errorf("error reading %s: %w", workFile, err)
	}

	workModules, err := workspaceModules(workFile, data)
	if err != nil {
		return nil, "", fmt.Errorf("error parsing workspace: %w", err)
	}

	return workModules, filepath.Dir(workFile), nil
}

// workspaceModules parses a go.work file and returns a map of
// module path → absolute local directory for each use directive.
func workspaceModules(workFile string, workData []byte) (map[string]string, error) {
	workFileParsed, err := modfile.ParseWork(workFile, workData, nil)
	if err != nil {
		return nil, err
	}
	workDir := filepath.Dir(workFile)
	modules := make(map[string]string, len(workFileParsed.Use))
	for _, use := range workFileParsed.Use {
		abs := filepath.Join(workDir, use.Path)
		modPath, err := modulePathFromGoMod(abs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			continue
		}
		modules[modPath] = abs
	}
	return modules, nil
}

// findGoWork returns the path to the active go.work file via 'go env GOWORK'.
func findGoWork() (string, error) {
	out, err := exec.Command("go", "env", "GOWORK").Output()
	if err != nil {
		return "", fmt.Errorf("go env GOWORK: %w", err)
	}
	path := strings.TrimSpace(string(out))
	if path == "" || path == "off" {
		return "", fmt.Errorf("no active go.work file")
	}
	return path, nil
}

func readMatchedDeps(targetPkg string, prefix string) (map[string]bool, string, error) {

	// Run 'go list -json' to get the package and its transitive dependencies.
	cmd := exec.Command("go", "list", "-json", targetPkg)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, "", fmt.Errorf("error running 'go list': %v\n%s", err, stderr.String())
	}

	// Decode the JSON stream.
	// (Using a decoder because 'go list ./...' can output multiple JSON objects)
	decoder := json.NewDecoder(&stdout)
	matchedDeps := make(map[string]bool)
	var pkg Package

	for decoder.More() {
		if err := decoder.Decode(&pkg); err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding JSON: %v\n", err)
			os.Exit(1)
		}

		if strings.HasPrefix(pkg.ImportPath, prefix) {
			matchedDeps[pkg.ImportPath] = true
		}

		for _, dep := range pkg.Deps {
			if strings.HasPrefix(dep, prefix) {
				matchedDeps[dep] = true
			}
		}
	}
	return matchedDeps, pkg.Root, nil
}

// modulePathFromGoMod reads the module path from the go.mod file in dir.
func modulePathFromGoMod(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("read go.mod in %s: %w", dir, err)
	}
	f, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return "", fmt.Errorf("parse go.mod in %s: %w", dir, err)
	}
	if f.Module == nil {
		return "", fmt.Errorf("no module directive in %s/go.mod", dir)
	}
	return f.Module.Mod.Path, nil
}

func isSubdir(parent, child string) (bool, error) {
	absParent, err := filepath.Abs(parent)
	if err != nil {
		return false, err
	}
	absChild, err := filepath.Abs(child)
	if err != nil {
		return false, err
	}
	rel, err := filepath.Rel(absParent, absChild)
	if err != nil {
		return false, err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false, nil
	}
	return true, nil
}
