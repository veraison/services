// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package goplugin

import (
	"fmt"
	"net/rpc"

	goplugin "github.com/hashicorp/go-plugin"
	"github.com/veraison/services/pkg/plugin"
)

var handshakeConfig = goplugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "VERAISON_PLUGIN",
	MagicCookieValue: "VERAISON",
}

type GoPlugin[I plugin.IPluggable] struct {
	Name string
	Impl I
}

func (p *GoPlugin[I]) Server(*goplugin.MuxBroker) (interface{}, error) {
	return GetRPCServer(p.Name, p.Impl), nil
}

func (p *GoPlugin[I]) Client(b *goplugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return GetRPCClient(p.Name, p.Impl, c), nil
}

var pluggableImplementationMap = map[string]goplugin.Plugin{}

func RegisterImplementation[I plugin.IPluggable](name string, i I) error {
	if other, ok := pluggableImplementationMap[name]; ok {
		return fmt.Errorf(
			"%s and %s are both registering as %s",
			i.GetName(),
			other.(*GoPlugin[I]).Impl.GetName(),
			name,
		)
	}

	pluggableImplementationMap[name] = &GoPlugin[I]{
		Name: name,
		Impl: i,
	}

	return nil
}

func Serve() {
	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluggableImplementationMap,
	})
}
