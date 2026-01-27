// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"encoding/json"
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

func (s *CoservProxyRPCServer) GetSupportedMediaTypes(args interface{}, resp *[]byte) error {
	var err error
	mts := s.Impl.GetSupportedMediaTypes()

	*resp, err = json.Marshal(mts)
	return err
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

func (c *CoservProxyRPCClient) GetSupportedMediaTypes() map[string][]string {
	var (
		resp   []byte
		unused any
	)

	err := c.client.Call("Plugin.GetSupportedMediaTypes", &unused, &resp)
	if err != nil {
		log.Errorf("Plugin.GetSupportedMediaTypes RPC call failed: %v", err)
		return nil
	}

	var ret map[string][]string
	if err := json.Unmarshal(resp, &ret); err != nil {
		log.Error(err)
	}

	return ret
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
