# Package [cloudeng.io/go/cmd/goannotate/annotators](https://pkg.go.dev/cloudeng.io/go/cmd/goannotate/annotators?tab=doc)
[![CircleCI](https://circleci.com/gh/cloudengio/go.gotools.svg?style=svg)](https://circleci.com/gh/cloudengio/go.gotools) [![Go Report Card](https://goreportcard.com/badge/cloudeng.io/go/cmd/goannotate/annotators)](https://goreportcard.com/report/cloudeng.io/go/cmd/goannotate/annotators)

```go
import cloudeng.io/go/cmd/goannotate/annotators
```


## Constants

### TagName
```go
TagName = "annotator"

```
TagName is the struct tag used to document annotator configuration fields.



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

### Func Describe
```go
func Describe(t interface{}, msg string) (string, error)
```
Describe generates a description for the supplied type based on its struct
tags.

### Func Description
```go
func Description(name string) string
```
Description returns the description for the annotator or annotation.

### Func MustDescribe
```go
func MustDescribe(t interface{}, msg string) string
```
MustDescribe is like describe except that panics on an error.

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
	Type                string   `annotator:"name of annotator type."`
	Name                string   `annotator:"name of annotation."`
	Packages            []string `annotator:"packages to be annotated"`
	Interfaces          []string `annotator:"list of interfaces whose implementations are to have logging calls added to them."`
	Functions           []string `annotator:"list of functionms that are to have function calls added to them."`
	ContextType         string   `yaml:"contextType" annotator:"type for the context parameter and result."`
	Import              string   `annotator:"import patrh for the logging function."`
	Logcall             string   `annotator:"invocation for the logging function."`
	AtLeastStatements   int      `yaml:"atLeastStatements" annotator:"the number of statements that must be present in a function in order for it to be annotated."`
	NoAnnotationComment string   `yaml:"noAnnotationComment" annotator:"do not annotate functions that contain this comment"`
	Concurrency         int      `annotator:"the number of goroutines to use, zero for a sensible default."`

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
	Type        string   `annotator:"name of annotator type."`
	Name        string   `annotator:"name of annotation."`
	Packages    []string `annotator:"packages to be annotated"`
	Copyright   string   `annotator:"desired copyright notice."`
	License     string   `annotator:"desired license notice."`
	Concurrency int      `annotator:"the number of goroutines to use, zero for a sensible default."`
}
```
EnsureCopyrightAndLicense represents an annotator that can insert or replace
copyright and license headers from go source code files.

### Type RmLogCall
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
RmLogCall represents an annotor for removing logging calls.

### Type Spec
```go
type Spec struct {
	yaml.MapSlice
	Name string // Name identifies a particular configuration of an annotator type.
	Type string // Type identifies the annotation to be performed.
}
```
Spec represents the yaml configuration for an annotation. It has a common
field for the type and name of the annotator but all other fields are
delegated to the Unmarshal method of the annotator specuifed by the Type
field.



