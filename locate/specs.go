package locate

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"
)

// parseSpecAndRegexp parses a type/interface spec that allows for a regular
// expression in the package local component.
func parseSpecAndRegexp(typ string) (string, string) {
	idx := strings.LastIndex(typ, "/")
	if idx < 0 {
		if idx := strings.Index(typ, "."); idx >= 0 {
			// <package>.<expr>
			return typ[:idx], typ[idx+1:]
		}
		// No / and no ., treat name as package with wildcard.
		return typ, ".*"
	}
	dir := typ[:idx]
	tail := typ[idx+1:]
	idx = strings.Index(tail, ".")
	if idx < 0 {
		return path.Join(dir, tail), ".*"
	}
	return path.Join(dir, tail[:idx]), tail[idx+1:]
}

func getPathAndRegexp(typ string) (string, *regexp.Regexp, error) {
	path, expr := parseSpecAndRegexp(typ)
	if len(path) == 0 {
		return "", nil, fmt.Errorf("no package name in %v", typ)
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return "", nil, fmt.Errorf("failed to compile regexp: %q: %w", expr, err)
	}
	return path, re, nil
}

func dedup(inputs []string) []string {
	deduped := []string{}
	dedup := map[string]bool{}
	for _, p := range inputs {
		if dedup[p] || (len(p) == 0) {
			continue
		}
		dedup[p] = true
		deduped = append(deduped, p)
	}
	return deduped
}

func packagesFromSpecs(specs []string) []string {
	pkgs := make([]string, 0, len(specs))
	for _, spec := range specs {
		path, _ := parseSpecAndRegexp(spec)
		pkgs = append(pkgs, path)
	}
	return pkgs
}

// packagesToLoad extracts the package names from all input specs
// and packages and dedups them in preparation for loading all of them.
func packagesToLoad(ctx context.Context, ifcs, fns, pkgs []string) ([]string, error) {
	all := make([]string, 0, len(ifcs)+len(fns)+len(pkgs))
	all = append(all, packagesFromSpecs(ifcs)...)
	all = append(all, packagesFromSpecs(fns)...)
	all = append(all, pkgs...)
	return dedup(all), nil
}
