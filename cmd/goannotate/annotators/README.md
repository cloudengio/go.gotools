# annotators
--
    import "cloudeng.io/go/cmd/goannotate/annotators"

go:generate go run github.com/robertkrimen/godocdown/godocdown -o README.md
cloudeng.io/go/cmd/goannotate/annotators

## Usage

```go
const TagName = "annotator"
```

```go
var (
	Verbose = false
	Trace   = false
)
```

#### func  Available

```go
func Available() []string
```
Available lists all available annotations.

#### func  Describe

```go
func Describe(t interface{}, msg string) (string, error)
```

#### func  DescribeX

```go
func DescribeX(name string) string
```
DescribeX returns the description for the annotator or annotation.

#### func  MustDescribe

```go
func MustDescribe(t interface{}, msg string) string
```

#### func  Register

```go
func Register(annotator Annotator)
```
Register registers a new annotator.

#### func  Registered

```go
func Registered() []string
```
Registered lists all registered annotatators.

#### func  Verbosef

```go
func Verbosef(format string, args ...interface{})
```
Verbosef is like fmt.Printf but will produce output if the Verbose variable is
true.

#### type AddLogCall

```go
type AddLogCall struct {
	Type                 string   `annotator:"name of annotator type."`
	Name                 string   `annotator:"name of annotation."`
	Packages             []string `annotator:"packages to be annotated"`
	Interfaces           []string `annotator:"list of interfaces whose implementations are to have logging calls added to them."`
	Functions            []string `annotator:"list of functionms that are to have function calls added to them."`
	ContextType          string   `yaml:"contextType" annotator:"type for the context parameter and result."`
	Import               string   `annotator:"import patrh for the logging function."`
	Logcall              string   `annotator:"invocation for the logging function."`
	IgnoreEmptyFunctions bool     `yaml:"ignoreEmptyFunctions" annotator:"if set empty functions are ignored."`
	Concurrency          int      `annotator:"the number of goroutines to use, zero for a sensible default."`

	// Used for templates.
	FunctionName string `yaml:",omitempty"`
	Tag          string `yaml:",omitempty"`
	ContextParam string `yaml:",omitempty"`
	Location     string `yaml:",omitempty"`
	Params       string `yaml:",omitempty"`
	Results      string `yaml:",omitempty"`
}
```

AddLogCall represents an annotator for adding a function call that logs the
entry and exit to every function and method that is matched by the locator.

#### func (*AddLogCall) Describe

```go
func (lc *AddLogCall) Describe() string
```
Describe implements annotators.Annotation.

#### func (*AddLogCall) Do

```go
func (lc *AddLogCall) Do(ctx context.Context, root string, pkgs []string) error
```
Do implements annotators.Annotation.

#### func (*AddLogCall) New

```go
func (lc *AddLogCall) New(name string) Annotation
```
New implements annotators.Annotator.

#### func (*AddLogCall) UnmarshalYAML

```go
func (lc *AddLogCall) UnmarshalYAML(buf []byte) error
```
UnmarshalYAML implements annotators.Annotation.

#### type Annotation

```go
type Annotation interface {
	// UnmarshalYAML unmarshals the annotator's yaml configuration.
	UnmarshalYAML(buf []byte) error
	// Do runs the annotator. Root specifies an alternate location for the
	// modified files, if it is empty, files are modified in place. The original
	// directory structure will be mirrored under root.
	// Packages is the set of packages to be annotated as requested on the
	// command line and which overrides any configured ones.
	Do(ctx context.Context, root string, packages []string) error
	Describe() string
}
```

Annotation represents a configured instance of an Annotator.

#### func  Lookup

```go
func Lookup(name string) Annotation
```
Lookup returns the annotation with the specifed typeName, if any.

#### type Annotator

```go
type Annotator interface {
	// New creates a new instance of T. It used to create new configurations
	// as specified in config files, namely an 'annotation'
	New(name string) Annotation
	Describe() string
}
```

Annotator represents the interface that all annotators must implement.

#### type EnsureCopyrightAndLicense

```go
type EnsureCopyrightAndLicense struct {
	Type        string   `annotator:"name of annotator type."`
	Name        string   `annotator:"name of annotation."`
	Packages    []string `annotator:"packages to be annotated"`
	Copyright   string   `annotator:"desired copyright notice."`
	License     string   `annotator:"desired license notice."`
	Concurrency int      `annotator:"the number of goroutines to use, zero for a sensible default."`
}
```


#### func (*EnsureCopyrightAndLicense) Describe

```go
func (ec *EnsureCopyrightAndLicense) Describe() string
```
Describe implements annotators.Annotations.

#### func (*EnsureCopyrightAndLicense) Do

```go
func (ec *EnsureCopyrightAndLicense) Do(ctx context.Context, root string, pkgs []string) error
```
Do implements annotators.Annotations.

#### func (*EnsureCopyrightAndLicense) New

```go
func (ec *EnsureCopyrightAndLicense) New(name string) Annotation
```
New implements annotators.Annotators.

#### func (*EnsureCopyrightAndLicense) UnmarshalYAML

```go
func (ec *EnsureCopyrightAndLicense) UnmarshalYAML(buf []byte) error
```
UnmarshalYAML implements annotators.Annotations.

#### type RmLogCall

```go
type RmLogCall struct {
	Type        string   `annotator:"name of annotator type."`
	Name        string   `annotator:"name of annotation."`
	Packages    []string `annotator:"packages to be annotated"`
	Interfaces  []string `annotator:"list of interfaces whose implementations are to have logging function calls removed from."`
	Functions   []string `annotator:"list of functionms that are to have function calls removed from."`
	Logcall     string   `annotator:"the logging function call to be removed"`
	Comment     string   `annotator:"optional comment that must appear in the comments associated with the function call if it is to be removed."`
	Deferred    bool     `annotator:"if set requires that the function to be removed must be defered."`
	Concurrency int      `annotator:"the number of goroutines to use, zero for a sensible default."`
}
```


#### func (*RmLogCall) Describe

```go
func (rc *RmLogCall) Describe() string
```
Describe implements annotators.Annotation.

#### func (*RmLogCall) Do

```go
func (rc *RmLogCall) Do(ctx context.Context, root string, pkgs []string) error
```
Do implements annotators.Annotation.

#### func (*RmLogCall) New

```go
func (rc *RmLogCall) New(name string) Annotation
```
New implements annotators.Annotator.

#### func (*RmLogCall) UnmarshalYAML

```go
func (rc *RmLogCall) UnmarshalYAML(buf []byte) error
```
UnmarshalYAML implements annotators.Annotation.

#### type Spec

```go
type Spec struct {
	yaml.MapSlice
	Name string // Name identifies a particular configuration of an annotator type.
	Type string // Type identifies the annotation to be peformed.
}
```

Spec represents the yaml configuration for an annotation. It has a common field
for the type and name of the annotator but all other fields are delegated to the
Unmarshal method of the annotator specuifed by the Type field.

#### func (*Spec) UnmarshalYAML

```go
func (s *Spec) UnmarshalYAML(unmarshal func(interface{}) error) error
```
UnmarshalYAML implements yaml.Unmarshaler.
