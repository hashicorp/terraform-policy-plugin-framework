// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package plugins

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-plugin"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/go-s2-plugin/sentinel-plugin/proto"
	proto_cty "github.com/hashicorp/go-s2-plugin/sentinel-plugin/proto/cty"
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
	var fns []*proto_cty.Function
	for name, fn := range functions {
		fns = append(fns, proto_cty.FromCtyFunction(name, fn))
	}
	return &proto.ListFunctionsResponse{
		Functions: fns,
	}, nil
}

func (g *GrpcServer) ExecuteFunction(_ context.Context, request *proto.ExecuteFunctionRequest) (*proto.ExecuteFunctionResponse, error) {
	fn, ok := functions[request.Name]
	if !ok {
		return nil, fmt.Errorf("function %q not found", request.Name)
	}

	args := make([]cty.Value, len(request.Arguments))
	for i, arg := range request.Arguments {
		parameters := fn.Params()
		if i >= len(parameters) {
			if fn.VarParam() == nil {
				return nil, errors.New("too many arguments")
			}
			args[i] = arg.ToCtyValue(fn.VarParam().Type)
			continue
		}

		args[i] = arg.ToCtyValue(parameters[i].Type)
	}

	ret, err := fn.Call(args)
	if err != nil {
		return nil, err
	}

	returnType, err := fn.ReturnTypeForValues(args)
	if err != nil {
		return nil, err
	}

	return &proto.ExecuteFunctionResponse{
		Result: proto_cty.FromCtyValue(ret, returnType),
	}, nil
}
