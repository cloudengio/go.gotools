# Package [cloudeng.io/go/cmd/goannotate/annotators/internal](https://pkg.go.dev/cloudeng.io/go/cmd/goannotate/annotators/internal?tab=doc)
[![CircleCI](https://circleci.com/gh/cloudengio/go.gotools.svg?style=svg)](https://circleci.com/gh/cloudengio/go.gotools) [![Go Report Card](https://goreportcard.com/badge/cloudeng.io/go/cmd/goannotate/annotators/internal)](https://goreportcard.com/report/cloudeng.io/go/cmd/goannotate/annotators/internal)

```go
import cloudeng.io/go/cmd/goannotate/annotators/internal
```


## Constants

### DocTagName
```go
DocTagName = "annotator"

```
DocTagName is the struct tag used to document annotator configuration
fields.



## Functions
### Func DelegatedYAML
```go
func DelegatedYAML(v interface{}, unmarshal func(interface{}) error) error
```
DelegatedYAML will unmarshal the yaml configuration into a yaml.MapSlice and
named fields. Given:

    struct {
      YAML.MapSlice
      Type string `yaml:"type"`
    }

it will unmarshal the entire config into the MapSlice and if any of the
fields in MapSlice have a key 'type', the value for that key will assigned
to the Type field.

### Func Indent
```go
func Indent(text string, indent int) string
```

### Func MustDescribe
```go
func MustDescribe(t interface{}, detail string) string
```

### Func RemarshalYAML
```go
func RemarshalYAML(v yaml.MapSlice, unmarshal func(buf []byte) error) error
```
RemarshalYAML will marshal the supplied yaml.MapSlice to a buf and then
invoke the supplied unmarshal function.



