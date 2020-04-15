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
	annotators     = map[string]Annotator{}
	configurations = map[string]Annotation{}
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

// Annotator represents the interface that all annotators must implement.
type Annotator interface {
	// New creates a new instance of T. It used to create new configurations
	// as specified in config files, namely an 'annotation'
	New(name string) Annotation
	Describe() string
}

// Annotation represents a configured instance of an Annotator.
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

// Register registers a new annotator.
func Register(annotator Annotator) {
	typ := reflect.TypeOf(annotator).Elem()
	name := typ.PkgPath() + "." + typ.Name()
	annotators[name] = annotator
}

// Registered lists all registered annotatators.
func Registered() []string {
	av := []string{}
	for k := range annotators {
		av = append(av, k)
	}
	sort.Strings(av)
	return av
}

// DescribeX returns the description for the annotator or annotation.
func DescribeX(name string) string {
	if an, ok := configurations[name]; ok {
		return an.Describe()

	}
	if an, ok := annotators[name]; ok {
		return an.Describe()
	}
	return ""
}

// Available lists all available annotations.
func Available() []string {
	av := []string{}
	for k := range configurations {
		av = append(av, k)
	}
	sort.Strings(av)
	return av
}

// Lookup returns the annotation with the specifed typeName, if any.
func Lookup(name string) Annotation {
	return configurations[name]
}

// Spec represents the yaml configuration for an annotation. It has a common
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
	annotator := annotators[s.Type]
	if annotator == nil {
		return fmt.Errorf("failed to find an implementation for %s", s.Type)
	}
	annotation := annotator.New(s.Name)
	buf, err := yaml.Marshal(s.MapSlice)
	if err != nil {
		return fmt.Errorf("failed to marshal mapslice for %s", s.Type)
	}
	if err := annotation.UnmarshalYAML(buf); err != nil {
		return err
	}
	if _, ok := configurations[s.Name]; ok {
		return fmt.Errorf("annotator configuration %v already exists", s.Name)
	}
	configurations[s.Name] = annotation
	return nil
}
