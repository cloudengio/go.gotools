// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package locate provides a means for obtaining the location of comments,
// functions and implementations of interfaces in go source code,
// with a view to annotating that source code programmatically.
package locate

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"os/exec"
	"sort"
	"strings"
	"sync"

	"cloudeng.io/errors"
	"cloudeng.io/sync/errgroup"
	"golang.org/x/tools/go/packages"
)

type traceFunc func(string, ...interface{})

// T represents the ability to locate functions and interface implementations.
type T struct {
	options                options
	loader                 *loader
	interfacePackages      []string
	functionPackages       []string
	implementationPackages []string
	commentExpressions     []string

	mu sync.Mutex

	// GUARDED_BY(mu), indexed by <package-path>.<name>
	interfaces map[string]interfaceDesc
	// GUARDED_BY(mu), indexed by types.Func.FullName() which includes
	// the receiver and hence is unique.
	functions map[string]funcDesc
	// GUARDED_BY(mu), indexed by the regular expression that matched them.
	comments map[string][]commentDesc
	// GUARDED_BY(mu), indexed by filename.
	dirty map[string]HitMask
}

// HitMask encodes the type of object found in a given file.
type HitMask int

const (
	// HasComment is set if the current file contains a comment.
	HasComment HitMask = 1 << iota
	// HasFunction is set if the current file contains a function.
	HasFunction
	// HasInterface is set if the current file contains an interface.
	HasInterface
	hitSentinel
)

var hitNames = []string{
	"comment",
	"function",
	"interface",
}

func (hm HitMask) String() string {
	parts := []string{}
	hit := 0
	for {
		mask := 1 << hit
		if mask == int(hitSentinel) {
			break
		}
		if (mask & int(hm)) != 0 {
			parts = append(parts, hitNames[hit])
		}
		hit++
	}
	return strings.Join(parts, ", ")
}

type options struct {
	concurrency               int
	tests                     bool
	ignoreMissingFunctionsEtc bool
	includeMethods            bool
	trace                     func(string, ...interface{})
}

// Option represents an option for controlling the behaviour of
// locate.T instances.
type Option func(*options)

// Concurrency sets the number of goroutines to use. 0 implies no limit.
func Concurrency(c int) Option {
	return func(o *options) {
		o.concurrency = c
	}
}

// Trace sets a trace function.
func Trace(fn func(string, ...interface{})) Option {
	return func(o *options) {
		o.trace = fn
	}
}

// IgnoreMissingFuctionsEtc prevents errors due to packages not containing
// any exported matching interfaces and functions.
func IgnoreMissingFuctionsEtc() Option {
	return func(o *options) {
		o.ignoreMissingFunctionsEtc = true
	}
}

// IncludeTests includes test code from all requested packages.
func IncludeTests() Option {
	return func(o *options) {
		o.tests = true
	}
}

// IncludeMethods controls if methods as well as functions are matched against
// the function spec.
func IncludeMethods(val bool) Option {
	return func(o *options) {
		o.includeMethods = val
	}
}

// New returns a new instance of T.
func New(options ...Option) *T {
	t := &T{
		interfaces: make(map[string]interfaceDesc),
		functions:  make(map[string]funcDesc),
		dirty:      make(map[string]HitMask),
		comments:   make(map[string][]commentDesc),
	}
	t.loader = newLoader(t.trace)
	for _, fn := range options {
		fn(&t.options)
	}
	return t
}

func (t *T) trace(format string, args ...interface{}) {
	if t.options.trace == nil {
		return
	}
	t.options.trace(format, args...)
}

// AddInterfaces adds interfaces whose implementations are to be located.
// The interface names are specified as fully qualified type names with a
// regular expression being accepted for the package local component, or
// as a go list expression (ie. one that starts with ./ or contains '...').
// For example, all of the following match all interfaces in
// acme.com/a/b:
//
//	acme.com/a/b
//	acme.com/a/b.
//	acme.com/a/b..*
//
// Note that the . separator in the type name is not used as part of the
// regular expression. The following will match a subset of the interfaces:
//
//	acme.com/a/b.prefix
//	acme.com/a/b.thisInterface$
//
// Note that the two forms 'go list' and <package>.<regex> cannot be combined.
func (t *T) AddInterfaces(interfaces ...string) {
	t.interfacePackages = append(t.interfacePackages, interfaces...)
}

