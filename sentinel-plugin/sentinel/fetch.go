// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package sentinel

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-s2/sentinel/evaluate"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"

	"github.com/hashicorp/go-s2-plugin/sentinel-plugin/proto"
	proto_cty "github.com/hashicorp/go-s2-plugin/sentinel-plugin/proto/cty"
	"github.com/hashicorp/go-s2-plugin/sentinel-plugin/proto/diagnostics"
)

var _ proto.FetchServer = (*fetchServer)(nil)

// fetchServer is the internal implementation of the FetchService server.
//
// Users don't need to use this directly, but it is used by the plugin to
// implement the FetchService interface.
type fetchServer struct {
	impl      evaluate.Fetch
	functions map[string]function.Function
}

func (s *fetchServer) Function(_ context.Context, request *proto.FunctionRequest) (*proto.FunctionResponse, error) {
	if fn, ok := s.functions[request.Name]; ok {
		var args []cty.Value
		for ix, argument := range request.Arguments {
			var argumentType cty.Type
			if params := fn.Params(); ix < len(params) {
				argumentType = params[ix].Type
			} else {
				argumentType = fn.VarParam().Type
			}
			args = append(args, argument.ToCtyValue(argumentType))
		}

		rt, err := fn.ReturnTypeForValues(args)
		if err != nil {
			return nil, err
		}

		val, err := fn.Call(args)
		if err != nil {
			return nil, err
		}
		return &proto.FunctionResponse{
			Result: proto_cty.FromCtyValue(val, rt),
		}, nil
	}
	return nil, fmt.Errorf("function %q not found", request.Name)
}

func (s *fetchServer) Fetch(ctx context.Context, req *proto.FetchRequest) (*proto.FetchResponse, error) {
	value, diags := s.impl(ctx, req.Type, req.Name, req.Requests.ToCtyValue(cty.DynamicPseudoType))
	return &proto.FetchResponse{
		Value:       proto_cty.FromCtyValue(value, cty.DynamicPseudoType),
		Diagnostics: diagnostics.FromHclDiagnostics(diags, nil),
	}, nil
}
