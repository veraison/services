// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"encoding/json"
	"net/rpc"

	"github.com/fxamacker/cbor/v2"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/services/log"
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/proto"
	"go.uber.org/zap"
)

var EndorsementStoreRPC = &plugin.RPCChannel[IEndorsementStorePlugin]{
	GetClient: getStoreClient,
	GetServer: getStoreServer,
}

func getStoreServer(i IEndorsementStorePlugin) any {
	return &StoreRPCServer{Impl: i}
}

func getStoreClient(c *rpc.Client) any {
	return &StoreRPCClient{
		client: c,
		logger: log.Named("endorsementstore-rpc"),
	}
}

type StoreRPCServer struct {
	Impl IEndorsementStorePlugin
}

func (s *StoreRPCServer) Init(args []byte, resp *any) error {
	var (
		params *plugin.Parameters
		err    error
	)
	if args == nil {
		params = plugin.NewParameters()
	} else {
		if params, err = plugin.ParametersFromJSON(args); err != nil {
			return err
		}
	}

	return s.Impl.Init(params)
}

func (s *StoreRPCServer) Fini(args any, resp *any) error {
	return s.Impl.Fini()
}

func (s *StoreRPCServer) GetName(args any, resp *string) error {
	*resp = s.Impl.GetName()
	return nil
}

func (s *StoreRPCServer) GetAttestationScheme(args any, resp *string) error {
	*resp = s.Impl.GetAttestationScheme()
	return nil
}

func (s *StoreRPCServer) GetSupportedMediaTypes(args any, resp *[]byte) error {
	var err error
	mts := s.Impl.GetSupportedMediaTypes()

	*resp, err = json.Marshal(mts)
	return err
}

func (s *StoreRPCServer) GetValueTriples(params *proto.GetEndorsementsArgs, resp *[]byte) error {
	var env comid.Environment
	if err := (&env).FromCBOR(params.Environment); err != nil {
		return err
	}
	refVals, err := s.Impl.GetValueTriples(&env, params.Label, params.MatchExactly)
	if err != nil {
		return err
	}
	*resp, err = cbor.Marshal(refVals)
	return err
}

func (s *StoreRPCServer) GetKeyTriples(params *proto.GetEndorsementsArgs, resp *[]byte) error {
	var env comid.Environment
	if err := (&env).FromCBOR(params.Environment); err != nil {
		return err
	}
	keys, err := s.Impl.GetKeyTriples(&env, params.Label, params.MatchExactly)
	if err != nil {
		return err
	}
	*resp, err = cbor.Marshal(keys)
	return err
}

func (s *StoreRPCServer) ExecuteCoservQuery(params *proto.EndorsementQueryIn, resp *[]byte) error {
	cos, err := s.Impl.ExecuteCoservQuery(params.MediaType, params.Query)
	if err != nil {
		return err
	}
	*resp, err = cos.ToCBOR()
	if err != nil {
		return err
	}
	return nil
}

func (s *StoreRPCServer) AddCorimBytes(params *proto.SubmitEndorsementsArgs, resp *any) error {
	return s.Impl.AddCorimBytes(params.Endorsement, params.Label, params.Activate)
}

type StoreRPCClient struct {
	client *rpc.Client
	logger *zap.SugaredLogger
}

func (c *StoreRPCClient) Init(params *plugin.Parameters) error {
	var (
		unused any
		args   []byte
		err    error
	)

	if params != nil {
		if args, err = params.MarshalJSON(); err != nil {
			return err
		}
	}

	return c.client.Call("Plugin.Init", args, &unused)
}

func (c *StoreRPCClient) Fini() error {
	var unused any
	return c.client.Call("Plugin.Fini", &unused, &unused)
}

func (c *StoreRPCClient) GetName() string {
	var (
		resp   string
		unused any
	)

	err := c.client.Call("Plugin.GetName", &unused, &resp)
	if err != nil {
		log.Errorf("Plugin.GetName RPC call failed: %v", err)
		return ""
	}

	return resp
}

func (c *StoreRPCClient) GetAttestationScheme() string {
	var (
		resp   string
		unused any
	)

	err := c.client.Call("Plugin.GetAttestationScheme", &unused, &resp)
	if err != nil {
		log.Errorf("Plugin.GetAttestationScheme RPC call failed: %v", err)
		return ""
	}

	return resp
}

func (c *StoreRPCClient) GetSupportedMediaTypes() map[string][]string {
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

func (c *StoreRPCClient) GetValueTriples(env *comid.Environment, label string, exact bool) ([]*comid.ValueTriple, error) {
	c.logger.Debugw("value triples lookup", "environment", env)
	envCbor, err := cbor.Marshal(env)
	if err != nil {
		return nil, err
	}
	args := proto.GetEndorsementsArgs{
		Environment:  envCbor,
		Label:        label,
		MatchExactly: exact,
	}
	var rawResp []byte
	if err := c.client.Call("Plugin.GetValueTriples", &args, &rawResp); err != nil {
		return nil, ParseError(err)
	}
	var ret []*comid.ValueTriple
	if err := cbor.Unmarshal(rawResp, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *StoreRPCClient) GetKeyTriples(env *comid.Environment, label string, exact bool) ([]*comid.KeyTriple, error) {
	c.logger.Debugw("key triples lookup", "environment", env)
	envCbor, err := cbor.Marshal(env)
	if err != nil {
		return nil, err
	}
	args := proto.GetEndorsementsArgs{
		Environment:  envCbor,
		Label:        label,
		MatchExactly: exact,
	}
	var rawResp []byte
	if err := c.client.Call("Plugin.GetKeyTriples", &args, &rawResp); err != nil {
		return nil, ParseError(err)
	}
	var ret []*comid.KeyTriple
	if err := cbor.Unmarshal(rawResp, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *StoreRPCClient) ExecuteCoservQuery(mediaType, query string) (*coserv.Coserv, error) {
	c.logger.Debugw("coserv request", "media-type", mediaType, "query", query)
	args := proto.EndorsementQueryIn{
		MediaType: mediaType,
		Query:     query,
	}

	var rawResp []byte
	if err := c.client.Call("Plugin.ExecuteCoservQuery", &args, &rawResp); err != nil {
		return nil, ParseError(err)
	}
	var ret coserv.Coserv
	if err := ret.FromCBOR(rawResp); err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *StoreRPCClient) AddCorimBytes(data []byte, label string, activate bool) error {
	args := proto.SubmitEndorsementsArgs{
		Endorsement: data,
		Label:       label,
		Activate:    activate,
	}
	var unused any
	if err := c.client.Call("Plugin.AddCorimBytes", &args, &unused); err != nil {
		return ParseError(err)
	}
	return nil
}
