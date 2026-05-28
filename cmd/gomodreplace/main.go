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
	"bufio"
	"bytes"
	"encoding/json"
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
	matchedDeps, pkg, err := readMatchedDeps(targetPkg, prefix)
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

	fmt.Printf("\nWorkspace modules (from %s) containing matched dependencies:\n", filepath.Join(workDir, "go.work"))

	if err := processDependencies(inWorkspace, workDir, modFile, pkg, apply); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing dependencies: %v\n", err)
		os.Exit(1)
	}
}

func processDependencies(inWorkspace map[string]string, workDir string, modFile *modfile.File, pkg Package, apply bool) error {

	for modPath, dir := range inWorkspace {
		isInRepo, err := hasCommonPrefix(dir, workDir, workDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking paths: %v\n", err)
			continue
		}
		if isInRepo && modFile.Module.Mod.Path != modPath {
			rp, err := filepath.Rel(pkg.Root, dir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error computing relative path: %v\n", err)
				continue
			}

			if apply {
				output, err := exec.Command("go", "mod", "edit", "--replace", fmt.Sprintf("%s=%s", modPath, rp)).CombinedOutput()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error applying replace directive: %v\nOutput: %s\n", err, string(output))
				} else {
					fmt.Printf("Applied: go mod edit --replace %s=%s\n", modPath, rp)
				}
			} else {
				fmt.Printf("go mod edit --replace %s=%s\n", modPath, rp)
			}
		}
	}
	return nil
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

	workDir := filepath.Dir(workFile)
	workModules, err := workspaceModules(workDir, data)
	if err != nil {
		return nil, "", fmt.Errorf("error parsing workspace: %w", err)
	}

	return workModules, workDir, nil
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

func readMatchedDeps(targetPkg string, prefix string) (map[string]bool, Package, error) {

	// Run 'go list -json' to get the package and its transitive dependencies.
	cmd := exec.Command("go", "list", "-json", targetPkg)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, Package{}, fmt.Errorf("error running 'go list': %v\n%s", err, stderr.String())
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
	return matchedDeps, pkg, nil
}

// workspaceModules parses a go.work file and returns a map of
// module path → absolute local directory for each use directive.
func workspaceModules(workDir string, workData []byte) (map[string]string, error) {
	dirs := useDirs(workData)
	modules := make(map[string]string, len(dirs))
	for _, rel := range dirs {
		abs := filepath.Join(workDir, rel)
		modPath, err := modulePathFromGoMod(abs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			continue
		}
		modules[modPath] = abs
	}
	return modules, nil
}

// useDirs extracts the directory paths from the use directives in a go.work file.
func useDirs(data []byte) []string {
	var dirs []string
	sc := bufio.NewScanner(bytes.NewReader(data))
	inBlock := false
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if line == "use (" {
			inBlock = true
			continue
		}
		if inBlock {
			if line == ")" {
				inBlock = false
				continue
			}
			dirs = append(dirs, line)
			continue
		}
		if strings.HasPrefix(line, "use ") {
			dirs = append(dirs, strings.TrimPrefix(line, "use "))
		}
	}
	return dirs
}

// modulePathFromGoMod reads the module path from the go.mod file in dir.
func modulePathFromGoMod(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("read go.mod in %s: %w", dir, err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.Fields(line)[1], nil
		}
	}
	return "", fmt.Errorf("no module directive in %s/go.mod", dir)
}

func hasCommonPrefix(path1, path2, prefix string) (bool, error) {
	// 1. Resolve to absolute paths
	abs1, err := filepath.Abs(path1)
	if err != nil {
		return false, err
	}
	abs2, err := filepath.Abs(path2)
	if err != nil {
		return false, err
	}
	absPrefix, err := filepath.Abs(prefix)
	if err != nil {
		return false, err
	}

	// 3. Ensure the prefix is correctly formatted for directory evaluation
	if !strings.HasSuffix(absPrefix, string(filepath.Separator)) {
		absPrefix += string(filepath.Separator)
	}

	if !strings.HasSuffix(abs1, string(filepath.Separator)) {
		abs1 += string(filepath.Separator)
	}
	if !strings.HasSuffix(abs2, string(filepath.Separator)) {
		abs2 += string(filepath.Separator)
	}

	// 4. Verify if both paths share the prefix directory
	return strings.HasPrefix(abs1, absPrefix) && strings.HasPrefix(abs2, absPrefix), nil
}
