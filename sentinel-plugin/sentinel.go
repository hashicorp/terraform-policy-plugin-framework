// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package sentinel_plugin

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/go-s2/sentinel"
	"github.com/hashicorp/go-s2/sentinel/types"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"google.golang.org/grpc"

	"github.com/hashicorp/go-s2-plugin/sentinel-plugin/proto"
	"github.com/hashicorp/go-s2-plugin/sentinel-plugin/proto/diagnostics"
	"github.com/hashicorp/go-s2-plugin/sentinel-plugin/proto/values"
)

var (
	_ Sentinel = (*sentinelClient)(nil)
)

// Sentinel represents a plugin clients connection to the Sentinel server.
//
// It implements the same sentinel.Engine interface that users of the main
// library use so users can use this as a drop-in replacement for the main
// library.
//
// However, the Sentinel plugin must first be configured with the Setup function
// and must be closed with the Close function when done.
type Sentinel interface {
	sentinel.Engine

	// Setup configures the Sentinel plugin with the given setup request. This
	// must be called before any policies are evaluated.
	Setup(ctx context.Context, request *proto.SetupRequest) (*proto.ServerCapabilities, hcl.Diagnostics)

	// Close closes the connection to the Sentinel plugin.
	Close()
}

// Connect launches a new Sentinel plugin process and connects to it. The
// returned Sentinel instance can be used to evaluate policies.
func Connect(ctx context.Context, pgm string, args ...string) (Sentinel, error) {
	cmd := exec.CommandContext(ctx, pgm, args...)

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]plugin.Plugin{
			"sentinel": &sentinelPlugin{},
		},
		Cmd: cmd,
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolGRPC,
		},
		Logger: NewLogger(),
	})

	rpc, err := client.Client()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to plugin: %v", err)
	}

	raw, err := rpc.Dispense("sentinel")
	if err != nil {
		return nil, fmt.Errorf("failed to dispense plugin: %v", err)
	}

	sc := raw.(*sentinelClient)
	sc.plugin = client
	return sc, nil
}

// sentinelClient is a Sentinel client that connects to a Sentinel plugin. This
// is the main implementation of the Sentinel interface, and simply forwards
// requests to the plugin.
type sentinelClient struct {
	plugin *plugin.Client

	broker *plugin.GRPCBroker
	client proto.SentinelClient
}

func (s *sentinelClient) Setup(ctx context.Context, request *proto.SetupRequest) (*proto.ServerCapabilities, hcl.Diagnostics) {
	response, err := s.client.Setup(ctx, request)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to setup Sentinel",
				Detail:   err.Error(),
			},
		}
	}
	return response.ServerCapabilities, diagnostics.ToHclDiagnostics(response.Diagnostics)
}

func (s *sentinelClient) EvaluatePoliciesFor(ctx context.Context, requestedType string, attrs, metadata cty.Value, opts *sentinel.EvaluateOpts) (types.EvaluateResult, hcl.Diagnostics) {
	var fetch uint32

	if opts != nil {
		var server *grpc.Server
		if opts.EvaluateUnknownFilters {
			return types.EvaluateResultError, hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "Evaluating unknown filters is not supported",
					Detail:   "opts.EvaluateUnknownFilters is set to true, but this is not supported by the current plugin architecture.",
				},
			}
		}

		fetch = s.broker.NextId()
		go s.broker.AcceptAndServe(fetch, func(grpcOpts []grpc.ServerOption) *grpc.Server {
			server = grpc.NewServer(grpcOpts...)
			proto.RegisterFetchServer(server, &fetchServer{
				impl: opts.Fetch,
			})
			return server
		})

		defer func() {
			server.Stop()
		}()
	}

	resp, err := s.client.Evaluate(ctx, &proto.EvaluateRequest{
		FetchService: fetch,
		Resource:     requestedType,
		Attrs:        values.FromCtyValue(attrs, cty.DynamicPseudoType),
		Metadata:     values.FromCtyValue(metadata, cty.DynamicPseudoType),
	})
	if err != nil {
		return types.EvaluateResultError, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to evaluate Sentinel policies",
				Detail:   err.Error(),
			},
		}
	}
	return resp.Result.ToSentinelEvaluateResult(), diagnostics.ToHclDiagnostics(resp.Diagnostics)
}

func (s *sentinelClient) Close() {
	s.plugin.Kill()
}

// sentinelPlugin provides the client implementation of the Sentinel plugin.
type sentinelPlugin struct {
	plugin.NetRPCUnsupportedPlugin
}

func (s sentinelPlugin) GRPCServer(*plugin.GRPCBroker, *grpc.Server) error {
	// This package is only implementing the client side of the Sentinel plugin.
	return fmt.Errorf("server configuration not supported")
}

func (s sentinelPlugin) GRPCClient(_ context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	return &sentinelClient{
		plugin: nil, // this will be set by the Connect function
		broker: broker,
		client: proto.NewSentinelClient(conn),
	}, nil
}
