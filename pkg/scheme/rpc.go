// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package scheme

import (
	"encoding/json"
	"fmt"
	"net/rpc"
	"reflect"

	"github.com/fxamacker/cbor/v2"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/ear"
	"github.com/veraison/services/pkg/fault"
	"github.com/veraison/services/pkg/log"
	"github.com/veraison/services/pkg/plugin"
	"go.uber.org/zap"
)

var SchemeHandlerRPC = &plugin.RPCChannel[IHandler]{
	GetClient: getSchemeClient,
	GetServer: getSchemeServer,
}

func getSchemeClient(c *rpc.Client) any {
	return &SchemeRPCClient{client: c, logger: log.Named("scheme-rpc")}
}

func getSchemeServer(i IHandler) any {
	return &SchemeRPCServer{Impl: i}
}

type SchemeRPCClient struct {
	client *rpc.Client
	logger *zap.SugaredLogger
}

func (o *SchemeRPCClient) Init(params *plugin.Parameters) error {
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

	return fault.ParseError(o.client.Call("Plugin.Init", args, &unused))
}

func (o *SchemeRPCClient) Fini() error {
	var unused any
	return fault.ParseError(o.client.Call("Plugin.Fini", &unused, &unused))
}

func (o *SchemeRPCClient) GetName() string {
	var (
		unused any
		resp   string
	)

	if err := o.client.Call("Plugin.GetName", &unused, &resp); err != nil {
		o.logger.Error(fault.ParseError(err))
		return ""
	}

	return resp
}

func (o *SchemeRPCClient) GetAttestationScheme() string {
	var (
		unused any
		resp   string
	)

	if err := o.client.Call("Plugin.GetAttestationScheme", &unused, &resp); err != nil {
		o.logger.Error(fault.ParseError(err))
		return ""
	}

	return resp
}

func (o *SchemeRPCClient) GetSupportedCorimProfiles() []string {
	var (
		unused any
		resp   []string
	)

	if err := o.client.Call("Plugin.GetSupportedCorimProfiles", &unused, &resp); err != nil {
		o.logger.Error(fault.ParseError(err))
		return []string{}
	}

	return resp
}

func (o *SchemeRPCClient) GetSupportedEvidenceMediaTypes() []string {
	var (
		unused any
		resp   []string
	)

	if err := o.client.Call("Plugin.GetSupportedEvidenceMediaTypes", &unused, &resp); err != nil {
		o.logger.Error(fault.ParseError(err))
		return []string{}
	}

	return resp
}

