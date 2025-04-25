// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package plugins

import (
	context "context"
	"fmt"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/hashicorp/terraform-policy-plugin-framework/policy-plugin/proto"
)

var (
	Handshake = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "TF_POLICY_PLUGIN",
		MagicCookieValue: "95ADAEF3D8C4",
	}

	_ plugin.GRPCPlugin  = (*PluginServer)(nil)
	_ proto.PluginServer = (*GrpcServer)(nil)
)

type PluginServer struct {
	plugin.NetRPCUnsupportedPlugin
}

func (p *PluginServer) GRPCServer(_ *plugin.GRPCBroker, server *grpc.Server) error {
	proto.RegisterPluginServer(server, new(GrpcServer))
	return nil
}

func (p *PluginServer) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
