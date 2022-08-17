// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package scheme

import (
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"
)

var (
	handshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "VERAISON_PLUGIN",
		MagicCookieValue: "VERAISON",
	}

	// pluginMap is the map of plugins we can dispense.
	pluginMap = map[string]plugin.Plugin{
		"scheme": &Plugin{},
	}
)

type SchemeGoPlugin struct {
	Path                string
	Name                string
	SupportedMediaTypes []string
	Handle              IScheme
	Client              *plugin.Client
}

func NewSchemeGoPlugin(path string) (*SchemeGoPlugin, error) {
	client := plugin.NewClient(
		&plugin.ClientConfig{
			HandshakeConfig: handshakeConfig,
			Plugins:         pluginMap,
			Cmd:             exec.Command(path),
		},
	)

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf(
			"unable to create the RPC client for %s: %w",
			path, err,
		)
	}

	protocolClient, err := rpcClient.Dispense("scheme")
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf(
			"unable to create a new instance of plugin %s: %w",
			path, err,
		)
	}

	handle, ok := protocolClient.(IScheme)
	if !ok {
		client.Kill()

		return nil, fmt.Errorf(
			"plugin %s does not provide an implementation of the IScheme interface",
			path,
		)
	}

	return &SchemeGoPlugin{
		Path:                path,
		Name:                handle.GetName(),
		SupportedMediaTypes: handle.GetSupportedMediaTypes(),
		Handle:              handle,
		Client:              client,
	}, nil
}