func (o *SchemeRPCClient) ValidateCorim(uc *corim.UnsignedCorim) (*ValidateCorimResponse, error) {
	toValidate, err := uc.ToCBOR()
	if err != nil {
		return nil, fmt.Errorf("mashalling CoRIM: %w", err)
	}

	var rawResp []byte
	if err = o.client.Call("Plugin.ValidateCorim", toValidate, &rawResp); err != nil {
		return nil, fault.ParseError(err)
	}

	var resp ValidateCorimResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (o *SchemeRPCClient) GetReferenceValueIDs(
	trustAnchors []*comid.KeyTriple,
	claims map[string]any,
) ([]*comid.Environment, error) {
	request := getReferenceValueIDsRequest{
		TrustAnchors: trustAnchors,
		Claims:       claims,
	}

	requestBytes, err := cbor.Marshal(request)
	if err != nil {
		return nil, err
	}

	var rawResp []byte
	if err = o.client.Call("Plugin.GetReferenceValueIDs", &requestBytes, &rawResp); err != nil {
		return nil, fault.ParseError(err)
	}

	var ret []*comid.Environment
	if err := cborDecoder.Unmarshal(rawResp, &ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func (o *SchemeRPCClient) ValidateEvidenceIntegrity(
	evidence *Evidence,
	trustAnchors []*comid.KeyTriple,
	endorsements []*comid.ValueTriple,
) error {
	request := validateEvidenceIntegrityRequest{
		Evidence:     evidence,
		TrustAnchors: trustAnchors,
		Endorsements: endorsements,
	}

	requestBytes, err := cbor.Marshal(request)
	if err != nil {
		return err
	}

	var unused []byte
	err = o.client.Call("Plugin.ValidateEvidenceIntegrity", &requestBytes, &unused)
	return fault.ParseError(err)
}

func (o *SchemeRPCClient) GetTrustAnchorIDs(
	evidence *Evidence,
) ([]*comid.Environment, error) {
	evidenceBytes, err := cbor.Marshal(evidence)
	if err != nil {
		return nil, err
	}

	var rawResp []byte
	if err := o.client.Call("Plugin.GetTrustAnchorIDs", &evidenceBytes, &rawResp); err != nil {
		return nil, fault.ParseError(err)
	}

	var ret []*comid.Environment
	if err := cborDecoder.Unmarshal(rawResp, &ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func (o *SchemeRPCClient) ExtractClaims(
	evidence *Evidence,
	trustAnchors []*comid.KeyTriple,
) (map[string]any, error) {
	request := extractClaimsRequest{
		Evidence:     evidence,
		TrustAnchors: trustAnchors,
	}

	requestBytes, err := cbor.Marshal(request)
	if err != nil {
		return nil, err
	}

	var resp []byte
	if err := o.client.Call("Plugin.ExtractClaims", &requestBytes, &resp); err != nil {
		return nil, fault.ParseError(err)
	}

	var claims map[string]any
	if err := cborDecoder.Unmarshal(resp, &claims); err != nil {
		return nil, err
	}

	return claims, nil
}

func (o *SchemeRPCClient) AppraiseClaims(
	claims map[string]any,
	endorsements []*comid.ValueTriple,
) (*ear.AttestationResult, error) {
	request := appraiseClaimsRequest{
		Claims:       claims,
		Endorsements: endorsements,
	}

	requestBytes, err := cbor.Marshal(request)
	if err != nil {
		return nil, err
	}

	var rawResp []byte
	if err := o.client.Call("Plugin.AppraiseClaims", &requestBytes, &rawResp); err != nil {
		return nil, fault.ParseError(err)
	}

	var ret ear.AttestationResult
	if err := json.Unmarshal(rawResp, &ret); err != nil {
		return nil, err
	}

	return &ret, nil
}

type SchemeRPCServer struct {
	Impl IHandler
}

func (o *SchemeRPCServer) Init(args []byte, resp *any) error {
	var params *plugin.Parameters
	var err error

	if args == nil {
		params = plugin.NewParameters()
	} else {
		if params, err = plugin.ParametersFromJSON(args); err != nil {
			return err
		}
	}

	return fault.SerializeError(o.Impl.Init(params))
}

func (o *SchemeRPCServer) Fini(args any, resp *any) error {
	return fault.SerializeError(o.Impl.Fini())
}

func (o *SchemeRPCServer) GetName(unused any, resp *string) error {
	*resp = o.Impl.GetName()
	return nil
}

func (o *SchemeRPCServer) GetAttestationScheme(unused any, resp *string) error {
	*resp = o.Impl.GetAttestationScheme()
	return nil
}

func (o *SchemeRPCServer) GetSupportedCorimProfiles(unused any, resp *[]string) error {
	*resp = o.Impl.GetSupportedCorimProfiles()
	return nil
}

func (o *SchemeRPCServer) GetSupportedEvidenceMediaTypes(unused any, resp *[]string) error {
	*resp = o.Impl.GetSupportedEvidenceMediaTypes()
	return nil
}

func (o *SchemeRPCServer) ValidateCorim(toValidate []byte, resp *[]byte) error {
	uc, err := corim.UnmarshalAndValidateUnsignedCorimFromCBOR(toValidate)
	if err != nil {
		*resp, err = json.Marshal(ValidateCorimResponse{
			IsValid: false,
			Message: err.Error(),
		})
		return err
	}

	ret, err := o.Impl.ValidateCorim(uc)
	if err != nil {
		return fault.SerializeError(err)
	}

	*resp, err = json.Marshal(ret)
	return err
}

func (o *SchemeRPCServer) GetReferenceValueIDs( // nolint:dupl
	requestBytes *[]byte,
	resp *[]byte,
) error {
	var request getReferenceValueIDsRequest
	if err := cborDecoder.Unmarshal(*requestBytes, &request); err != nil {
		return err
	}

	ret, err := o.Impl.GetReferenceValueIDs(request.TrustAnchors, request.Claims)
	if err != nil {
		return fault.SerializeError(err)
	}

	*resp, err = cbor.Marshal(ret)
	return err
}

func (o *SchemeRPCServer) ValidateEvidenceIntegrity(
	requestBytes *[]byte,
	unused *[]byte,
) error {
	var request validateEvidenceIntegrityRequest
	if err := cborDecoder.Unmarshal(*requestBytes, &request); err != nil {
		return err
	}

	err := o.Impl.ValidateEvidenceIntegrity(
		request.Evidence,
		request.TrustAnchors,
		request.Endorsements,
	)

	return fault.SerializeError(err)
}

func (o *SchemeRPCServer) GetTrustAnchorIDs(
	rawEvidence *[]byte,
	resp *[]byte,
) error {
	var evidence Evidence
	if err := cborDecoder.Unmarshal(*rawEvidence, &evidence); err != nil {
		return err
	}

	taIDs, err := o.Impl.GetTrustAnchorIDs(&evidence)
	if err != nil {
		return fault.SerializeError(err)
	}

	*resp, err = cbor.Marshal(taIDs)
	return err
}

func (o *SchemeRPCServer) ExtractClaims(
	requestBytes *[]byte,
	resp *[]byte,
) error {
	var request extractClaimsRequest
	if err := cborDecoder.Unmarshal(*requestBytes, &request); err != nil {
		return err
	}

	claims, err := o.Impl.ExtractClaims(request.Evidence, request.TrustAnchors)
	if err != nil {
		return fault.SerializeError(err)
	}

	*resp, err = cbor.Marshal(claims)
	return err
}

func (o *SchemeRPCServer) AppraiseClaims(
	requestBytes *[]byte,
	resp *[]byte,
) error {
	var request appraiseClaimsRequest
	if err := cborDecoder.Unmarshal(*requestBytes, &request); err != nil {
		return err
	}

	ret, err := o.Impl.AppraiseClaims(request.Claims, request.Endorsements)
	if err != nil {
		return fault.SerializeError(err)
	}

	*resp, err = json.Marshal(ret)
	return err
}

var cborDecoder cbor.DecMode

type extractClaimsRequest struct {
	Evidence     *Evidence          `cbor:"1,keyasint"`
	TrustAnchors []*comid.KeyTriple `cbor:"2,keyasint"`
}

type getReferenceValueIDsRequest struct {
	TrustAnchors []*comid.KeyTriple `cbor:"1,keyasint"`
	Claims       map[string]any     `cbor:"2,keyasint"`
}

type appraiseClaimsRequest struct {
	Endorsements []*comid.ValueTriple `cbor:"1,keyasint"`
	Claims       map[string]any       `cbor:"2,keyasint"`
}

type validateEvidenceIntegrityRequest struct {
	Evidence     *Evidence            `cbor:"1,keyasint"`
	TrustAnchors []*comid.KeyTriple   `cbor:"2,keyasint"`
	Endorsements []*comid.ValueTriple `cbor:"3,keyasint"`
}

func init() {
	decOpts := cbor.DecOptions{
		DefaultMapType: reflect.TypeFor[map[string]any](),
	}

	var err error
	cborDecoder, err = decOpts.DecMode()
	if err != nil {
		panic(err)
	}
}
