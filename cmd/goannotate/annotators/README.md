# Package [cloudeng.io/go/cmd/goannotate/annotators](https://pkg.go.dev/cloudeng.io/go/cmd/goannotate/annotators?tab=doc)
[![CircleCI](https://circleci.com/gh/cloudengio/go.gotools.svg?style=svg)](https://circleci.com/gh/cloudengio/go.gotools) [![Go Report Card](https://goreportcard.com/badge/cloudeng.io/go/cmd/goannotate/annotators)](https://goreportcard.com/report/cloudeng.io/go/cmd/goannotate/annotators)

```go
import cloudeng.io/go/cmd/goannotate/annotators
```


## Constants

### AddLogCallDescription
```go
AddLogCallDescription = `
AddLogCall is an annotator to add function calls that are intended to log entry and exit from functions. The calls will be added as the first statement in the specified function.
`

```
AddLogCallDescription documents AddLogCall.



## Variables
### Verbose
```go
// Verbose controls verbose logging.
Verbose = false

```



## Functions
### Func Available
```go
func Available() []string
```
Available lists all available annotations.

### Func Description
```go
func Description(name string) string
```
Description returns the description for the annotator or annotation.

### Func Register
```go
func Register(annotator Annotator)
```
Register registers a new annotator.

### Func Registered
```go
func Registered() []string
```
Registered lists all registered annotatators.

### Func Verbosef
```go
func Verbosef(format string, args ...interface{})
```
Verbosef is like fmt.Printf but will produce output if the Verbose variable
is true.



## Types
### Type AddLogCall
```go
type AddLogCall struct {
	EssentialOptions `yaml:",inline"`
	LocateOptions    `yaml:",inline"`

	AtLeastStatements   int            `yaml:"atLeastStatements" annotator:"the number of statements that must be present in a function in order for it to be annotated."`
	NoAnnotationComment string         `yaml:"noAnnotationComment" annotator:"do not annotate functions that contain this comment"`
	CallGenerator       functions.Spec `yaml:"callGenerator" annotator:"the spec for the function call to be generated"`
}
```
AddLogCall represents an annotator for adding a function call that logs the
entry and exit to every function and method that is matched by the locator.

### Type Annotation
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
	// Describe returns a description for the annotation.
	Describe() string
}
```
Annotation represents a configured instance of an Annotator.

### Type Annotator
```go
type Annotator interface {
	// New creates a new instance of T. It used to create new configurations
	// as specified in config files, namely an 'annotation'
	New(name string) Annotation
	Describe() string
}
```
Annotator represents the interface that all annotators must implement.

### Type EnsureCopyrightAndLicense
```go
type EnsureCopyrightAndLicense struct {
	EssentialOptions `yaml:",inline"`

	Copyright string `yaml:"copyright" annotator:"desired copyright notice."`
	License   string `yaml:"license" annotator:"desired license notice."`
}
```
EnsureCopyrightAndLicense represents an annotator that can insert or replace
copyright and license headers from go source code files.

### Type EssentialOptions
```go
type EssentialOptions struct {
	Type        string   `yaml:"type" annotator:"name of annotator type."`
	Name        string   `yaml:"name" annotator:"name of annotation."`
	Packages    []string `yaml:"packages" annotator:"packages to be annotated"`
	Concurrency int      `yaml:"concurrency" annotator:"the number of goroutines to use, zero for a sensible default."`
}
```
EssentialOptions represents the configuration options required for all
annotations.

### Type LocateOptions
```go
type LocateOptions struct {
	Interfaces []string `yaml:"interfaces" annotator:"list of interfaces whose implementations are to be annoated."`
	Functions  []string `yaml:"functions" annotator:"list of functions that are to be annotated."`
}
```
LocateOptions represents the configuration options used to locate specific
interfaces and/or functions.

### Type RmLogCall
```go
type RmLogCall struct {
	EssentialOptions `yaml:",inline"`
	LocateOptions    `yaml:",inline"`

	FunctionNameRE string `yaml:"functionNameRE" annotator:"the function call (regexp) to be removed"`
	Comment        string `yaml:"comment" annotator:"optional comment that must appear in the comments associated with the function call if it is to be removed."`
	Deferred       bool   `yaml:"deferred" annotator:"if set requires that the function to be removed must be defered."`
}
```
RmLogCall represents an annotor for removing logging calls.

### Type Spec
```go
type Spec struct {
	yaml.MapSlice
	Name string `yaml:"name"` // Name identifies a particular configuration of an annotator type.
	Type string `yaml:"type"` // Type identifies the annotation to be performed.
}
```
Spec represents the yaml configuration for an annotation. It has a common
field for the type and name of the annotator but all other fields are
delegated to the Unmarshal method of the annotator specuifed by the Type
field.