// AddFunctions adds functions to be located. The function names are specified
// as fully qualified names with a regular expression being accepted for the
// package local component as per AddInterfaces.
func (t *T) AddFunctions(functions ...string) {
	t.functionPackages = append(t.functionPackages, functions...)
}

// AddPackages adds packages that will be searched for implementations
// of interfaces specified via AddInterfaces.
func (t *T) AddPackages(packages ...string) {
	t.implementationPackages = append(t.implementationPackages, packages...)
}

// AddComments adds regular expressions to be matched against the contents
// of comments.
func (t *T) AddComments(comments ...string) {
	t.commentExpressions = append(t.commentExpressions, comments...)
}

// Do locates implementations of previously added interfaces and functions.
func (t *T) Do(ctx context.Context) error {
	errs := errors.M{}
	interfaces, err := listPackagesOrSpecs(ctx, t.interfacePackages)
	errs.Append(err)
	functions, err := listPackagesOrSpecs(ctx, t.functionPackages)
	errs.Append(err)
	var packages []string
	if len(t.implementationPackages) > 0 {
		packages, err = listPackages(ctx, t.implementationPackages)
		errs.Append(err)
	}
	if err := errs.Err(); err != nil {
		return err
	}

	packages = dedup(packages)
	allPackages, err := packagesToLoad(ctx, interfaces, functions, packages)
	if err != nil {
		return err
	}
	comments := dedup(t.commentExpressions)
	if err := t.loader.loadPaths(allPackages, t.options.tests); err != nil {
		return err
	}
	if err := t.findInterfaces(ctx, interfaces); err != nil {
		return err
	}
	grp, ctx := errgroup.WithContext(ctx)
	grp.GoContext(ctx, func() error {
		return t.findFunctions(ctx, functions)
	})
	grp.GoContext(ctx, func() error {
		return t.findImplementations(ctx, packages)
	})
	grp.GoContext(ctx, func() error {
		return t.findComments(ctx, comments)
	})
	return grp.Wait()
}

// MakeCommentMaps creates a new ast.CommentMap for every processed file.
// CommentMaps are expensive to create and hence should be created once and
// reused.
func (t *T) MakeCommentMaps() map[*ast.File]ast.CommentMap {
	cmaps := map[*ast.File]ast.CommentMap{}
	t.WalkFiles(func(_ string, pkg *packages.Package, _ ast.CommentMap, file *ast.File, _ HitMask) {
		cmaps[file] = ast.NewCommentMap(pkg.Fset, file, file.Comments)

	})
	return cmaps
}

type sortByPos struct {
	name    string
	pos     token.Position
	payload interface{}
}

func sorter(sorted []sortByPos) {
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].pos.Filename == sorted[j].pos.Filename {
			return sorted[i].pos.Offset < sorted[j].pos.Offset
		}
		return sorted[i].pos.Filename < sorted[j].pos.Filename
	})
}

// IsGoListPath returns true if path will be passed to 'go list' to be resolved
// rather than being treated as a <package>.<regex> spec.
func IsGoListPath(path string) bool {
	return strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") ||
		strings.Contains(path, "...")
}

func listPackagesOrSpecs(ctx context.Context, specs []string) ([]string, error) {
	var expanded []string
	var tolist []string
	for _, spec := range specs {
		if IsGoListPath(spec) {
			tolist = append(tolist, spec)
			continue
		}
		expanded = append(expanded, spec)
	}
	if len(tolist) > 0 {
		listed, err := listPackages(ctx, tolist)
		if err != nil {
			return nil, err
		}
		expanded = append(expanded, listed...)
	}
	return dedup(expanded), nil
}

func listPackages(ctx context.Context, packages []string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "go", append([]string{"list"}, packages...)...)
	out, err := cmd.Output()
	if err != nil {
		cl := strings.Join(cmd.Args, ", ")
		exitErr := err.(*exec.ExitError)
		return nil, fmt.Errorf("failed to run %v: %v\n%s", cl, err, exitErr.Stderr)
	}
	parts := strings.Split(string(out), "\n")
	paths := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		paths = append(paths, p)
	}
	return paths, nil
}

// WalkPackages calls the supplied function for each package loaded. The
// function is called in lexicographic order of package path.
func (t *T) WalkPackages(fn func(pkg *packages.Package)) {
	t.loader.walkPackages(fn)
}
