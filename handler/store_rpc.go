// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"encoding/json"
	"fmt"
	"net/rpc"

	"github.com/veraison/services/plugin"
	"github.com/veraison/services/proto"
)

/*
  Server-side RPC adapter around the Decoder plugin implementation
  (plugin-side)
*/

var StoreHandlerRPC = &plugin.RPCChannel[IStoreHandler]{
	GetClient: getStoreClient,
	GetServer: getStoreServer,
}

func getStoreClient(c *rpc.Client) interface{} {
	return &StoreRPCClient{client: c}
}

func getStoreServer(i IStoreHandler) interface{} {
	return &StoreRPCServer{Impl: i}
}

type StoreRPCServer struct {
	Impl IStoreHandler
}

func (s *StoreRPCServer) GetName(args interface{}, resp *string) error {
	*resp = s.Impl.GetName()
	return nil
}

func (s *StoreRPCServer) GetAttestationScheme(args interface{}, resp *string) error {
	*resp = s.Impl.GetAttestationScheme()
	return nil
}

func (s *StoreRPCServer) GetSupportedMediaTypes(args interface{}, resp *[]string) error {
	*resp = s.Impl.GetSupportedMediaTypes()
	return nil
}

type SynthKeysArgs struct {
	TenantID        string
	EndorsementJSON []byte
}

func (s *StoreRPCServer) SynthKeysFromRefValue(args SynthKeysArgs, resp *[]string) error {
	var (
		err    error
		refVal Endorsement
	)

	err = json.Unmarshal(args.EndorsementJSON, &refVal)
	if err != nil {
		return fmt.Errorf("unmarshaling reference value: %w", err)
	}

	*resp, err = s.Impl.SynthKeysFromRefValue(args.TenantID, &refVal)

	return err
}

func (s *StoreRPCServer) SynthKeysFromTrustAnchor(args SynthKeysArgs, resp *[]string) error {
	var (
		err error
		ta  Endorsement
	)

	err = json.Unmarshal(args.EndorsementJSON, &ta)
	if err != nil {
		return fmt.Errorf("unmarshaling trust anchor: %w", err)
	}

	*resp, err = s.Impl.SynthKeysFromTrustAnchor(args.TenantID, &ta)

	return err
}

func (s *StoreRPCServer) GetTrustAnchorIDs(data []byte, resp *[]string) error {
	var (
		err   error
		token proto.AttestationToken
	)

	err = json.Unmarshal(data, &token)
	if err != nil {
		return fmt.Errorf("unmarshaling attestation token: %w", err)
	}

	*resp, err = s.Impl.GetTrustAnchorIDs(&token)

	return err
}

type GetRefValueIDsArgs struct {
	TenantID     string
	TrustAnchors []string
	Claims       []byte
}

func (s *StoreRPCServer) GetRefValueIDs(args GetRefValueIDsArgs, resp *[]string) error {
	var claims map[string]interface{}

	err := json.Unmarshal(args.Claims, &claims)
	if err != nil {
		return fmt.Errorf("unmarshaling token: %w", err)
	}

	*resp, err = s.Impl.GetRefValueIDs(args.TenantID, args.TrustAnchors, claims)
	if err != nil {
		return err
	}

	return err
}

/*
  RPC client
  (plugin caller side)
*/

type StoreRPCClient struct {
	client *rpc.Client
}

func (c StoreRPCClient) Close() error {
	var (
		unused0 interface{}
		unused1 interface{}
	)

	return c.client.Call("Plugin.Close", unused0, unused1)
}

func (c StoreRPCClient) GetName() string {
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

func (c StoreRPCClient) GetAttestationScheme() string {
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

func (c StoreRPCClient) GetSupportedMediaTypes() []string {
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

func (s *StoreRPCClient) SynthKeysFromRefValue(tenantID string, refVal *Endorsement) ([]string, error) {
	var (
		err  error
		resp []string
		args SynthKeysArgs
	)

	args.TenantID = tenantID

	args.EndorsementJSON, err = json.Marshal(refVal)
	if err != nil {
		return nil, fmt.Errorf("marshaling reference value: %w", err)
	}

	err = s.client.Call("Plugin.SynthKeysFromRefValue", args, &resp)
	if err != nil {
		err = ParseError(err)
		return nil, fmt.Errorf("Plugin.SynthKeysFromRefValue RPC call failed: %w", err) // nolint
	}

	return resp, nil
}

func (s *StoreRPCClient) SynthKeysFromTrustAnchor(tenantID string, ta *Endorsement) ([]string, error) {
	var (
		err  error
		resp []string
		args SynthKeysArgs
	)

	args.TenantID = tenantID

	args.EndorsementJSON, err = json.Marshal(ta)
	if err != nil {
		return nil, fmt.Errorf("marshaling trust anchor: %w", err)
	}

	err = s.client.Call("Plugin.SynthKeysFromTrustAnchor", args, &resp)
	if err != nil {
		err = ParseError(err)
		return nil, fmt.Errorf("Plugin.SynthKeysFromTrustAnchor RPC call failed: %w", err) // nolint
	}

	return resp, nil
}

func (s *StoreRPCClient) GetTrustAnchorIDs(token *proto.AttestationToken) ([]string, error) {
	var (
		err  error
		data []byte
		resp []string
	)

	data, err = json.Marshal(token)
	if err != nil {
		return []string{""}, fmt.Errorf("marshaling token: %w", err)
	}

	err = s.client.Call("Plugin.GetTrustAnchorIDs", data, &resp)
	if err != nil {
		err = ParseError(err)
		return []string{""}, fmt.Errorf("Plugin.GetTrustAnchorIDs RPC call failed: %w", err) // nolint
	}

	return resp, nil
}

func (s *StoreRPCClient) GetRefValueIDs(
	tenantID string,
	trustAnchors []string,
	claims map[string]interface{},
) ([]string, error) {
	var (
		err  error
		resp []string
	)

	args := GetRefValueIDsArgs{
		TenantID:     tenantID,
		TrustAnchors: trustAnchors,
	}

	args.Claims, err = json.Marshal(claims)
	if err != nil {
		return nil, err
	}

	err = s.client.Call("Plugin.GetRefValueIDs", args, &resp)
	if err != nil {
		err = ParseError(err)
		return nil, fmt.Errorf("Plugin.GetRefValueIDs RPC call failed: %w", err) // nolint
	}

	return resp, nil
}
