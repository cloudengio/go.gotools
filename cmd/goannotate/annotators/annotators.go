// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package annotators

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"cloudeng.io/cmdutil/flags"
	"cloudeng.io/go/cmd/goannotate/annotators/internal"
	"gopkg.in/yaml.v2"
)

var (
	// Verbose controls verbose logging.
	Verbose        = false
	annotators     = map[string]Annotator{}
	configurations = map[string]Annotation{}
)

// EssentialOptions represents the configuration options required for all
// annotations.
type EssentialOptions struct {
	Type        string   `yaml:"type" annotator:"name of annotator type."`
	Name        string   `yaml:"name" annotator:"name of annotation."`
	Packages    []string `yaml:"packages" annotator:"packages to be annotated"`
	Concurrency int      `yaml:"concurrency" annotator:"the number of goroutines to use, zero for a sensible default."`
}

// LocateOptions represents the configuration options used to locate specific
// interfaces and/or functions.
type LocateOptions struct {
	Interfaces []string `yaml:"interfaces" annotator:"list of interfaces whose implementations are to be annoated."`
	Functions  []string `yaml:"functions" annotator:"list of functions that are to be annotated."`
}

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
	// Describe returns a description for the annotation.
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

// Description returns the description for the annotator or annotation.
func Description(name string) string {
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

// Lookup returns the annotation with the specified typeName, if any.
func Lookup(name string) Annotation {
	return configurations[name]
}

// Spec represents the yaml configuration for an annotation. It has a common
// field for the type and name of the annotator but all other fields are
// delegated to the Unmarshal method of the annotator specuifed by the Type field.
type Spec struct {
	yaml.MapSlice
	Name string `yaml:"name"` // Name identifies a particular configuration of an annotator type.
	Type string `yaml:"type"` // Type identifies the annotation to be performed.
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (s *Spec) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := internal.DelegatedYAML(s, unmarshal); err != nil {
		return err
	}
	if !flags.AllSet(s.Type, s.Name) {
		return fmt.Errorf("one of Type or Name not set")
	}
	annotator := annotators[s.Type]
	if annotator == nil {
		return fmt.Errorf("failed to find an annotator for %s", s.Type)
	}
	annotation := annotator.New(s.Name)
	if err := internal.RemarshalYAML(s.MapSlice, annotation.UnmarshalYAML); err != nil {
		return err
	}
	configurations[s.Name] = annotation
	return nil
}
