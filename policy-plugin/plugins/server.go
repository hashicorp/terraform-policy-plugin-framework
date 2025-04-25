// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package plugins

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-plugin"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/msgpack"

	"github.com/hashicorp/terraform-policy-plugin-framework/policy-plugin/proto"
)

func Serve() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]plugin.Plugin{
			"plugin": new(PluginServer),
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

type GrpcServer struct{}

func (g *GrpcServer) Setup(context.Context, *proto.PluginSetupRequest) (*proto.PluginSetupResponse, error) {
	// Nothing to do at the moment.
	return new(proto.PluginSetupResponse), nil
}

func (g *GrpcServer) ListFunctions(context.Context, *proto.ListFunctionsRequest) (*proto.ListFunctionsResponse, error) {
	fns := make(map[string]*proto.Function, len(functions))
	for name, function := range functions {
		fn, err := proto.FromCtyFunction(function)
		if err != nil {
			return nil, err
		}

		fns[name] = fn
	}
	return &proto.ListFunctionsResponse{
		Functions: fns,
	}, nil
}

func (g *GrpcServer) ExecuteFunction(_ context.Context, request *proto.ExecuteFunctionRequest) (*proto.ExecuteFunctionResponse, error) {
	function, ok := functions[request.Name]
	if !ok {
		return nil, fmt.Errorf("function %q not found", request.Name)
	}

	parameters := function.Params()
	variadicParameter := function.VarParam()

	args := make([]cty.Value, len(request.Arguments))
	for i, argument := range request.Arguments {
		if i >= len(parameters) {
			if variadicParameter == nil {
				return nil, errors.New("too many arguments")
			}

			arg, err := msgpack.Unmarshal(argument, variadicParameter.Type)
			if err != nil {
				return nil, err
			}

			args[i] = arg
			continue
		}

		arg, err := msgpack.Unmarshal(argument, parameters[i].Type)
		if err != nil {
			return nil, err
		}
		args[i] = arg
	}

	ret, err := function.Call(args)
	if err != nil {
		return nil, err
	}

	returnType, err := function.ReturnTypeForValues(args)
	if err != nil {
		return nil, err
	}

	result, err := msgpack.Marshal(ret, returnType)
	if err != nil {
		return nil, err
	}

	return &proto.ExecuteFunctionResponse{
		Result: result,
	}, nil
}
