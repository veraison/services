// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"fmt"
	"net/rpc"

	"github.com/veraison/services/log"
	"github.com/veraison/services/plugin"
)

var CoservProxyHandlerRPC = &plugin.RPCChannel[ICoservProxyHandler]{
	GetClient: getCoservClient,
	GetServer: getCoservServer,
}

func getCoservClient(c *rpc.Client) interface{} {
	return &CoservProxyRPCClient{client: c}
}

func getCoservServer(i ICoservProxyHandler) interface{} {
	return &CoservProxyRPCServer{Impl: i}
}

type CoservProxyRPCServer struct {
	Impl ICoservProxyHandler
}

func (s *CoservProxyRPCServer) GetName(args interface{}, resp *string) error {
	*resp = s.Impl.GetName()
	return nil
}

func (s *CoservProxyRPCServer) GetAttestationScheme(args interface{}, resp *string) error {
	*resp = s.Impl.GetAttestationScheme()
	return nil
}

func (s *CoservProxyRPCServer) GetSupportedMediaTypes(args interface{}, resp *[]string) error {
	*resp = s.Impl.GetSupportedMediaTypes()
	return nil
}

type GetEndorsementArgs struct {
	TenantID string
	Query    string
}

func (s *CoservProxyRPCServer) GetEndorsements(args GetEndorsementArgs, resp *[]byte) (err error) {
	*resp, err = s.Impl.GetEndorsements(args.TenantID, args.Query)
	return err
}

type CoservProxyRPCClient struct {
	client *rpc.Client
}

func (c *CoservProxyRPCClient) GetName() string {
	var (
		resp   string
		unused interface{}
	)

	err := c.client.Call("Plugin.GetName", &unused, &resp)
	if err != nil {
		log.Errorf("Plugin.GetName RPC call failed: %v", err)
		return ""
	}

	return resp
}

func (c *CoservProxyRPCClient) GetAttestationScheme() string {
	var (
		resp   string
		unused interface{}
	)

	err := c.client.Call("Plugin.GetAttestationScheme", &unused, &resp)
	if err != nil {
		log.Errorf("Plugin.GetAttestationScheme RPC call failed: %v", err)
		return ""
	}

	return resp
}

func (c *CoservProxyRPCClient) GetSupportedMediaTypes() []string {
	var (
		resp   []string
		unused interface{}
	)

	err := c.client.Call("Plugin.GetSupportedMediaTypes", &unused, &resp)
	if err != nil {
		log.Errorf("Plugin.GetSupportedMediaTypes RPC call failed: %v", err)
		return nil
	}

	return resp
}

func (c *CoservProxyRPCClient) GetEndorsements(
	tenantID string,
	query string,
) (resp []byte, err error) {
	args := GetEndorsementArgs{TenantID: tenantID, Query: query}

	err = c.client.Call("Plugin.GetEndorsements", args, &resp)
	if err != nil {
		return nil, fmt.Errorf("Plugin.GetEndorsements RPC call failed: %w", ParseError(err))
	}

	return resp, nil
}
