// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package plugin

import (
	"fmt"
	"net/rpc"

	"go.uber.org/zap"

	"github.com/veraison/services/log"
)

type RPCChannel[I IPluggable] struct {
	GetClient func(c *rpc.Client) interface{}
	GetServer func(i I) interface{}
}

var rpcMap map[string]interface{}
var logger *zap.SugaredLogger

func init() {
	rpcMap = make(map[string]interface{})
}

// note: we cannot simply initialize logger
func getLogger() *zap.SugaredLogger {
	if logger == nil {
		logger = log.Named("plugin.rpc")
	}
	return logger
}

func registerRPCChannel[I IPluggable](name string, ch *RPCChannel[I]) error {
	if _, ok := rpcMap[name]; ok {
		return fmt.Errorf("RPC channel for %q already registred", name)
	}

	rpcMap[name] = ch
	getLogger().Debugw("registered RPC channel", "name", name)

	return nil
}

func GetRPCServer[I IPluggable](name string, impl I) interface{} {
	getLogger().Debugw("GetRPCServer", "name", name, "type", GetTypeName[I]())
	i, ok := rpcMap[name]
	if !ok {
		return nil
	}

	ch := i.(*RPCChannel[I])

	return ch.GetServer(impl)
}

func GetRPCClient[I IPluggable](name string, impl I, c *rpc.Client) interface{} {
	getLogger().Debugw("GetRPCClient", "name", name, "type", GetTypeName[I]())
	i, ok := rpcMap[name]
	if !ok {
		return nil
	}

	ch := i.(*RPCChannel[I])

	return ch.GetClient(c)
}
