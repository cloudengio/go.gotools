# [locateutil](https://pkg.go.dev/cloudeng.io/go/locate/locateutil?tab=doc)
[![CircleCI](https://circleci.com/gh/cloudengio/go.gotools.svg?style=svg)](https://circleci.com/gh/cloudengio/go.gotools) [![Go Report Card](https://goreportcard.com/badge/cloudeng.io/go/locate/locateutil)](https://goreportcard.com/report/cloudeng.io/go/locate/locateutil)

```go
import cloudeng.io/go/locate/locateutil
```

Package locateutil provides utility routines for use with its parent locate
package.

## Functions
### Func CommentGroupBounds
```go
func CommentGroupBounds(comments []*ast.CommentGroup) (first, last token.Pos)
```
CommentGroupBounds returns the lowest and largest token.Pos of any of the
supplied CommentGroups.

### Func CommentGroupsContain
```go
func CommentGroupsContain(comments []*ast.CommentGroup, text string) bool
```
CommentGroupsContain returns if any of the supplied CommentGroups contain
'text'.

### Func FunctionCalls
```go
func FunctionCalls(decl *ast.FuncDecl, callname string, deferred bool) []ast.Node
```
FunctionCalls determines if the supplied function declaration contains a
call 'callname' where callname is either a function name or a selector (eg.
foo.bar). If deferred is true the function call must be defer'ed.

### Func HasBody
```go
func HasBody(decl *ast.FuncDecl) bool
```
HasBody returns true of the function has a body.

### Func ImportBlock
```go
func ImportBlock(file *ast.File) (start, end token.Pos)
```
ImportBlock returns the start and end positions of an import statement or
import block for the supplied file.

### Func InterfaceType
```go
func InterfaceType(typ types.Type) *types.Interface
```
InterfaceType returns the underlying *types.Interface if typ represents an
interface or nil otherwise.

### Func IsAbstract
```go
func IsAbstract(fn *types.Func) bool
```
IsAbstract returns true if the function declaration is abstract.

### Func IsInterfaceDefinition
```go
func IsInterfaceDefinition(pkg *packages.Package, obj types.Object) *types.Interface
```
IsInterfaceDefinition returns the interface type that the suplied object
defines in the specified package, if any. This specifically excludes
embedded types which are defined in other packages and anonymous interfaces.



## Types
### Type FuncDesc
```go
type FuncDesc struct {
	Type     *types.Func
	Abstract bool
	Decl     *ast.FuncDecl
	File     *ast.File
	Position token.Position
	Package  *packages.Package
}
```
FuncDesc represents a function definition, declaration and the file and
position within that file. Decl will be nil if Abstract is true.



