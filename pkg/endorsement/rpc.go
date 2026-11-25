// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package endorsement

import (
	"net/rpc"

	"github.com/fxamacker/cbor/v2"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/services/pkg/fault"
	"github.com/veraison/services/pkg/log"
	"github.com/veraison/services/pkg/plugin"
	"go.uber.org/zap"
)

var EndorsementStoreRPC = &plugin.RPCChannel[IHandler]{
	GetClient: getStoreClient,
	GetServer: getStoreServer,
}

func getStoreServer(i IHandler) any {
	return &StoreRPCServer{Impl: i}
}

func getStoreClient(c *rpc.Client) any {
	return &StoreRPCClient{
		client: c,
		logger: log.Named("endorsementstore-rpc"),
	}
}

type StoreRPCServer struct {
	Impl IHandler
}

func (o *StoreRPCServer) Init(args []byte, resp *any) error {
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

	return o.Impl.Init(params)
}

func (o *StoreRPCServer) Fini(args any, resp *any) error {
	return o.Impl.Fini()
}

func (o *StoreRPCServer) GetName(args any, resp *string) error {
	*resp = o.Impl.GetName()
	return nil
}

func (o *StoreRPCServer) Can(args *[]byte, resp *bool) error {
	var capabilities []Capability
	if err := cbor.Unmarshal(*args, &capabilities); err != nil {
		return err
	}

	*resp = o.Impl.Can(capabilities...)
	return nil
}

func (o *StoreRPCServer) GetValueTriples(params *GetEndorsementsRequest, resp *[]byte) error {
	var env comid.Environment
	if err := (&env).FromCBOR(params.Environment); err != nil {
		return err
	}
	refVals, err := o.Impl.GetValueTriples(&env, params.Label, params.MatchExactly)
	if err != nil {
		return err
	}
	*resp, err = cbor.Marshal(refVals)
	return err
}

func (o *StoreRPCServer) GetKeyTriples(params *GetEndorsementsRequest, resp *[]byte) error {
	var env comid.Environment
	if err := (&env).FromCBOR(params.Environment); err != nil {
		return err
	}
	keys, err := o.Impl.GetKeyTriples(&env, params.Label, params.MatchExactly)
	if err != nil {
		return err
	}
	*resp, err = cbor.Marshal(keys)
	return err
}

func (o *StoreRPCServer) ExecuteCoservQuery(params *ExecuteCoservQueryRequest, resp *[]byte) error {
	cos, err := o.Impl.ExecuteCoservQuery(params.MediaType, params.Query)
	if err != nil {
		return err
	}
	*resp, err = cos.ToCBOR()
	if err != nil {
		return err
	}
	return nil
}

func (o *StoreRPCServer) AddCorimBytes(params *AddCorimBytesRequest, resp *any) error {
	return o.Impl.AddCorimBytes(params.Endorsement, params.Label, params.Activate)
}

func (o *StoreRPCServer) SetEndorsementsState(params *SetEndorsementsStateRequest,
	resp *any) error {
	return o.Impl.SetEndorsementsState(params.TenantId, params.Data, params.SetActive)
}

type StoreRPCClient struct {
	client *rpc.Client
	logger *zap.SugaredLogger
}

func (o *StoreRPCClient) Init(params *plugin.Parameters) error {
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

	return o.client.Call("Plugin.Init", args, &unused)
}

func (o *StoreRPCClient) Fini() error {
	var unused any
	return o.client.Call("Plugin.Fini", &unused, &unused)
}

func (o *StoreRPCClient) GetName() string {
	var (
		resp   string
		unused any
	)

	err := o.client.Call("Plugin.GetName", &unused, &resp)
	if err != nil {
		log.Errorf("Plugin.GetName RPC call failed: %v", err)
		return ""
	}

	return resp
}

func (o *StoreRPCClient) Can(capabilities ...Capability) bool {
	args, err := cbor.Marshal(capabilities)
	if err != nil {
		panic(err)
	}

	var resp bool

	err = o.client.Call("Plugin.Can", &args, &resp)
	if err != nil {
		log.Errorf("Plugin.Can RPC call failed: %v", err)
		return false
	}

	return resp
}

func (o *StoreRPCClient) GetValueTriples( //nolint:dupl
	env *comid.Environment,
	label string,
	exact bool,
) ([]*comid.ValueTriple, error) {
	o.logger.Debugw("value triples lookup", "environment", env)
	envCbor, err := cbor.Marshal(env)
	if err != nil {
		return nil, err
	}
	args := GetEndorsementsRequest{
		Environment:  envCbor,
		Label:        label,
		MatchExactly: exact,
	}
	var rawResp []byte
	if err := o.client.Call("Plugin.GetValueTriples", &args, &rawResp); err != nil {
		return nil, fault.ParseError(err)
	}
	var ret []*comid.ValueTriple
	if err := cbor.Unmarshal(rawResp, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (o *StoreRPCClient) GetKeyTriples( //nolint:dupl
	env *comid.Environment,
	label string,
	exact bool,
) ([]*comid.KeyTriple, error) {
	o.logger.Debugw("key triples lookup", "environment", env)
	envCbor, err := cbor.Marshal(env)
	if err != nil {
		return nil, err
	}
	args := GetEndorsementsRequest{
		Environment:  envCbor,
		Label:        label,
		MatchExactly: exact,
	}
	var rawResp []byte
	if err := o.client.Call("Plugin.GetKeyTriples", &args, &rawResp); err != nil {
		return nil, fault.ParseError(err)
	}
	var ret []*comid.KeyTriple
	if err := cbor.Unmarshal(rawResp, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (o *StoreRPCClient) ExecuteCoservQuery(mediaType, query string) (*coserv.Coserv, error) {
	o.logger.Debugw("coserv request", "media-type", mediaType, "query", query)
	args := ExecuteCoservQueryRequest{
		MediaType: mediaType,
		Query:     query,
	}

	var rawResp []byte
	if err := o.client.Call("Plugin.ExecuteCoservQuery", &args, &rawResp); err != nil {
		return nil, fault.ParseError(err)
	}
	var ret coserv.Coserv
	if err := ret.FromCBOR(rawResp); err != nil {
		return nil, err
	}
	return &ret, nil
}

func (o *StoreRPCClient) AddCorimBytes(data []byte, label string, activate bool) error {
	args := AddCorimBytesRequest{
		Endorsement: data,
		Label:       label,
		Activate:    activate,
	}
	var unused any
	if err := o.client.Call("Plugin.AddCorimBytes", &args, &unused); err != nil {
		return fault.ParseError(err)
	}
	return nil
}

func (o *StoreRPCClient) SetEndorsementsState(label string, request []byte, setActive bool) error {
	args := SetEndorsementsStateRequest{
		TenantId:  label,
		Data:      request,
		SetActive: setActive,
	}
	var unused any
	if err := o.client.Call("Plugin.SetEndorsementsState", &args, &unused); err != nil {
		return fault.ParseError(err)
	}
	return nil
}
