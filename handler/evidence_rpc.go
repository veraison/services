// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"encoding/json"
	"fmt"
	"net/rpc"

	"github.com/veraison/ear"
	"github.com/veraison/services/log"
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/proto"
)

var EvidenceHandlerRPC = &plugin.RPCChannel[IEvidenceHandler]{
	GetClient: getClient,
	GetServer: getServer,
}

func getClient(c *rpc.Client) interface{} {
	return &RPCClient{client: c}
}

func getServer(i IEvidenceHandler) interface{} {
	return &RPCServer{Impl: i}
}

type RPCServer struct {
	Impl IEvidenceHandler
}

func (s *RPCServer) GetName(args interface{}, resp *string) error {
	*resp = s.Impl.GetName()
	return nil
}

func (s *RPCServer) GetAttestationScheme(args interface{}, resp *string) error {
	*resp = s.Impl.GetAttestationScheme()
	return nil
}

func (s *RPCServer) GetSupportedMediaTypes(args interface{}, resp *[]string) error {
	*resp = s.Impl.GetSupportedMediaTypes()
	return nil
}

type ExtractClaimsArgs struct {
	Token        []byte
	TrustAnchors []string
}

func (s *RPCServer) ExtractClaims(args ExtractClaimsArgs, resp *[]byte) error {
	var token proto.AttestationToken

	err := json.Unmarshal(args.Token, &token)
	if err != nil {
		return fmt.Errorf("unmarshaling token: %w", err)
	}

	extracted, err := s.Impl.ExtractClaims(&token, args.TrustAnchors)
	if err != nil {
		return err
	}

	*resp, err = json.Marshal(extracted)

	return err
}

type ValidateEvidenceIntegrityArgs struct {
	Token        []byte
	TrustAnchors []string
	Endorsements []string
}

func (s *RPCServer) ValidateEvidenceIntegrity(args ValidateEvidenceIntegrityArgs, resp *[]byte) error {
	var token proto.AttestationToken

	err := json.Unmarshal(args.Token, &token)
	if err != nil {
		return fmt.Errorf("unmarshaling token: %w", err)
	}

	err = s.Impl.ValidateEvidenceIntegrity(&token, args.TrustAnchors, args.Endorsements)

	return err
}

type AppraiseEvidenceArgs struct {
	Evidence     []byte
	Endorsements []string
}

func (s *RPCServer) AppraiseEvidence(args AppraiseEvidenceArgs, resp *[]byte) error {
	var (
		ec  proto.EvidenceContext
		err error
	)

	err = json.Unmarshal(args.Evidence, &ec)
	if err != nil {
		return fmt.Errorf("unmarshaling evidence: %w", err)
	}

	attestation, err := s.Impl.AppraiseEvidence(&ec, args.Endorsements)
	if err != nil {
		return err
	}

	*resp, err = json.Marshal(attestation)

	return err
}

type RPCClient struct {
	client *rpc.Client
}

func (s *RPCClient) GetName() string {
	var (
		resp   string
		unused interface{}
	)

	err := s.client.Call("Plugin.GetName", &unused, &resp)
	if err != nil {
		log.Errorf("Plugin.GetName RPC call failed: %v", err) // nolint
		return ""
	}

	return resp
}

func (s *RPCClient) GetAttestationScheme() string {
	var (
		resp   string
		unused interface{}
	)

	err := s.client.Call("Plugin.GetAttestationScheme", &unused, &resp)
	if err != nil {
		log.Errorf("Plugin.GetAttestationScheme RPC call failed: %v", err) // nolint
		return ""
	}

	return resp
}

func (s *RPCClient) GetSupportedMediaTypes() []string {
	var (
		err    error
		resp   []string
		unused interface{}
	)

	err = s.client.Call("Plugin.GetSupportedMediaTypes", &unused, &resp)
	if err != nil {
		log.Errorf("Plugin.GetSupportedMediaTypes RPC call failed: %v", err)
		return nil
	}

	return resp
}

func (s *RPCClient) ExtractEvidence(token *proto.AttestationToken, trustAnchors []string) (*ExtractedClaims, error) {
	var (
		err       error
		args      ExtractClaimsArgs
		resp      []byte
		extracted ExtractedClaims
	)

	args.Token, err = json.Marshal(token)
	if err != nil {
		return nil, fmt.Errorf("marshaling token: %w", err)
	}
	args.TrustAnchors = trustAnchors

	err = s.client.Call("Plugin.ExtractEvidence", args, &resp)
	if err != nil {
		err = ParseError(err)
		return nil, fmt.Errorf("Plugin.ExtractEvidence RPC call failed: %w", err) // nolint
	}

	err = json.Unmarshal(resp, &extracted)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling extracted evidence: %w", err)
	}

	return &extracted, nil
}

func (s *RPCClient) ValidateEvidenceIntegrity(
	token *proto.AttestationToken,
	trustAnchors []string,
	endorsements []string,
) error {
	var (
		err  error
		args ValidateEvidenceIntegrityArgs
		resp []byte
	)

	args.Token, err = json.Marshal(token)
	if err != nil {
		return fmt.Errorf("marshaling token: %w", err)
	}
	args.TrustAnchors = trustAnchors
	args.Endorsements = endorsements

	err = s.client.Call("Plugin.ValidateEvidenceIntegrity", args, &resp)

	return ParseError(err)
}

func (s *RPCClient) AppraiseEvidence(ec *proto.EvidenceContext, endorsements []string) (*ear.AttestationResult, error) {
	var (
		args   AppraiseEvidenceArgs
		result ear.AttestationResult
		err    error
		resp   []byte
	)

	args.Evidence, err = json.Marshal(ec)
	if err != nil {
		return nil, fmt.Errorf("marshaling evidence: %w", err)
	}

	args.Endorsements = endorsements

	err = s.client.Call("Plugin.AppraiseEvidence", args, &resp)
	if err != nil {
		err = ParseError(err)
		return nil, fmt.Errorf("Plugin.AppraiseEvidence RPC call failed: %w", err) // nolint
	}

	err = json.Unmarshal(resp, &result)

	return &result, err
}

func (s *RPCClient) ExtractClaims(token *proto.AttestationToken, trustAnchors []string) (*ExtractedClaims, error) {
	var (
		err             error
		args            ExtractClaimsArgs
		extractedClaims ExtractedClaims
	)

	args.Token, err = json.Marshal(token)
	if err != nil {
		return nil, fmt.Errorf("marshaling token: %w", err)
	}

	args.TrustAnchors = trustAnchors

	var resp []byte
	err = s.client.Call("Plugin.ExtractClaims", args, &resp)
	if err != nil {
		err = ParseError(err)
		return nil, fmt.Errorf("Plugin.ExtractClaims RPC call failed: %w", err) // nolint
	}

	err = json.Unmarshal(resp, &extractedClaims)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling extracted claims: %w", err)
	}

	return &extractedClaims, nil
}
