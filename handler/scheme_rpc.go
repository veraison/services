// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"encoding/json"
	"fmt"
	"net/rpc"
	"reflect"

	"github.com/fxamacker/cbor/v2"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/ear"
	"github.com/veraison/services/log"
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/vts/appraisal"
	"go.uber.org/zap"
)

var SchemeHandlerRPC = &plugin.RPCChannel[ISchemeHandler]{
	GetClient: getSchemeClient,
	GetServer: getSchemeServer,
}

func getSchemeClient(c *rpc.Client) any {
	return &SchemeRPCClient{client: c, logger: log.Named("scheme-rpc")}
}

func getSchemeServer(i ISchemeHandler) any {
	return &SchemeRPCServer{Impl: i}
}

type SchemeRPCClient struct {
	client *rpc.Client
	logger *zap.SugaredLogger
}

func (o *SchemeRPCClient) GetName() string {
	var (
		unused any
		resp   string
	)

	if err := o.client.Call("Plugin.GetName", &unused, &resp); err != nil {
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
		return ""
	}

	return resp
}

func (o *SchemeRPCClient) GetSupportedMediaTypes() map[string][]string {
	var (
		unused any
		resp   []byte
	)

	if err := o.client.Call("Plugin.GetSupportedMediaTypes", &unused, &resp); err != nil {
		return nil
	}

	var ret map[string][]string
	if err := json.Unmarshal(resp, &ret); err != nil {
		o.logger.Error(err)
	}

	return ret
}

func (o *SchemeRPCClient) GetSupportedProvisioningMediaTypes() []string {
	var (
		unused any
		resp   []string
	)

	if err := o.client.Call("Plugin.GetSupportedProvisioningMediaTypes", &unused, &resp); err != nil {
		return []string{}
	}

	return resp
}

