// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package proto

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

func (parameter *FunctionParameter) ToCtyParameter() (function.Parameter, error) {
	returnType, err := ctyjson.UnmarshalType(parameter.Type)
	if err != nil {
		return function.Parameter{}, err
	}

	return function.Parameter{
		Name:             parameter.Name,
		Description:      parameter.Description,
		Type:             returnType,
		AllowNull:        parameter.AllowNull,
		AllowUnknown:     parameter.AllowUnknown,
		AllowDynamicType: parameter.AllowDynamic,
		AllowMarked:      parameter.AllowMarked,
	}, nil
}

func FromCtyParameter(parameter function.Parameter) (*FunctionParameter, error) {
	returnType, err := ctyjson.MarshalType(parameter.Type)
	if err != nil {
		return nil, err
	}

	return &FunctionParameter{
		Name:         parameter.Name,
		Description:  parameter.Description,
		Type:         returnType,
		AllowNull:    parameter.AllowNull,
		AllowUnknown: parameter.AllowUnknown,
		AllowDynamic: parameter.AllowDynamicType,
		AllowMarked:  parameter.AllowMarked,
	}, nil
}

func FromCtyFunction(fn function.Function) (*Function, error) {
	var types []cty.Type

	var parameters []*FunctionParameter
	for _, param := range fn.Params() {
		types = append(types, param.Type)
		parameterType, err := FromCtyParameter(param)
		if err != nil {
			return nil, err
		}
		parameters = append(parameters, parameterType)
	}

	var variadic *FunctionParameter
	if v := fn.VarParam(); v != nil {
		types = append(types, v.Type)
		variadic, _ = FromCtyParameter(*v)
	}

	returns, err := fn.ReturnType(types)
	if err != nil {
		return nil, err
	}

	returnType, err := ctyjson.MarshalType(returns)
	if err != nil {
		return nil, err
	}

	return &Function{
		Parameters:        parameters,
		VariadicParameter: variadic,
		ReturnType:        returnType,
		Description:       fn.Description(),
	}, nil
}
