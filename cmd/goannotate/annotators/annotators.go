package annotators

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	Verbose    = false
	Trace      = false
	annotators = map[string]T{}
)

func Verbosef(format string, args ...interface{}) {
	if !Verbose {
		return
	}
	out := strings.Builder{}
	out.WriteString(fmt.Sprintf(format, args...))
	fmt.Print(out.String())
}

type T interface {
	Unmarshal(buf []byte) error
	Do(ctx context.Context, packages []string) error
	Describe() string
}

func Register(annotator T) {
	typ := reflect.TypeOf(annotator).Elem()
	name := typ.PkgPath() + "." + typ.Name()
	annotators[name] = annotator
}

func Available() []string {
	av := []string{}
	for k := range annotators {
		av = append(av, k)
	}
	sort.Strings(av)
	return av
}

func Lookup(name string) T {
	return annotators[name]
}

type Spec struct {
	yaml.MapSlice
	Type string
}

func (s *Spec) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(&s.MapSlice); err != nil {
		return err
	}
	for _, v := range s.MapSlice {
		if n, ok := v.Key.(string); ok && n == "type" {
			if tmp, ok := v.Value.(string); ok {
				s.Type = tmp
			}
		}
	}
	if len(s.Type) == 0 {
		return fmt.Errorf("failed to find 'Type' field in %s", s.MapSlice)
	}
	an := Lookup(s.Type)
	if an == nil {
		return fmt.Errorf("failed to find an implementation for %s", s.Type)
	}
	buf, err := yaml.Marshal(s.MapSlice)
	if err != nil {
		return fmt.Errorf("failed to marshal mapslice for %s", s.Type)
	}
	return an.Unmarshal(buf)
}
