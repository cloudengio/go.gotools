# [derive](https://pkg.go.dev/cloudeng.io/go/derive?tab=doc)
[![CircleCI](https://circleci.com/gh/cloudengio/go.gotools.svg?style=svg)](https://circleci.com/gh/cloudengio/go.gotools) [![Go Report Card](https://goreportcard.com/badge/cloudeng.io/go/derive)](https://goreportcard.com/report/cloudeng.io/go/derive)

```go
import cloudeng.io/go/derive
```

Package derive provides support for deriving go code from existing code.
Functions are provided for obtaining the text representations of existing go
types and function signatures that can be used when generating annotations.

## Constants

### ContextType
```go
ContextType = "context.Context"

```
ContextType is the standard go context type.



## Functions
### Func ArgsForParams
```go
func ArgsForParams(signature *types.Signature, ignoreAtPosition ...int) (format string, arguments []string)
```
ArgsForParams returns the format and arguments to use to log the function's
arguments. The option ignoreAtPosition arguments specify that those
positions should be ignored altogether. This is useful for handling
context.Context like arguments which need often need to be handled
separately.

### Func ArgsForResults
```go
func ArgsForResults(signature *types.Signature) (format string, arguments []string)
```
ArgsForResults returns the format and arguments to use to log the function's
results.

### Func FormatForVar
```go
func FormatForVar(v *types.Var) (string, string)
```
FormatForVar determines an appropriate format spec and argument for a single
function argument or result. The format spec is intended to be passed to a
fmt style logging function. It takes care to ensure that the log output is
bounded as follows:

    1. strings and types that implement stringer are printed as %.10s
    2. slices and maps have only their length printed
    3. errors are printed as %v with no other restrictions
    4. runes are printed as %c, bytes as %02x and pointers are as %02x
    5. for all other types, only the name of the variable is printed

### Func HasContext
```go
func HasContext(signature *types.Signature) (string, bool)
```
HasContext returns true and the name of the first parameter to the function
if that first parameter is context.Context.

### Func HasCustomContext
```go
func HasCustomContext(signature *types.Signature, customContext string) (string, bool)
```
HasCustomContext returns true and the name of the first parameter to the
function if that first parameter is the specified customContext.

### Func ParamAt
```go
func ParamAt(signature *types.Signature, pos int) (varName, typeName string, ok bool)
```
ParamAt returns the name and type of the parameter at pos. It returns false
if no such parameter exists.



