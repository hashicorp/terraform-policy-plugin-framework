// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package policy

import "github.com/hashicorp/go-plugin"

// Handshake is the handshake configuration for the Terraform Policy plugin.
// This is shared by the client and server.
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "TF_POLICY_PLUGIN",
	MagicCookieValue: "6F11ED78A2AB",
}
