// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package cty

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

func (parameter *FunctionParameter) ToCtyParameter() function.Parameter {
	return function.Parameter{
		Name:             parameter.Name,
		Description:      parameter.Description,
		Type:             parameter.Type.ToCtyType(),
		AllowNull:        parameter.AllowNull,
		AllowUnknown:     parameter.AllowUnknown,
		AllowDynamicType: parameter.AllowDynamic,
		AllowMarked:      parameter.AllowMarked,
	}
}

func FromCtyParameter(parameter function.Parameter) *FunctionParameter {
	return &FunctionParameter{
		Name:         parameter.Name,
		Description:  parameter.Description,
		Type:         FromCtyType(parameter.Type),
		AllowNull:    parameter.AllowNull,
		AllowUnknown: parameter.AllowUnknown,
		AllowDynamic: parameter.AllowDynamicType,
		AllowMarked:  parameter.AllowMarked,
	}
}

func FromCtyFunction(name string, fn function.Function) *Function {
	var types []cty.Type

	var parameters []*FunctionParameter
	for _, param := range fn.Params() {
		types = append(types, param.Type)
		parameters = append(parameters, FromCtyParameter(param))
	}

	var variadic *FunctionParameter
	if v := fn.VarParam(); v != nil {
		types = append(types, v.Type)
		variadic = FromCtyParameter(*v)
	}

	returns, err := fn.ReturnType(types)
	if err != nil {
		return nil
	}

	return &Function{
		Name:              name,
		Parameters:        parameters,
		VariadicParameter: variadic,
		ReturnType:        FromCtyType(returns),
		Description:       fn.Description(),
	}
}
