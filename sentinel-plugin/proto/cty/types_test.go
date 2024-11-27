// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package cty

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestTypeRoundtrip(t *testing.T) {
	tcs := map[string]struct {
		ctyType   cty.Type
		protoType *Type
	}{
		"unknown": {
			ctyType: cty.DynamicPseudoType,
			protoType: &Type{
				Identifier: Type_UNKNOWN,
			},
		},
		"string": {
			ctyType: cty.String,
			protoType: &Type{
				Identifier: Type_STRING,
			},
		},
		"boolean": {
			ctyType: cty.Bool,
			protoType: &Type{
				Identifier: Type_BOOLEAN,
			},
		},
		"number": {
			ctyType: cty.Number,
			protoType: &Type{
				Identifier: Type_NUMBER,
			},
		},
		"list": {
			ctyType: cty.List(cty.String),
			protoType: &Type{
				Identifier: Type_LIST,
				Type: &Type_ElementType{
					ElementType: &Type{
						Identifier: Type_STRING,
					},
				},
			},
		},
		"set": {
			ctyType: cty.Set(cty.String),
			protoType: &Type{
				Identifier: Type_SET,
				Type: &Type_ElementType{
					ElementType: &Type{
						Identifier: Type_STRING,
					},
				},
			},
		},
		"map": {
			ctyType: cty.Map(cty.String),
			protoType: &Type{
				Identifier: Type_MAP,
				Type: &Type_ElementType{
					ElementType: &Type{
						Identifier: Type_STRING,
					},
				},
			},
		},
		"object": {
			ctyType: cty.Object(map[string]cty.Type{
				"foo": cty.String,
				"bar": cty.Number,
			}),
			protoType: &Type{
				Identifier: Type_OBJECT,
				Type: &Type_ObjectType{
					ObjectType: &ObjectType{
						Attributes: map[string]*Type{
							"foo": {
								Identifier: Type_STRING,
							},
							"bar": {
								Identifier: Type_NUMBER,
							},
						},
					},
				},
			},
		},
		"tuple": {
			ctyType: cty.Tuple([]cty.Type{cty.String, cty.Number}),
			protoType: &Type{
				Identifier: Type_TUPLE,
				Type: &Type_TupleType{
					TupleType: &TupleType{
						Elements: []*Type{
							{
								Identifier: Type_STRING,
							},
							{
								Identifier: Type_NUMBER,
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			t.Run("FromCtyType", func(t *testing.T) {
				if diff := cmp.Diff(FromCtyType(tc.ctyType), tc.protoType, protocmp.Transform()); len(diff) > 0 {
					t.Errorf("%v", diff)
				}
			})
			t.Run("ToCtyType", func(t *testing.T) {
				if diff := cmp.Diff(tc.protoType.ToCtyType(), tc.ctyType, ctydebug.CmpOptions); len(diff) > 0 {
					t.Errorf("%v", diff)
				}
			})
		})
	}

}
