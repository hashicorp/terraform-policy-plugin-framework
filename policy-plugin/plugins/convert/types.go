// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package convert

import (
	"fmt"
	"reflect"

	"github.com/zclconf/go-cty/cty"
)

func ToCtyType(from reflect.Type) (cty.Type, error) {
	return toCtyType(from, nil)
}

func toCtyType(from reflect.Type, path Path) (cty.Type, error) {
	if from.Kind() == reflect.Interface {
		// We can't support interface types because we need to know the concrete
		// type when converting back and forth between Go and cty. Users can
		// implement the conversion themselves if they need this.
		return cty.NilType, withPath(path, fmt.Errorf("interface types not allowed"))
	}

	if from.Kind() == reflect.Ptr {
		// unpack pointers
		from = from.Elem()
	}

	switch from.Kind() {
	case reflect.Bool:
		return cty.Bool, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return cty.Number, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return cty.Number, nil
	case reflect.Float32, reflect.Float64:
		return cty.Number, nil
	case reflect.String:
		return cty.String, nil
	case reflect.Map:
		if key := from.Key(); key.Kind() != reflect.String {
			return cty.NilType, withPath(path, fmt.Errorf("map keys must be strings, but was %s", key.Kind()))
		}
		element, err := toCtyType(from.Elem(), path.WithIndex("\"*\""))
		if err != nil {
			return cty.NilType, err
		}
		return cty.Map(element), nil
	case reflect.Slice:
		element, err := toCtyType(from.Elem(), path.WithIndex("*"))
		if err != nil {
			return cty.NilType, err
		}
		return cty.List(element), nil
	case reflect.Struct:
		fields := make(map[string]cty.Type)
		for i := 0; i < from.NumField(); i++ {
			field := from.Field(i)
			attr := field.Tag.Get("cty")
			if len(attr) == 0 {
				// skip untagged fields
				continue
			}

			path := path.Append(attr)
			if field.PkgPath != "" {
				return cty.NilType, withPath(path, fmt.Errorf("unexported fields not allowed"))
			}
			element, err := toCtyType(field.Type, path)
			if err != nil {
				return cty.NilType, err
			}
			fields[attr] = element
		}
		return cty.Object(fields), nil
	default:
		return cty.NilType, withPath(path, fmt.Errorf("unsupported type %s", from.Kind()))
	}
}
