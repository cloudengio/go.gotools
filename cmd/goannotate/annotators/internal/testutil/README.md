# Package [cloudeng.io/go/cmd/goannotate/annotators/internal/testutil](https://pkg.go.dev/cloudeng.io/go/cmd/goannotate/annotators/internal/testutil?tab=doc)
[![CircleCI](https://circleci.com/gh/cloudengio/go.gotools.svg?style=svg)](https://circleci.com/gh/cloudengio/go.gotools) [![Go Report Card](https://goreportcard.com/badge/cloudeng.io/go/cmd/goannotate/annotators/internal/testutil)](https://goreportcard.com/report/cloudeng.io/go/cmd/goannotate/annotators/internal/testutil)

```go
import cloudeng.io/go/cmd/goannotate/annotators/internal/testutil
```


## Functions
### Func CompareDiffReports
```go
func CompareDiffReports(t *testing.T, a, b []DiffReport)
```

### Func DiffOneFile
```go
func DiffOneFile(t *testing.T, a, b string) string
```

### Func LocatePackages
```go
func LocatePackages(ctx context.Context, t *testing.T, pkgs ...string) *locate.T
```
LocatePackages runs a locator.T with pkgs as the argument to .AddFunctions
and .AddPackages.

### Func SetupAnnotators
```go
func SetupAnnotators(t *testing.T) (string, func())
```
SetupAnnotators reads ./testdata/config.yaml and initializes the
annototators package, creates a temp directory and a cleanup function to
remove the test directory on test failures.

### Func SetupFunctions
```go
func SetupFunctions(t *testing.T)
```
SetupFunctions reads ./testdata/config.yaml and initializes the
annotators/functions package.



## Types
### Type DiffReport
```go
type DiffReport struct {
	Name string
	Diff string
}
```

### Functions

```go
func DiffMultipleFiles(t *testing.T, a, b []string) []DiffReport
```






