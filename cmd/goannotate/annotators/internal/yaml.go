// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package internal

import (
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
)

/*
func IsYAMLKey(v yaml.MapItem, k string) (string, bool) {
	if n, ok := v.Key.(string); ok && n == k {
		tmp, ok := v.Value.(string)
		return tmp, ok
	}
	return "", false
}
*/

var mapSlice reflect.Type

func init() {
	mapSlice = reflect.TypeOf(yaml.MapSlice{})
}

func findMapSlice(typ reflect.Type) (int, error) {
	for nf := 0; nf < typ.NumField(); nf++ {
		field := typ.Field(nf)
		if field.Type.AssignableTo(mapSlice) {
			return nf, nil
		}
	}
	return 0, fmt.Errorf("%T does not contain a yaml.MapSlice", typ)
}

func fieldWithTag(typ reflect.Type, tag string) int {
	for nf := 0; nf < typ.NumField(); nf++ {
		field := typ.Field(nf)
		tags, ok := field.Tag.Lookup("yaml")
		if !ok {
			continue
		}
		parts := strings.Split(tags, ",")
		if len(parts) > 0 && parts[0] == tag {
			return nf
		}

	}
	return -1
}

// DelegatedYAML will unmarshal the yaml configuration into a yaml.MapSlice
// and named fields. Given:
//
//	struct {
//	  YAML.MapSlice
//	  Type string `yaml:"type"`
//	}
//
// it will unmarshal the entire config into the MapSlice and if any of
// the fields in MapSlice have a key 'type', the value for that key will
// assigned to the Type field.
func DelegatedYAML(v interface{}, unmarshal func(interface{}) error) error {
	typ := reflect.TypeOf(v)
	val := reflect.ValueOf(v)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = reflect.Indirect(val)
	}
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("%T is not a struct", v)
	}
	mapSliceField, err := findMapSlice(typ)
	if err != nil {
		return err
	}
	ms := val.Field(mapSliceField)
	if err := unmarshal(ms.Addr().Interface()); err != nil {
		return err
	}
	for i := 0; i < ms.Len(); i++ {
		pair := ms.Index(i)
		name := pair.Field(0).Elem().String()
		if f := fieldWithTag(typ, name); f >= 0 {
			if pair.Field(1).Elem().Type().AssignableTo(typ.Field(f).Type) {
				val.Field(f).Set(pair.Field(1).Elem())
			}
		}
	}
	return nil
}

// RemarshalYAML will marshal the supplied yaml.MapSlice to a buf and then invoke
// the supplied unmarshal function.
func RemarshalYAML(v yaml.MapSlice, unmarshal func(buf []byte) error) error {
	buf, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	return unmarshal(buf)
}
