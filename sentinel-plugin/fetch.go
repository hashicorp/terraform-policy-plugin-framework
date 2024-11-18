// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package sentinel_plugin

import (
	"context"

	"github.com/hashicorp/go-s2/sentinel/evaluate"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/go-s2-plugin/sentinel-plugin/proto"
	"github.com/hashicorp/go-s2-plugin/sentinel-plugin/proto/diagnostics"
	"github.com/hashicorp/go-s2-plugin/sentinel-plugin/proto/values"
)

var _ proto.FetchServer = (*fetchServer)(nil)

// fetchServer is the internal implementation of the FetchService server.
//
// Users don't need to use this directly, but it is used by the plugin to
// implement the FetchService interface.
type fetchServer struct {
	impl evaluate.Fetch
}

func (s *fetchServer) Fetch(ctx context.Context, req *proto.FetchRequest) (*proto.FetchResponse, error) {
	value, diags := s.impl(ctx, req.Type, req.Name, req.Requests.ToCtyValue(cty.DynamicPseudoType))
	return &proto.FetchResponse{
		Value:       values.FromCtyValue(value, cty.DynamicPseudoType),
		Diagnostics: diagnostics.FromHclDiagnostics(diags),
	}, nil
}
