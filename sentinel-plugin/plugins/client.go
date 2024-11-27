// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package plugins

import (
	"context"
	"fmt"
	"os/exec"

	go_plugin "github.com/hashicorp/go-plugin"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"

	"github.com/hashicorp/go-s2-plugin/sentinel-plugin/proto"
	proto_cty "github.com/hashicorp/go-s2-plugin/sentinel-plugin/proto/cty"
	"github.com/hashicorp/go-s2-plugin/sentinel-plugin/proto/diagnostics"
)

func Connect(ctx context.Context, plugin string, path string) (*PluginClient, error) {
	cmd := exec.CommandContext(ctx, path)

	client := go_plugin.NewClient(&go_plugin.ClientConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]go_plugin.Plugin{
			"plugin": &PluginServer{},
		},
		Cmd: cmd,
		AllowedProtocols: []go_plugin.Protocol{
			go_plugin.ProtocolGRPC,
		},
		Logger: NewLogger(plugin),
	})

	rpc, err := client.Client()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	raw, err := rpc.Dispense("plugin")
	if err != nil {
		return nil, fmt.Errorf("failed to dispense plugin: %w", err)
	}

	gc := raw.(*PluginClient)
	gc.plugin = client
	return gc, nil
}

type PluginClient struct {
	plugin *go_plugin.Client
	client proto.PluginClient
}

func (client *PluginClient) Setup(ctx context.Context) hcl.Diagnostics {
	response, err := client.client.Setup(ctx, new(proto.PluginSetupRequest))
	if err != nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to setup plugin",
				Detail:   err.Error(),
			},
		}
	}
	return diagnostics.ToHclDiagnostics(response.Diagnostics)
}

func (client *PluginClient) ListFunctions(ctx context.Context) (map[string]function.Function, hcl.Diagnostics) {
	response, err := client.client.ListFunctions(ctx, new(proto.ListFunctionsRequest))
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to list functions",
				Detail:   err.Error(),
			},
		}
	}

	fns := make(map[string]function.Function, len(response.Functions))
	for _, fn := range response.Functions {
		fns[fn.Name] = function.New(&function.Spec{
			Description: fn.Description,
			Params: func() []function.Parameter {
				var params []function.Parameter
				for _, param := range fn.Parameters {
					params = append(params, param.ToCtyParameter())
				}
				return params
			}(),
			VarParam: func() *function.Parameter {
				if fn.VariadicParameter != nil {
					param := fn.VariadicParameter.ToCtyParameter()
					return &param
				}
				return nil
			}(),
			Type: function.StaticReturnType(fn.ReturnType.ToCtyType()),
			Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
				value, err := client.ExecuteFunction(ctx, fn.Name, retType, args...)
				if err != nil {
					return cty.NullVal(retType), err
				}
				return value, nil
			},
		})
	}
	return fns, nil
}

func (client *PluginClient) ExecuteFunction(ctx context.Context, name string, ret cty.Type, args ...cty.Value) (cty.Value, error) {
	response, err := client.client.ExecuteFunction(ctx, &proto.ExecuteFunctionRequest{
		Name: name,
		Arguments: func() []*proto_cty.Value {
			var arguments []*proto_cty.Value
			for _, arg := range args {
				arguments = append(arguments, proto_cty.FromCtyValue(arg, arg.Type()))
			}
			return arguments
		}(),
	})
	if err != nil {
		return cty.NullVal(ret), err
	}
	return response.Result.ToCtyValue(ret), nil
}

func (client *PluginClient) Stop() {
	client.plugin.Kill()
}
