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
	Verbose        = false
	Trace          = false
	annotators     = map[string]T{}
	configurations = map[string]T{}
)

// Verbosef is like fmt.Printf but will produce output if the Verbose
// variable is true.
func Verbosef(format string, args ...interface{}) {
	if !Verbose {
		return
	}
	out := strings.Builder{}
	out.WriteString(fmt.Sprintf(format, args...))
	fmt.Print(out.String())
}

// T represents the interface that all annotators must implement.
type T interface {
	New(name string) T
	Unmarshal(buf []byte) error
	Do(ctx context.Context, packages []string) error
	Describe() string
}

// Register registers a new annotator.
func Register(annotator T) {
	typ := reflect.TypeOf(annotator).Elem()
	name := typ.PkgPath() + "." + typ.Name()
	annotators[name] = annotator
}

// Available lists the types of all available annotators.
func Available() []string {
	av := []string{}
	for k := range configurations {
		av = append(av, k)
	}
	sort.Strings(av)
	return av
}

// Lookup returns the annotator with the specifed typeName, if any.
func Lookup(typeName string) T {
	return configurations[typeName]
}

// Spec represents the yaml configuration for all annotators. It has a common
// field for the type and name of the annotator but all other fields are
// delegated to the Unmarshal method of the annotator specuifed by the Type field.
type Spec struct {
	yaml.MapSlice
	Name string // Name identifies a particular configuration of an annotator type.
	Type string // Type identifies the annotation to be peformed.
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (s *Spec) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(&s.MapSlice); err != nil {
		return err
	}
	isKey := func(v yaml.MapItem, k string) (string, bool) {
		if n, ok := v.Key.(string); ok && n == k {
			tmp, ok := v.Value.(string)
			return tmp, ok
		}
		return "", false
	}
	for _, v := range s.MapSlice {
		if tmp, ok := isKey(v, "type"); ok {
			s.Type = tmp
		}
		if tmp, ok := isKey(v, "name"); ok {
			s.Name = tmp
		}
	}
	if len(s.Type) == 0 {
		return fmt.Errorf("failed to find 'Type' field in %s", s.MapSlice)
	}
	an := annotators[s.Type]
	if an == nil {
		return fmt.Errorf("failed to find an implementation for %s", s.Type)
	}
	an = an.New(s.Name)
	buf, err := yaml.Marshal(s.MapSlice)
	if err != nil {
		return fmt.Errorf("failed to marshal mapslice for %s", s.Type)
	}
	if err := an.Unmarshal(buf); err != nil {
		return err
	}
	if _, ok := configurations[s.Name]; ok {
		return fmt.Errorf("annotator configuration %v already exists", s.Name)
	}
	configurations[s.Name] = an
	return nil
}
