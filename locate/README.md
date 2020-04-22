# Package [cloudeng.io/go/locate](https://pkg.go.dev/cloudeng.io/go/locate?tab=doc)
[![CircleCI](https://circleci.com/gh/cloudengio/go.gotools.svg?style=svg)](https://circleci.com/gh/cloudengio/go.gotools) [![Go Report Card](https://goreportcard.com/badge/cloudeng.io/go/locate)](https://goreportcard.com/report/cloudeng.io/go/locate)

```go
import cloudeng.io/go/locate
```

Package locate provides a means for obtaining the location of comments,
functions and implementations of interfaces in go source code, with a view
to annotating that source code programmatically.

## Functions
### Func IsGoListPath
```go
func IsGoListPath(path string) bool
```
IsGoListPath returns true if path will be passed to 'go list' to be resolved
rather than being treated as a <package>.<regex> spec.



## Types
### Type HitMask
```go
type HitMask int
```
HitMask encodes the type of object found in a given file.

### Type Option
```go
type Option func(*options)
```
Option represents an option for controlling the behaviour of locate.T
instances.

### Type T
```go
type T struct {
	// contains filtered or unexported fields
}
```
T represents the ability to locate functions and interface implementations.



