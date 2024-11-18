// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package values

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
)

func TypeIdentifier(t cty.Type) Type_Identifier {
	switch {
	case t == cty.DynamicPseudoType, t == cty.NilType:
		return Type_UNKNOWN
	case t.IsPrimitiveType():
		switch t {
		case cty.Bool:
			return Type_BOOLEAN
		case cty.Number:
			return Type_NUMBER
		case cty.String:
			return Type_STRING
		default:
			panic(fmt.Errorf("unsupported primitive type %q", t.FriendlyName()))
		}
	case t.IsCollectionType():
		switch {
		case t.IsListType():
			return Type_LIST
		case t.IsSetType():
			return Type_SET
		case t.IsMapType():
			return Type_MAP
		default:
			panic(fmt.Errorf("unsupported collection type %q", t.FriendlyName()))
		}
	case t.IsObjectType():
		return Type_OBJECT
	case t.IsTupleType():
		return Type_TUPLE
	default:
		panic(fmt.Errorf("unsupported type %q", t.FriendlyName()))
	}
}

func FromCtyType(t cty.Type) *Type {
	if t == cty.NilType {
		return nil
	}

	identifier := TypeIdentifier(t)
	switch identifier {
	case Type_UNKNOWN, Type_BOOLEAN, Type_NUMBER, Type_STRING:
		return &Type{Identifier: identifier}
	case Type_LIST, Type_SET, Type_MAP:
		return &Type{
			Identifier: identifier,
			Type: &Type_ElementType{
				ElementType: FromCtyType(t.ElementType()),
			},
		}
	case Type_OBJECT:
		fields := make(map[string]*Type, len(t.AttributeTypes()))
		for name, attrType := range t.AttributeTypes() {
			fields[name] = FromCtyType(attrType)
		}
		return &Type{
			Identifier: identifier,
			Type: &Type_ObjectType{
				ObjectType: &ObjectType{
					Attributes: fields,
				},
			},
		}
	case Type_TUPLE:
		elems := make([]*Type, 0, len(t.TupleElementTypes()))
		for _, elemType := range t.TupleElementTypes() {
			elems = append(elems, FromCtyType(elemType))
		}
		return &Type{
			Identifier: identifier,
			Type: &Type_TupleType{
				TupleType: &TupleType{
					Elements: elems,
				},
			},
		}
	default:
		panic(fmt.Errorf("unsupported type %q", t.FriendlyName()))
	}
}

func (t *Type) ToCtyType() cty.Type {
	if t == nil {
		return cty.NilType
	}

	switch t.Identifier {
	case Type_UNKNOWN_IDENTIFIER, Type_UNKNOWN:
		return cty.DynamicPseudoType
	case Type_BOOLEAN:
		return cty.Bool
	case Type_NUMBER:
		return cty.Number
	case Type_STRING:
		return cty.String
	case Type_LIST:
		return cty.List(t.GetElementType().ToCtyType())
	case Type_SET:
		return cty.Set(t.GetElementType().ToCtyType())
	case Type_MAP:
		return cty.Map(t.GetElementType().ToCtyType())
	case Type_OBJECT:
		attrs := make(map[string]cty.Type, len(t.GetObjectType().Attributes))
		for name, attr := range t.GetObjectType().Attributes {
			attrs[name] = attr.ToCtyType()
		}
		return cty.Object(attrs)
	case Type_TUPLE:
		elems := make([]cty.Type, len(t.GetTupleType().Elements))
		for i, elem := range t.GetTupleType().Elements {
			elems[i] = elem.ToCtyType()
		}
		return cty.Tuple(elems)
	default:
		panic(fmt.Errorf("unsupported type %q", t.Identifier))
	}
}
