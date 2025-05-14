// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package plugins

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

type structure struct {
	Field string `cty:"field"`
}

type structureWithPointer struct {
	Field *string `cty:"field"`
}

func TestRegisterFunction(t *testing.T) {
	tcs := []struct {
		name     string
		fn       interface{}
		args     []cty.Value
		expected cty.Value
	}{
		{
			name: "const",
			fn: func() (string, error) {
				return "hello", nil
			},
			args:     nil,
			expected: cty.StringVal("hello"),
		},
		{
			name: "returnsStructure",
			fn: func() (structure, error) {
				return structure{Field: "hello"}, nil
			},
			args:     nil,
			expected: cty.ObjectVal(map[string]cty.Value{"field": cty.StringVal("hello")}),
		},
		{
			name: "acceptsStructure",
			fn: func(s structure) (string, error) {
				return s.Field, nil
			},
			args:     []cty.Value{cty.ObjectVal(map[string]cty.Value{"field": cty.StringVal("hello")})},
			expected: cty.StringVal("hello"),
		},
		{
			name: "returnsStructureWithPointer",
			fn: func() (structureWithPointer, error) {
				value := "hello"
				return structureWithPointer{Field: &value}, nil
			},
			args:     nil,
			expected: cty.ObjectVal(map[string]cty.Value{"field": cty.StringVal("hello")}),
		},
		{
			name: "acceptsStructureWithPointer",
			fn: func(s structureWithPointer) (*string, error) {
				return s.Field, nil
			},
			args:     []cty.Value{cty.ObjectVal(map[string]cty.Value{"field": cty.StringVal("hello")})},
			expected: cty.StringVal("hello"),
		},
		{
			name: "acceptsStructureWithPointerNil",
			fn: func(s structureWithPointer) (*string, error) {
				return s.Field, nil
			},
			args:     []cty.Value{cty.ObjectVal(map[string]cty.Value{"field": cty.NullVal(cty.String)})},
			expected: cty.NullVal(cty.String),
		},
		{
			name: "variadicSingle",
			fn: func(s string, ss ...string) ([]string, error) {
				return append(ss, s), nil
			},
			args:     []cty.Value{cty.StringVal("hello"), cty.StringVal("world")},
			expected: cty.ListVal([]cty.Value{cty.StringVal("world"), cty.StringVal("hello")}),
		},
		{
			name: "variadicEmpty",
			fn: func(s string, ss ...string) ([]string, error) {
				return append(ss, s), nil
			},
			args:     []cty.Value{cty.StringVal("hello")},
			expected: cty.ListVal([]cty.Value{cty.StringVal("hello")}),
		},
		{
			name: "variadicMulti",
			fn: func(s string, ss ...string) ([]string, error) {
				return append(ss, s), nil
			},
			args:     []cty.Value{cty.StringVal("hello"), cty.StringVal("world"), cty.StringVal("foo")},
			expected: cty.ListVal([]cty.Value{cty.StringVal("world"), cty.StringVal("foo"), cty.StringVal("hello")}),
		},
		{
			name: "int",
			fn: func(i int) (int, error) {
				return i, nil
			},
			args:     []cty.Value{cty.NumberIntVal(42)},
			expected: cty.NumberIntVal(42),
		},
		{
			name: "int8",
			fn: func(i int8) (int8, error) {
				return i, nil
			},
			args:     []cty.Value{cty.NumberIntVal(42)},
			expected: cty.NumberIntVal(42),
		},
		{
			name: "int16",
			fn: func(i int16) (int16, error) {
				return i, nil
			},
			args:     []cty.Value{cty.NumberIntVal(42)},
			expected: cty.NumberIntVal(42),
		},
		{
			name: "int32",
			fn: func(i int32) (int32, error) {
				return i, nil
			},
			args:     []cty.Value{cty.NumberIntVal(42)},
			expected: cty.NumberIntVal(42),
		},
		{
			name: "int64",
			fn: func(i int64) (int64, error) {
				return i, nil
			},
			args:     []cty.Value{cty.NumberIntVal(42)},
			expected: cty.NumberIntVal(42),
		},
		{
			name: "uint",
			fn: func(i uint) (uint, error) {
				return i, nil
			},
			args:     []cty.Value{cty.NumberIntVal(42)},
			expected: cty.NumberIntVal(42),
		},
		{
			name: "uint8",
			fn: func(i uint8) (uint8, error) {
				return i, nil
			},
			args:     []cty.Value{cty.NumberIntVal(42)},
			expected: cty.NumberIntVal(42),
		},
		{
			name: "uint16",
			fn: func(i uint16) (uint16, error) {
				return i, nil
			},
			args:     []cty.Value{cty.NumberIntVal(42)},
			expected: cty.NumberIntVal(42),
		},
		{
			name: "uint32",
			fn: func(i uint32) (uint32, error) {
				return i, nil
			},
			args:     []cty.Value{cty.NumberIntVal(42)},
			expected: cty.NumberIntVal(42),
		},
		{
			name: "uint64",
			fn: func(i uint64) (uint64, error) {
				return i, nil
			},
			args:     []cty.Value{cty.NumberIntVal(42)},
			expected: cty.NumberIntVal(42),
		},
		{
			name: "float32",
			fn: func(i float32) (float32, error) {
				return i, nil
			},
			args:     []cty.Value{cty.NumberFloatVal(42.0)},
			expected: cty.NumberFloatVal(42.0),
		},
		{
			name: "float64",
			fn: func(i float64) (float64, error) {
				return i, nil
			},
			args:     []cty.Value{cty.NumberFloatVal(42.0)},
			expected: cty.NumberFloatVal(42.0),
		},
		{
			name: "string",
			fn: func(s string) (string, error) {
				return s, nil
			},
			args:     []cty.Value{cty.StringVal("hello")},
			expected: cty.StringVal("hello"),
		},
		{
			name: "empty string",
			fn: func(s string) (string, error) {
				return s, nil
			},
			args:     []cty.Value{cty.StringVal("")},
			expected: cty.StringVal(""),
		},
		{
			name: "bool",
			fn: func(b bool) (bool, error) {
				return b, nil
			},
			args:     []cty.Value{cty.BoolVal(true)},
			expected: cty.BoolVal(true),
		},
		{
			name: "bool - false",
			fn: func(b bool) (bool, error) {
				return b, nil
			},
			args:     []cty.Value{cty.BoolVal(false)},
			expected: cty.BoolVal(false),
		},
		{
			name: "slice",
			fn: func(s []string) ([]string, error) {
				return s, nil
			},
			args:     []cty.Value{cty.ListVal([]cty.Value{cty.StringVal("hello"), cty.StringVal("world")})},
			expected: cty.ListVal([]cty.Value{cty.StringVal("hello"), cty.StringVal("world")}),
		},
		{
			name: "nullSlice",
			fn: func(s []string) ([]string, error) {
				return s, nil
			},
			args:     []cty.Value{cty.NullVal(cty.List(cty.String))},
			expected: cty.NullVal(cty.List(cty.String)),
		},
		{
			name: "emptySlice",
			fn: func(s []string) ([]string, error) {
				return s, nil
			},
			args:     []cty.Value{cty.ListValEmpty(cty.String)},
			expected: cty.ListValEmpty(cty.String),
		},
		{
			name: "map",
			fn: func(m map[string]string) (map[string]string, error) {
				return m, nil
			},
			args:     []cty.Value{cty.MapVal(map[string]cty.Value{"hello": cty.StringVal("world")})},
			expected: cty.MapVal(map[string]cty.Value{"hello": cty.StringVal("world")}),
		},
		{
			name: "nullMap",
			fn: func(m map[string]string) (map[string]string, error) {
				return m, nil
			},
			args:     []cty.Value{cty.NullVal(cty.Map(cty.String))},
			expected: cty.NullVal(cty.Map(cty.String)),
		},
		{
			name: "emptyMap",
			fn: func(m map[string]string) (map[string]string, error) {
				return m, nil
			},
			args:     []cty.Value{cty.MapValEmpty(cty.String)},
			expected: cty.MapValEmpty(cty.String),
		},
		{
			name: "null",
			fn: func(value *string) (*string, error) {
				return value, nil
			},
			args:     []cty.Value{cty.NullVal(cty.String)},
			expected: cty.NullVal(cty.String),
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			RegisterFunction(tc.name, tc.fn)

			returned, err := functions[tc.name].Call(tc.args)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if diff := cmp.Diff(tc.expected, returned, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}
