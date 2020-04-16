// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package annotators

import (
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
)

// TagName is the struct tag used to document annotator configuration fields.
const TagName = "annotator"

func formatTags(typ reflect.Type) string {
	names := []string{}
	docs := []string{}
	max := 0
	for nf := 0; nf < typ.NumField(); nf++ {
		field := typ.Field(nf)
		doc, ok := field.Tag.Lookup(TagName)
		if !ok {
			continue
		}
		name := strings.ToLower(field.Name)
		if yamltag, ok := field.Tag.Lookup("yaml"); ok {
			if parts := strings.Split(yamltag, ","); len(parts) > 0 {
				name = parts[0]
			}
		}
		if l := len(name); l > max {
			max = l
		}
		names = append(names, name)
		docs = append(docs, doc)
	}

	out := strings.Builder{}
	for i, name := range names {
		out.WriteString("\t")
		out.WriteString(name)
		out.WriteString(":")
		out.WriteString(strings.Repeat(" ", max-len(name)+1))
		out.WriteString(docs[i])
		out.WriteString("\n")
	}
	return out.String()
}

// MustDescribe is like describe except that panics on an error.
func MustDescribe(t interface{}, msg string) string {
	r, err := Describe(t, msg)
	if err != nil {
		panic(err)
	}
	return r
}

// Describe generates a description for the supplied type based on its
// struct tags.
func Describe(t interface{}, msg string) (string, error) {
	typ := reflect.TypeOf(t)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return "", fmt.Errorf("%T is not a struct", t)
	}
	out := &strings.Builder{}
	out.WriteString(typ.PkgPath() + "." + typ.Name() + ":\n")
	out.WriteString(msg)
	out.WriteString("\nThe available configuration fields are:\n")
	out.WriteString(formatTags(typ))
	enc := yaml.NewEncoder(out)
	err := enc.Encode(t)
	return out.String(), err
}
