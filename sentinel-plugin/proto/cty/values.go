// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package cty

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
)

func FromCtyValue(v cty.Value, t cty.Type) *Value {
	if v == cty.NilVal {
		return nil
	}

	var protoType *Type
	if t == cty.DynamicPseudoType {
		t = v.Type()
		protoType = FromCtyType(t)
	}
	value := fromCtyValue(v, t)
	if value != nil {
		value.Type = protoType
	}
	return value
}

func fromCtyValue(v cty.Value, t cty.Type) *Value {
	if v == cty.NilVal {
		return nil
	}

	v, marks := v.Unmark()
	value := &Value{
		Marks: FromMarks(marks),
	}

	if !v.IsKnown() {
		value.Unknown = true
	} else if !v.IsNull() {
		switch identifier := TypeIdentifier(t); identifier {
		case Type_UNKNOWN:
			// do nothing
		case Type_BOOLEAN:
			value.Value = &Value_BooleanValue{BooleanValue: v.True()}
		case Type_NUMBER:
			value.Value = &Value_StringValue{StringValue: v.AsBigFloat().Text('f', -1)}
		case Type_STRING:
			value.Value = &Value_StringValue{StringValue: v.AsString()}
		case Type_LIST, Type_SET:
			elements := v.AsValueSlice()
			values := make([]*Value, len(elements))
			for i, elem := range elements {
				values[i] = fromCtyValue(elem, t.ElementType())
			}
			value.Value = &Value_ListValue{ListValue: &ListValue{Values: values}}
		case Type_MAP:
			elements := v.AsValueMap()
			values := make(map[string]*Value, len(elements))
			for key, elem := range elements {
				values[key] = fromCtyValue(elem, t.ElementType())
			}
			value.Value = &Value_MapValue{MapValue: &MapValue{Values: values}}
		case Type_OBJECT:
			attrs := make(map[string]*Value, len(v.AsValueMap()))
			for name, attr := range v.AsValueMap() {
				attrs[name] = fromCtyValue(attr, t.AttributeType(name))
			}
			value.Value = &Value_MapValue{MapValue: &MapValue{Values: attrs}}
		case Type_TUPLE:
			elems := v.AsValueSlice()
			values := make([]*Value, len(elems))
			for i, elem := range elems {
				values[i] = fromCtyValue(elem, t.TupleElementType(i))
			}
			value.Value = &Value_ListValue{ListValue: &ListValue{Values: values}}
		default:
			panic(fmt.Errorf("unsupported type %q", identifier))
		}
	}

	return value
}

func (v *Value) ToCtyValue(t cty.Type) cty.Value {
	if v == nil {
		return cty.NilVal
	}
	if t == cty.DynamicPseudoType {
		// Then the type should be encoded in the value.
		return v.toCtyValue(v.Type.ToCtyType())
	}
	return v.toCtyValue(t)
}

func (v *Value) toCtyValue(t cty.Type) (value cty.Value) {
	if v == nil {
		return cty.NilVal
	}

	if v.Unknown {
		value = cty.UnknownVal(t)
	} else if v.Value == nil {
		value = cty.NullVal(t)
	} else {
		switch identifier := TypeIdentifier(t); identifier {
		case Type_UNKNOWN:
			value = cty.DynamicVal
		case Type_BOOLEAN:
			value = cty.BoolVal(v.GetBooleanValue())
		case Type_NUMBER:
			value = cty.MustParseNumberVal(v.GetStringValue())
		case Type_STRING:
			value = cty.StringVal(v.GetStringValue())
		case Type_LIST:
			elements := make([]cty.Value, len(v.GetListValue().Values))
			for i, elem := range v.GetListValue().Values {
				elements[i] = elem.toCtyValue(t.ElementType())
			}
			if len(elements) == 0 {
				value = cty.ListValEmpty(t.ElementType())
			} else {
				value = cty.ListVal(elements)
			}
		case Type_SET:
			elements := make([]cty.Value, len(v.GetListValue().Values))
			for i, elem := range v.GetListValue().Values {
				elements[i] = elem.toCtyValue(t.ElementType())
			}
			if len(elements) == 0 {
				value = cty.SetValEmpty(t.ElementType())
			} else {
				value = cty.SetVal(elements)
			}
		case Type_MAP:
			items := make(map[string]cty.Value, len(v.GetMapValue().Values))
			for key, value := range v.GetMapValue().Values {
				items[key] = value.toCtyValue(t.ElementType())
			}
			if len(items) == 0 {
				value = cty.MapValEmpty(t.ElementType())
			} else {
				value = cty.MapVal(items)
			}
		case Type_OBJECT:
			attrs := make(map[string]cty.Value, len(v.GetMapValue().Values))
			for name, attr := range v.GetMapValue().Values {
				attrs[name] = attr.toCtyValue(t.AttributeType(name))
			}
			value = cty.ObjectVal(attrs)
		case Type_TUPLE:
			elems := make([]cty.Value, len(v.GetListValue().Values))
			for i, elem := range v.GetListValue().Values {
				elems[i] = elem.toCtyValue(t.TupleElementType(i))
			}
			value = cty.TupleVal(elems)
		default:
			panic(fmt.Errorf("unsupported type %q", identifier))
		}
	}

	marks := ToMarks(v.Marks)
	if len(marks) > 0 {
		value = value.WithMarks(marks)
	}
	return value
}
