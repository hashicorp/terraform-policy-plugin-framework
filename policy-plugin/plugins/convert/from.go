// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package convert

import (
	"fmt"
	"reflect"

	"github.com/zclconf/go-cty/cty"
)

func FromCtyValue(val cty.Value, target reflect.Type) (reflect.Value, error) {
	return fromCtyValue(val, target, nil)
}

func fromCtyValue(in cty.Value, want reflect.Type, path Path) (reflect.Value, error) {
	if in.IsNull() {
		return reflect.Zero(want), nil
	}

	var pointer *reflect.Type
	if want.Kind() == reflect.Pointer {
		pointer = &want
		want = want.Elem()
	}

	var value reflect.Value

	switch want.Kind() {
	case reflect.Bool:
		value = reflect.ValueOf(in.True())
	case reflect.Int:
		val, _ := in.AsBigFloat().Int64()
		value = reflect.ValueOf(int(val))
	case reflect.Int8:
		val, _ := in.AsBigFloat().Int64()
		value = reflect.ValueOf(int8(val))
	case reflect.Int16:
		val, _ := in.AsBigFloat().Int64()
		value = reflect.ValueOf(int16(val))
	case reflect.Int32:
		val, _ := in.AsBigFloat().Int64()
		value = reflect.ValueOf(int32(val))
	case reflect.Int64:
		val, _ := in.AsBigFloat().Int64()
		value = reflect.ValueOf(val)
	case reflect.Uint:
		val, _ := in.AsBigFloat().Uint64()
		value = reflect.ValueOf(uint(val))
	case reflect.Uint8:
		val, _ := in.AsBigFloat().Uint64()
		value = reflect.ValueOf(uint8(val))
	case reflect.Uint16:
		val, _ := in.AsBigFloat().Uint64()
		value = reflect.ValueOf(uint16(val))
	case reflect.Uint32:
		val, _ := in.AsBigFloat().Uint64()
		value = reflect.ValueOf(uint32(val))
	case reflect.Uint64:
		val, _ := in.AsBigFloat().Uint64()
		value = reflect.ValueOf(val)
	case reflect.Float32:
		val, _ := in.AsBigFloat().Float64()
		value = reflect.ValueOf(float32(val))
	case reflect.Float64:
		val, _ := in.AsBigFloat().Float64()
		value = reflect.ValueOf(val)
	case reflect.String:
		value = reflect.ValueOf(in.AsString())
	case reflect.Slice:
		out := reflect.MakeSlice(reflect.SliceOf(want.Elem()), in.LengthInt(), in.LengthInt())
		for i := 0; i < in.LengthInt(); i++ {
			elem, err := fromCtyValue(in.Index(cty.NumberIntVal(int64(i))), want.Elem(), path.WithIndex(fmt.Sprintf("%d", i)))
			if err != nil {
				return reflect.Zero(want), err
			}
			out.Index(i).Set(elem)
		}
		value = out
	case reflect.Map:
		out := reflect.MakeMapWithSize(want, in.LengthInt())
		elemType := want.Elem()
		for key, value := range in.AsValueMap() {
			elem, err := fromCtyValue(value, elemType, path.WithIndex(fmt.Sprintf("%q", key)))
			if err != nil {
				return reflect.Zero(want), err
			}
			out.SetMapIndex(reflect.ValueOf(key), elem)
		}
		value = out
	case reflect.Struct:
		out := reflect.New(want).Elem()
		for i := 0; i < want.NumField(); i++ {
			field := want.Field(i)
			attr := field.Tag.Get("cty")
			if len(attr) == 0 {
				// skip untagged fields
				continue
			}

			elem, err := fromCtyValue(in.GetAttr(attr), field.Type, path.Append(attr))
			if err != nil {
				return reflect.Zero(want), err
			}
			out.Field(i).Set(elem)
		}
		value = out
	default:
		return reflect.Zero(want), withPath(path, fmt.Errorf("unsupported type %s", want.Kind()))
	}

	if pointer != nil {
		ptr := reflect.New(want)
		ptr.Elem().Set(value)
		return ptr, nil
	}

	return value, nil
}
