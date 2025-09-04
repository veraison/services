// Copyright 2022-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"encoding/json"
	"fmt"
	"net/rpc"

	"github.com/veraison/services/plugin"
)

/*
  Server-side RPC adapter around the Decoder plugin implementation
  (plugin-side)
*/

var EndorsementHandlerRPC = &plugin.RPCChannel[IEndorsementHandler]{
	GetClient: getEndorsementClient,
	GetServer: geEndorsementtServer,
}

func getEndorsementClient(c *rpc.Client) interface{} {
	return &EndorsementRPCClient{client: c}
}

func geEndorsementtServer(i IEndorsementHandler) interface{} {
	return &EndorsementRPCServer{Impl: i}
}

type EndorsementRPCServer struct {
	Impl IEndorsementHandler
}

func (s *EndorsementRPCServer) Init(params EndorsementHandlerParams, unused interface{}) error {
	return s.Impl.Init(params)
}

func (s EndorsementRPCServer) Close(unused0 interface{}, unused1 interface{}) error {
	return s.Impl.Close()
}

func (s *EndorsementRPCServer) GetName(args interface{}, resp *string) error {
	*resp = s.Impl.GetName()
	return nil
}

func (s *EndorsementRPCServer) GetAttestationScheme(args interface{}, resp *string) error {
	*resp = s.Impl.GetAttestationScheme()
	return nil
}

func (s *EndorsementRPCServer) GetSupportedMediaTypes(args interface{}, resp *[]string) error {
	*resp = s.Impl.GetSupportedMediaTypes()
	return nil
}

func (s EndorsementRPCServer) Decode(args []byte, resp *[]byte) error {
	var decodeArgs struct {
		Data       []byte
		MediaType  string
		CACertPool []byte
	}

	if err := json.Unmarshal(args, &decodeArgs); err != nil {
		return fmt.Errorf("failed to unmarshal decode arguments: %w", err)
	}

	j, err := s.Impl.Decode(decodeArgs.Data, decodeArgs.MediaType, decodeArgs.CACertPool)
	if err != nil {
		return fmt.Errorf("plugin %q returned error: %w", s.Impl.GetName(), err)
	}

	*resp, err = json.Marshal(j)
	if err != nil {
		return fmt.Errorf("failed to marshal plugin response: %w", err)
	}

	return nil
}

/*
  RPC client
  (plugin caller side)
*/

type EndorsementRPCClient struct {
	client *rpc.Client
}

func (c EndorsementRPCClient) Init(params EndorsementHandlerParams) error {
	var unused interface{}

	return c.client.Call("Plugin.Init", params, &unused)
}

func (c EndorsementRPCClient) Close() error {
	var (
		unused0 interface{}
		unused1 interface{}
	)

	return c.client.Call("Plugin.Close", unused0, unused1)
}

func (c EndorsementRPCClient) GetName() string {
	var (
		err    error
		resp   string
		unused interface{}
	)

	err = c.client.Call("Plugin.GetName", &unused, &resp)
	if err != nil {
		return ""
	}

	return resp
}

func (c EndorsementRPCClient) GetAttestationScheme() string {
	var (
		err    error
		resp   string
		unused interface{}
	)

	err = c.client.Call("Plugin.GetAttestationScheme", &unused, &resp)
	if err != nil {
		return ""
	}

	return resp
}

func (c EndorsementRPCClient) GetSupportedMediaTypes() []string {
	var (
		err    error
		resp   []string
		unused interface{}
	)

	err = c.client.Call("Plugin.GetSupportedMediaTypes", &unused, &resp)
	if err != nil {
		return nil
	}

	return resp
}

func (c EndorsementRPCClient) Decode(data []byte, mediaType string, caCertPool []byte) (*EndorsementHandlerResponse, error) {
	var (
		err  error
		resp EndorsementHandlerResponse
		j    []byte
	)

	decodeArgs := struct {
		Data       []byte
		MediaType  string
		CACertPool []byte
	}{
		Data:       data,
		MediaType:  mediaType,
		CACertPool: caCertPool,
	}

	args, err := json.Marshal(decodeArgs)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling RPC arguments: %w", err)
	}

	err = c.client.Call("Plugin.Decode", args, &j)
	if err != nil {
		return nil, fmt.Errorf("RPC server returned error: %w", err)
	}

	err = json.Unmarshal(j, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshaling response from RPC server: %w", err)
	}

	return &resp, nil
}
