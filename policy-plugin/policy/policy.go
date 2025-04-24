// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package policy

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform-policy-core/policy"
	"github.com/hashicorp/terraform-policy-core/policy/types"
	"github.com/zclconf/go-cty/cty"
	"google.golang.org/grpc"

	"github.com/hashicorp/terraform-policy-plugin-framework/policy-plugin/proto"
	proto_cty "github.com/hashicorp/terraform-policy-plugin-framework/policy-plugin/proto/cty"
	"github.com/hashicorp/terraform-policy-plugin-framework/policy-plugin/proto/diagnostics"
)

var (
	_ Policy = (*policyClient)(nil)
)

// Policy represents a plugin clients connection to the Terraform Policy server.
//
// It implements the same policy.Engine interface that users of the main
// library use so users can use this as a drop-in replacement for the main
// library.
//
// However, the Terraform Policy plugin must first be configured with the Setup
// function and must be closed with the Close function when done.
type Policy interface {
	policy.Engine

	// Setup configures the Terraform Policy plugin with the given setup
	// request. This must be called before any policies are evaluated.
	Setup(ctx context.Context, request *proto.PolicySetupRequest) (*proto.PolicySetupResponse_ServerCapabilities, hcl.Diagnostics)

	// Close closes the connection to the Terraform Policy plugin.
	Close()
}

// Connect launches a new Terraform Policy plugin process and connects to it.
// The returned Terraform Policy instance can be used to evaluate policies.
func Connect(ctx context.Context, pgm string, args ...string) (Policy, error) {
	cmd := exec.CommandContext(ctx, pgm, args...)

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]plugin.Plugin{
			"policy": &policyPlugin{},
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

	raw, err := rpc.Dispense("policy")
	if err != nil {
		return nil, fmt.Errorf("failed to dispense plugin: %v", err)
	}

	sc := raw.(*policyClient)
	sc.plugin = client
	return sc, nil
}

// policyClient is a Terraform Policy client that connects to a Terraform Policy
// plugin. This is the main implementation of the Terraform Policy interface,
// and simply forwards requests to the plugin.
type policyClient struct {
	plugin *plugin.Client

	broker *plugin.GRPCBroker
	client proto.PolicyClient
}

func (s *policyClient) Setup(ctx context.Context, request *proto.PolicySetupRequest) (*proto.PolicySetupResponse_ServerCapabilities, hcl.Diagnostics) {
	response, err := s.client.Setup(ctx, request)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to setup Terraform Policy",
				Detail:   err.Error(),
			},
		}
	}
	return response.ServerCapabilities, diagnostics.ToHclDiagnostics(response.Diagnostics)
}

func (s *policyClient) EvaluatePoliciesFor(ctx context.Context, consumer string, requestedType string, attrs, metadata cty.Value, opts *policy.EvaluateOpts) (types.EvaluateResult, hcl.Diagnostics) {
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
				impl:      opts.Fetch,
				functions: opts.Functions,
			})
			return server
		})

		defer func() {
			server.Stop()
		}()
	}

	resp, err := s.client.Evaluate(ctx, &proto.PolicyEvaluateRequest{
		FetchService: fetch,
		Consumer:     consumer,
		Resource:     requestedType,
		Attrs:        proto_cty.FromCtyValue(attrs, cty.DynamicPseudoType),
		Metadata:     proto_cty.FromCtyValue(metadata, cty.DynamicPseudoType),
		Functions: func() []*proto_cty.Function {
			var fns []*proto_cty.Function
			if opts != nil {
				for name, fn := range opts.Functions {
					fns = append(fns, proto_cty.FromCtyFunction(name, fn))
				}
			}
			return fns
		}(),
	})
	if err != nil {
		return types.EvaluateResultError, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to evaluate Terraform Policy files",
				Detail:   err.Error(),
			},
		}
	}
	return resp.Result.ToPolicyEvaluateResult(), diagnostics.ToHclDiagnostics(resp.Diagnostics)
}

func (s *policyClient) Close() {
	s.plugin.Kill()
}

// policyPlugin provides the client implementation of the Terraform Policy
// plugin.
type policyPlugin struct {
	plugin.NetRPCUnsupportedPlugin
}

func (s policyPlugin) GRPCServer(*plugin.GRPCBroker, *grpc.Server) error {
	// This package is only implementing the client side of the Terraform Policy
	// plugin.
	return fmt.Errorf("server configuration not supported")
}

func (s policyPlugin) GRPCClient(_ context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	return &policyClient{
		plugin: nil, // this will be set by the Connect function
		broker: broker,
		client: proto.NewPolicyClient(conn),
	}, nil
}
