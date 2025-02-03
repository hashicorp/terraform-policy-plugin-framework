// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package plugins

import (
	"fmt"
	"reflect"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"

	"github.com/hashicorp/go-s2-plugin/sentinel-plugin/plugins/convert"
)

var (
	functions map[string]function.Function
)

func init() {
	functions = make(map[string]function.Function)
}

// RegisterFunctionDirect registers a cty function with the given name.
func RegisterFunctionDirect(name string, fn function.Function) {
	if _, ok := functions[name]; ok {
		panic("function already registered")
	}
	functions[name] = fn
}

// CallFunction calls the function with the given name and arguments. This is
// mainly used for testing.
func CallFunction(name string, args ...cty.Value) (cty.Value, error) {
	fn, ok := functions[name]
	if !ok {
		return cty.NilVal, fmt.Errorf("function %s not found", name)
	}
	return fn.Call(args)
}

// RegisterFunction registers a Go function with the given name.
func RegisterFunction(name string, fn interface{}) {
	value := reflect.ValueOf(fn)
	if value.Kind() != reflect.Func {
		panic("fn must be a function")
	}

	if value.Type().NumOut() != 2 {
		panic("function must return two values")
	}
	if value.Type().Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		panic("second return value must be an error")
	}

	var args []function.Parameter
	var variadic *function.Parameter
	for ix := 0; ix < value.Type().NumIn(); ix++ {

		if value.Type().IsVariadic() && ix == value.Type().NumIn()-1 {
			in := value.Type().In(ix)
			param, err := convert.ToCtyType(in.Elem())
			if err != nil {
				panic(fmt.Errorf("invalid parameter %d for %s: %v", ix, name, err))
			}

			variadic = &function.Parameter{
				Type:      param,
				AllowNull: in.Kind() == reflect.Pointer || param.IsCollectionType(),
			}
			continue
		}

		in := value.Type().In(ix)
		param, err := convert.ToCtyType(in)
		if err != nil {
			panic(fmt.Errorf("invalid parameter %d for %s: %v", ix, name, err))
		}

		args = append(args, function.Parameter{
			Type:      param,
			AllowNull: in.Kind() == reflect.Pointer || param.IsCollectionType(),
		})
	}

	returnType, err := convert.ToCtyType(value.Type().Out(0))
	if err != nil {
		panic(fmt.Errorf("invalid return type: %v", err))
	}

	RegisterFunctionDirect(name, function.New(&function.Spec{
		Params:   args,
		VarParam: variadic,
		Type:     function.StaticReturnType(returnType),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {

			var arguments []reflect.Value
			for i, arg := range args {
				if value.Type().IsVariadic() && i >= value.Type().NumIn()-1 {
					want := value.Type().In(value.Type().NumIn() - 1)
					argument, err := convert.FromCtyValue(arg, want.Elem())
					if err != nil {
						return cty.NullVal(returnType), fmt.Errorf("failed to convert variadic argument %d: %w", i, err)
					}
					arguments = append(arguments, argument)
					continue
				}

				want := value.Type().In(i)
				argument, err := convert.FromCtyValue(arg, want)
				if err != nil {
					return cty.NullVal(returnType), fmt.Errorf("failed to convert argument %d: %w", i, err)
				}
				arguments = append(arguments, argument)
			}

			results := value.Call(arguments)
			if err := results[1].Interface(); err != nil {
				return cty.NilVal, err.(error)
			}

			value, err := convert.ToCtyValue(results[0], returnType)
			if err != nil {
				return cty.NilVal, fmt.Errorf("failed to convert result: %w", err)
			}
			return value, nil
		},
	}))
}
