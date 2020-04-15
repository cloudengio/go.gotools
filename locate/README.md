# locate
--
    import "cloudeng.io/go/locate"

Package locate provides a means for obtaining the location of comments,
functions and implementations of interfaces in go source code, with a view to
annotating that source code programmatically.

go:generate go run github.com/robertkrimen/godocdown/godocdown -o README.md
cloudeng.io/go/locate

## Usage

#### type HitMask

```go
type HitMask int
```


```go
const (
	HasComment HitMask = 1 << iota
	HasFunction
	HasInterface
)
```

#### func (HitMask) String

```go
func (hm HitMask) String() string
```

#### type Option

```go
type Option func(*options)
```

Option represents an option for controlling the behaviour of locate.T instances.

#### func  Concurrency

```go
func Concurrency(c int) Option
```
Concurrency sets the number of goroutines to use. 0 implies no limit.

#### func  IgnoreMissingFuctionsEtc

```go
func IgnoreMissingFuctionsEtc() Option
```
IgnoreMissingFuctionsEtc prevents errors due to packages not containing any
exported matching interfaces and functions.

#### func  Trace

```go
func Trace(fn func(string, ...interface{})) Option
```
Trace sets a trace function

#### type T

```go
type T struct {
}
```

T represents the ability to locate functions and interface implementations.

#### func  New

```go
func New(options ...Option) *T
```
New returns a new instance of T.

#### func (*T) AddComments

```go
func (t *T) AddComments(comments ...string)
```
AddComments adds regular expressions to be matched against the contents of
comments.

#### func (*T) AddFunctions

```go
func (t *T) AddFunctions(functions ...string)
```
AddFunctions adds functions to be located. The function names are specified as
fully qualified names with a regular expression being accepted for the package
local component as per AddInterfaces.

#### func (*T) AddInterfaces

```go
func (t *T) AddInterfaces(interfaces ...string)
```
AddInterfaces adds interfaces whose implementations are to be located. The
interface names are specified as fully qualified type names with a regular
expression being accepted for the package local component. For example, all of
the following match all interfaces in acme.com/a/b:

    acme.com/a/b
    acme.com/a/b.
    acme.com/a/b..*

Note that the . separator in the type name is not used as part of the regular
expression. The following will match a subset of the interfaces:

    acme.com/a/b.prefix
    acme.com/a/b.thisInterface$

#### func (*T) AddPackages

```go
func (t *T) AddPackages(packages ...string)
```
AddPackages adds packages that will be searched for implementations of
interfaces specified via AddInterfaces.

#### func (*T) Do

```go
func (t *T) Do(ctx context.Context) error
```
Do locates implementations of previously added interfaces and functions.

#### func (*T) WalkComments

```go
func (t *T) WalkComments(fn func(
	re string,
	absoluteFilename string,
	node ast.Node,
	cg *ast.CommentGroup,
	pkg *packages.Package,
	file *ast.File,
))
```
WalkComments calls the supplied function for each comment that was matched by
the specified regular expressions. The function is called with the absolute
filename, the node that the comment is associated with, the comment and the
packates.Package to which the file belongs and its ast. The function is called
in order of filename and then position within filename.

#### func (*T) WalkFiles

```go
func (t *T) WalkFiles(fn func(
	absoluteFilename string,
	pkg *packages.Package,
	comments ast.CommentMap,
	file *ast.File,
	has HitMask,
))
```
WalkFiles calls the supplied function for each file that contains a located
comment, interface or function, ordered by filename. The function is called with
the absolute file name of the file, the packages.Package to which it belongs and
its ast. The function is called in order of filename and then position within
filename.

#### func (*T) WalkFunctions

```go
func (t *T) WalkFunctions(fn func(
	fullname string,
	pkg *packages.Package,
	file *ast.File,
	fn *types.Func,
	decl *ast.FuncDecl,
	implements []string))
```
WalkFunctions calls the supplied function for each function location, ordered by
filename and then position within file. The function is called with the
packages.Package and ast for the file that contains the function, as well as the
type and declaration of the function and the list of interfaces that implements.
The function is called in order of filename and then position within filename.

#### func (*T) WalkInterfaces

```go
func (t *T) WalkInterfaces(fn func(
	fullname string,
	pkg *packages.Package,
	file *ast.File,
	decl *ast.TypeSpec,
	ifc *types.Interface))
```
WalkInterfaces calls the supplied function for each interface location, ordered
by filename and then position within file. The function is called with the
packages.Package and ast for the file that contains the interface, as well as
the type and declaration of the interface.

#### func (*T) WalkPackages

```go
func (t *T) WalkPackages(fn func(pkg *packages.Package))
```
WalkPackages calls the supplied function for each package loaded. The function
is called in lexicographic order of package path.
