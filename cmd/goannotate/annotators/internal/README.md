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

### Func Indent
```go
func Indent(text string, indent int) string
```

### Func IsYAMLKey
```go
func IsYAMLKey(v yaml.MapItem, k string) (string, bool)
```

### Func MustDescribe
```go
func MustDescribe(t interface{}, detail string) string
```

### Func RemarshalYAML
```go
func RemarshalYAML(v yaml.MapSlice, unmarshal func(buf []byte) error) error
```



