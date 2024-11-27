// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package convert

import (
	"fmt"
	"reflect"

	"github.com/zclconf/go-cty/cty"
)

func ToCtyValue(val reflect.Value, want cty.Type) (cty.Value, error) {
	return toCtyValue(val, want, nil)
}

func toCtyValue(in reflect.Value, want cty.Type, path Path) (cty.Value, error) {
	if in.IsZero() {
		return cty.NullVal(want), nil
	}

	if in.Type().Kind() == reflect.Pointer {
		// unpack pointers
		in = in.Elem()
	}

	switch want {
	case cty.Bool:
		return cty.BoolVal(in.Interface().(bool)), nil
	case cty.Number:
		switch in.Type().Kind() {
		case reflect.Int:
			return cty.NumberIntVal(int64(in.Interface().(int))), nil
		case reflect.Int8:
			return cty.NumberIntVal(int64(in.Interface().(int8))), nil
		case reflect.Int16:
			return cty.NumberIntVal(int64(in.Interface().(int16))), nil
		case reflect.Int32:
			return cty.NumberIntVal(int64(in.Interface().(int32))), nil
		case reflect.Int64:
			return cty.NumberIntVal(in.Interface().(int64)), nil
		case reflect.Uint:
			return cty.NumberIntVal(int64(in.Interface().(uint))), nil
		case reflect.Uint8:
			return cty.NumberIntVal(int64(in.Interface().(uint8))), nil
		case reflect.Uint16:
			return cty.NumberIntVal(int64(in.Interface().(uint16))), nil
		case reflect.Uint32:
			return cty.NumberIntVal(int64(in.Interface().(uint32))), nil
		case reflect.Uint64:
			return cty.NumberIntVal(int64(in.Interface().(uint64))), nil
		case reflect.Float32:
			return cty.NumberFloatVal(float64(in.Interface().(float32))), nil
		case reflect.Float64:
			return cty.NumberFloatVal(in.Interface().(float64)), nil
		}
	case cty.String:
		return cty.StringVal(in.Interface().(string)), nil
	}

	switch {
	case want.IsCollectionType():
		switch {
		case want.IsListType():
			out := make([]cty.Value, in.Len())
			for i := 0; i < in.Len(); i++ {
				value, err := toCtyValue(in.Index(i), want.ElementType(), path.WithIndex(fmt.Sprintf("%d", i)))
				if err != nil {
					return cty.NullVal(want), err
				}
				out[i] = value
			}
			if len(out) == 0 {
				return cty.ListValEmpty(want.ElementType()), nil
			}
			return cty.ListVal(out), nil
		case want.IsMapType():
			out := make(map[string]cty.Value)
			for _, key := range in.MapKeys() {
				value, err := toCtyValue(in.MapIndex(key), want.ElementType(), path.WithIndex(fmt.Sprintf("%q", key)))
				if err != nil {
					return cty.NullVal(want), err
				}
				out[key.String()] = value
			}
			if len(out) == 0 {
				return cty.MapValEmpty(want.ElementType()), nil
			}
			return cty.MapVal(out), nil
		case want.IsSetType():
			// this shouldn't be possible, so we'll let it panic later
		}
	case want.IsObjectType():
		out := make(map[string]cty.Value)
		for i := 0; i < in.NumField(); i++ {
			attr := in.Type().Field(i).Tag.Get("cty")
			if len(attr) == 0 {
				// skip untagged fields
				continue
			}
			value, err := toCtyValue(in.Field(i), want.AttributeType(attr), path.Append(attr))
			if err != nil {
				return cty.NullVal(want), err
			}
			out[attr] = value
		}
		return cty.ObjectVal(out), nil
	case want.IsTupleType():
		// this shouldn't be possible, so we'll let it panic later
	}
	panic(fmt.Errorf("unsupported type: %s", want.FriendlyName()))
}
