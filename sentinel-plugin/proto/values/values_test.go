// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package values

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestValuesRoundtrip(t *testing.T) {
	tcs := map[string]struct {
		ctyValue   cty.Value
		protoValue *Value
	}{
		"totally unknown": {
			ctyValue: cty.DynamicVal,
			protoValue: &Value{
				Type: &Type{
					Identifier: Type_UNKNOWN,
				},
				Unknown: true,
			},
		},
		"unknown": {
			ctyValue: cty.UnknownVal(cty.String),
			protoValue: &Value{
				Unknown: true,
			},
		},
		"null": {
			ctyValue:   cty.NullVal(cty.String),
			protoValue: &Value{},
		},
		"sensitive": {
			ctyValue: cty.StringVal("hello").Mark(Sensitive),
			protoValue: &Value{
				Value: &Value_StringValue{
					StringValue: "hello",
				},
				Sensitive: true,
			},
		},
		"sensitive null": {
			ctyValue: cty.NullVal(cty.String).Mark(Sensitive),
			protoValue: &Value{
				Sensitive: true,
			},
		},
		"sensitive unknown": {
			ctyValue: cty.UnknownVal(cty.String).Mark(Sensitive),
			protoValue: &Value{
				Sensitive: true,
				Unknown:   true,
			},
		},
		"string": {
			ctyValue: cty.StringVal("hello"),
			protoValue: &Value{
				Value: &Value_StringValue{
					StringValue: "hello",
				},
			},
		},
		"boolean": {
			ctyValue: cty.True,
			protoValue: &Value{
				Value: &Value_BooleanValue{
					BooleanValue: true,
				},
			},
		},
		"number": {
			ctyValue: cty.NumberIntVal(42),
			protoValue: &Value{
				Value: &Value_StringValue{
					StringValue: "42",
				},
			},
		},
		"list": {
			ctyValue: cty.ListVal([]cty.Value{cty.StringVal("hello"), cty.StringVal("world")}),
			protoValue: &Value{
				Value: &Value_ListValue{
					ListValue: &ListValue{
						Values: []*Value{
							{
								Value: &Value_StringValue{
									StringValue: "hello",
								},
							},
							{
								Value: &Value_StringValue{
									StringValue: "world",
								},
							},
						},
					},
				},
			},
		},
		"set": {
			ctyValue: cty.SetVal([]cty.Value{cty.StringVal("hello"), cty.StringVal("world")}),
			protoValue: &Value{
				Value: &Value_ListValue{
					ListValue: &ListValue{
						Values: []*Value{
							{
								Value: &Value_StringValue{
									StringValue: "hello",
								},
							},
							{
								Value: &Value_StringValue{
									StringValue: "world",
								},
							},
						},
					},
				},
			},
		},
		"map": {
			ctyValue: cty.MapVal(map[string]cty.Value{
				"hello": cty.StringVal("world"),
			}),
			protoValue: &Value{
				Value: &Value_MapValue{
					MapValue: &MapValue{
						Values: map[string]*Value{
							"hello": {
								Value: &Value_StringValue{
									StringValue: "world",
								},
							},
						},
					},
				},
			},
		},
		"object": {
			ctyValue: cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("world"),
			}),
			protoValue: &Value{
				Value: &Value_MapValue{
					MapValue: &MapValue{
						Values: map[string]*Value{
							"hello": {
								Value: &Value_StringValue{
									StringValue: "world",
								},
							},
						},
					},
				},
			},
		},
		"tuple": {
			ctyValue: cty.TupleVal([]cty.Value{cty.StringVal("hello"), cty.StringVal("world")}),
			protoValue: &Value{
				Value: &Value_ListValue{
					ListValue: &ListValue{
						Values: []*Value{
							{
								Value: &Value_StringValue{
									StringValue: "hello",
								},
							},
							{
								Value: &Value_StringValue{
									StringValue: "world",
								},
							},
						},
					},
				},
			},
		},
		"nested null": {
			ctyValue: cty.ObjectVal(map[string]cty.Value{
				"hello": cty.NullVal(cty.String),
			}),
			protoValue: &Value{
				Value: &Value_MapValue{
					MapValue: &MapValue{
						Values: map[string]*Value{
							"hello": {},
						},
					},
				},
			},
		},
		"nested unknown": {
			ctyValue: cty.ObjectVal(map[string]cty.Value{
				"hello": cty.UnknownVal(cty.String),
			}),
			protoValue: &Value{
				Value: &Value_MapValue{
					MapValue: &MapValue{
						Values: map[string]*Value{
							"hello": {
								Unknown: true,
							},
						},
					},
				},
			},
		},
		"nested sensitive": {
			ctyValue: cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("world").Mark(Sensitive),
			}),
			protoValue: &Value{
				Value: &Value_MapValue{
					MapValue: &MapValue{
						Values: map[string]*Value{
							"hello": {
								Value: &Value_StringValue{
									StringValue: "world",
								},
								Sensitive: true,
							},
						},
					},
				},
			},
		},
	}
	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			t.Run("FromCtyValue", func(t *testing.T) {
				if diff := cmp.Diff(FromCtyValue(tc.ctyValue, tc.ctyValue.Type()), tc.protoValue, protocmp.Transform()); len(diff) > 0 {
					t.Error(diff)
				}
			})
			t.Run("ToCtyValue", func(t *testing.T) {
				if diff := cmp.Diff(tc.protoValue.ToCtyValue(tc.ctyValue.Type()), tc.ctyValue, ctydebug.CmpOptions); len(diff) > 0 {
					t.Error(diff)
				}
			})
			t.Run("FromCtyValue (dynamic)", func(t *testing.T) {
				protoValue := proto.Clone(tc.protoValue).(*Value)
				protoValue.Type = FromCtyType(tc.ctyValue.Type())
				if diff := cmp.Diff(FromCtyValue(tc.ctyValue, cty.DynamicPseudoType), protoValue, protocmp.Transform()); len(diff) > 0 {
					t.Error(diff)
				}
			})
			t.Run("ToCtyValue (dynamic)", func(t *testing.T) {
				protoValue := proto.Clone(tc.protoValue).(*Value)
				protoValue.Type = FromCtyType(tc.ctyValue.Type())
				if diff := cmp.Diff(protoValue.ToCtyValue(cty.DynamicPseudoType), tc.ctyValue, ctydebug.CmpOptions); len(diff) > 0 {
					t.Error(diff)
				}
			})

		})
	}
}
