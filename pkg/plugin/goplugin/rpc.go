// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package goplugin

import (
	"fmt"
	"net/rpc"

	"github.com/veraison/services/pkg/plugin"
)

var rpcMap map[string]any

func init() {
	rpcMap = make(map[string]any)
}

func registerRPCChannel[I plugin.IPluggable](name string, ch *plugin.RPCChannel[I]) error {
	if _, ok := rpcMap[name]; ok {
		return fmt.Errorf("RPC channel for %q already registred", name)
	}

	rpcMap[name] = ch

	return nil
}

func GetRPCServer[I plugin.IPluggable](name string, impl I) any {
	i, ok := rpcMap[name]
	if !ok {
		return nil
	}

	ch := i.(*plugin.RPCChannel[I])

	return ch.GetServer(impl)
}

func GetRPCClient[I plugin.IPluggable](name string, impl I, c *rpc.Client) any {
	i, ok := rpcMap[name]
	if !ok {
		return nil
	}

	ch := i.(*plugin.RPCChannel[I])

	return ch.GetClient(c)
}