func (o *SchemeRPCClient) GetSupportedVerificationMediaTypes() []string {
	var (
		unused any
		resp   []string
	)

	if err := o.client.Call("Plugin.GetSupportedVerificationMediaTypes", &unused, &resp); err != nil {
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
		return nil, ParseError(err)
	}

	var resp ValidateCorimResponse
	if err = json.Unmarshal(rawResp, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (o *SchemeRPCClient) GetReferenceValueIDs(
	trustAnchors []*comid.KeyTriple,
	claims map[string]any,
) ([]*comid.Environment, error) {
	taCBOR, err := cbor.Marshal(trustAnchors)
	if err != nil {
		return nil, err
	}

	claimsCBOR, err := cbor.Marshal(claims)
	if err != nil {
		return nil, err
	}

	args := proto.GetReferenceValueIDsArgs{
		TrustAnchors: taCBOR,
		Claims: claimsCBOR,
	}

	var rawResp []byte
	if err = o.client.Call("Plugin.GetReferenceValueIDs", &args, &rawResp); err != nil {
		return nil, ParseError(err)
	}

	var ret []*comid.Environment
	if err := cbor.Unmarshal(rawResp, &ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func (o *SchemeRPCClient) ValidateEvidenceIntegrity(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
	endorsements []*comid.ValueTriple,
) error {
	taCBOR, err := cbor.Marshal(trustAnchors)
	if err != nil {
		return err
	}

	enCBOR, err := cbor.Marshal(endorsements)
	if err != nil {
		return err
	}

	args := proto.ValidateEvidenceIntegrityArgs{
		Evidence: evidence.ToProtobuf(),
		TrustAnchors: taCBOR,
		Endorsements: enCBOR,
	}

	var unused []byte
	err = o.client.Call("Plugin.ValidateEvidenceIntegrity", &args, &unused)
	return ParseError(err)
}

func (o *SchemeRPCClient) GetTrustAnchorIDs(
	evidence *appraisal.Evidence,
) ([]*comid.Environment, error) {
	args := evidence.ToProtobuf()

	var rawResp []byte
	if err := o.client.Call("Plugin.GetTrustAnchorIDs", &args, &rawResp); err != nil {
		return nil, ParseError(err)
	}

	var ret []*comid.Environment
	if err := cbor.Unmarshal(rawResp, &ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func (o *SchemeRPCClient) ExtractClaims(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
) (map[string]any, error) {
	taCBOR, err  := cbor.Marshal(trustAnchors)
	if err != nil {
		return nil, err
	}

	args := proto.ExtractClaimsArgs {
		Evidence: evidence.ToProtobuf(),
		TrustAnchors: taCBOR,
	}

	var resp []byte
	if err := o.client.Call("Plugin.ExtractClaims", &args, &resp); err != nil {
		return nil, ParseError(err)
	}

	var claims map[string]any
	if err := claimsDecMode.Unmarshal(resp, &claims); err != nil {
		return nil, err
	}

	return claims, nil
}

func (o *SchemeRPCClient) AppraiseClaims(
	claims map[string]any,
	endorsements []*comid.ValueTriple,
) (*ear.AttestationResult, error) {
	claimsCBOR, err := cbor.Marshal(claims)
	if err != nil {
		return nil, err
	}

	enCBOR, err := cbor.Marshal(endorsements)
	if err != nil {
		return nil, err
	}

	args := proto.AppraiseClaimsArgs{
		Claims: claimsCBOR,
		Endorsements: enCBOR,
	}

	var rawResp []byte
	if err := o.client.Call("Plugin.AppraiseClaims", &args, &rawResp); err != nil {
		return nil, ParseError(err)
	}

	var ret ear.AttestationResult
	if err := json.Unmarshal(rawResp, &ret); err != nil {
		return nil, err
	}

	return &ret, nil
}

type SchemeRPCServer struct {
	Impl ISchemeHandler
}

func (o *SchemeRPCServer) GetName(unused any, resp *string) error {
	*resp = o.Impl.GetName()
	return nil
}

func (o *SchemeRPCServer) GetAttestationScheme(unused any, resp *string) error {
	*resp = o.Impl.GetAttestationScheme()
	return nil
}

func (o *SchemeRPCServer) GetSupportedMediaTypes(unused any, resp *[]byte) error {
	var err error
	mts := o.Impl.GetSupportedMediaTypes()

	*resp, err = json.Marshal(mts)
	return err
}

func (o *SchemeRPCServer) GetSupportedProvisioningMediaTypes(unused any, resp *[]string) error {
	*resp = o.Impl.GetSupportedProvisioningMediaTypes()
	return nil
}

func (o *SchemeRPCServer) GetSupportedVerificationMediaTypes(unused any, resp *[]string) error {
	*resp = o.Impl.GetSupportedVerificationMediaTypes()
	return nil
}

func (o *SchemeRPCServer) ValidateCorim(toValidate []byte, resp *[]byte) error {
	uc, err := corim.UnmarshalAndValidateUnsignedCorimFromCBOR(toValidate)
	if err != nil {
		*resp, err = json.Marshal(ValidateCorimResponse{
			IsValid: false,
			Message: err.Error(),
		})
		return  err
	}


	ret, err := o.Impl.ValidateCorim(uc)
	if err != nil {
		return err
	}

	*resp, err = json.Marshal(ret)
	return err
}

func (o *SchemeRPCServer) GetReferenceValueIDs(
	params *proto.GetReferenceValueIDsArgs,
	resp *[]byte,
) error {
	var trustAnchors []*comid.KeyTriple
	if err := cbor.Unmarshal(params.TrustAnchors, &trustAnchors); err != nil {
		return err
	}

	var claims map[string]any
	if err := claimsDecMode.Unmarshal(params.Claims, &claims); err != nil {
		return err
	}

	ret, err := o.Impl.GetReferenceValueIDs(trustAnchors, claims)
	if err != nil {
		return err
	}

	*resp, err = cbor.Marshal(ret)
	return err
}

func (o *SchemeRPCServer) ValidateEvidenceIntegrity(
	params *proto.ValidateEvidenceIntegrityArgs,
	unused *[]byte,
) error {
	evidence := appraisal.NewEvidenceFromProtobuf(params.Evidence)

	var trustAnchors []*comid.KeyTriple
	if err := cbor.Unmarshal(params.TrustAnchors, &trustAnchors); err != nil {
		return err
	}

	var endorsements []*comid.ValueTriple
	if err := cbor.Unmarshal(params.Endorsements, &endorsements); err != nil {
		return err
	}

	return o.Impl.ValidateEvidenceIntegrity(evidence, trustAnchors, endorsements)
}

func (o *SchemeRPCServer) GetTrustAnchorIDs(
	params *proto.AttestationToken,
	resp *[]byte,
) error {
	evidence := appraisal.NewEvidenceFromProtobuf(params)

	taIDs, err := o.Impl.GetTrustAnchorIDs(evidence)
	if err != nil {
		return err
	}

	*resp, err = cbor.Marshal(taIDs)
	return err
}

func (o *SchemeRPCServer) ExtractClaims(
	params *proto.ExtractClaimsArgs,
	resp *[]byte,
) error {
	evidence := appraisal.NewEvidenceFromProtobuf(params.Evidence)

	var trustAnchors []*comid.KeyTriple
	if err := cbor.Unmarshal(params.TrustAnchors, &trustAnchors); err != nil {
		return err
	}

	claims, err := o.Impl.ExtractClaims(evidence, trustAnchors)
	if err != nil {
		return err
	}

	*resp, err = cbor.Marshal(claims)
	return err
}

func (o *SchemeRPCServer) AppraiseClaims(
	params *proto.AppraiseClaimsArgs,
	resp *[]byte,
) error {
	var endorsements []*comid.ValueTriple
	if err := cbor.Unmarshal(params.Endorsements, &endorsements); err != nil {
		return err
	}

	var claims map[string]any
	if err := claimsDecMode.Unmarshal(params.Claims, &claims); err != nil {
		return err
	}

	ret, err := o.Impl.AppraiseClaims(claims, endorsements)
	if err != nil {
		return err
	}

	*resp, err = json.Marshal(ret)
	return err
}

var claimsDecMode cbor.DecMode

func init() {
	decOpts := cbor.DecOptions{
		DefaultMapType: reflect.TypeFor[map[string]any](),
	}

	var err error
	claimsDecMode, err = decOpts.DecMode()
	if err != nil {
		panic(err)
	}
}
