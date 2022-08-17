// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package scheme

import (
	"encoding/json"
	"fmt"
	"log"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"github.com/veraison/services/proto"
)

type Plugin struct {
	Impl IScheme
}

func (p *Plugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &RPCServer{Impl: p.Impl}, nil
}

func (p *Plugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RPCClient{client: c}, nil
}

type RPCServer struct {
	Impl IScheme
}

func (s *RPCServer) GetName(args interface{}, resp *string) error {
	*resp = s.Impl.GetName()
	return nil
}

func (s *RPCServer) GetFormat(args interface{}, resp *proto.AttestationFormat) error {
	*resp = s.Impl.GetFormat()
	return nil
}

func (s *RPCServer) GetSupportedMediaTypes(args interface{}, resp *[]string) error {
	*resp = s.Impl.GetSupportedMediaTypes()
	return nil
}

type SynthKeysArgs struct {
	TenantID        string
	EndorsementJSON []byte
}

func (s *RPCServer) SynthKeysFromSwComponent(args SynthKeysArgs, resp *[]string) error {
	var (
		err    error
		swComp proto.Endorsement
	)

	err = json.Unmarshal(args.EndorsementJSON, &swComp)
	if err != nil {
		return fmt.Errorf("unmarshaling software component: %w", err)
	}

	*resp, err = s.Impl.SynthKeysFromSwComponent(args.TenantID, &swComp)

	return err
}

func (s *RPCServer) SynthKeysFromTrustAnchor(args SynthKeysArgs, resp *[]string) error {
	var (
		err error
		ta  proto.Endorsement
	)

	err = json.Unmarshal(args.EndorsementJSON, &ta)
	if err != nil {
		return fmt.Errorf("unmarshaling trust anchor: %w", err)
	}

	*resp, err = s.Impl.SynthKeysFromTrustAnchor(args.TenantID, &ta)

	return err
}

func (s *RPCServer) GetTrustAnchorID(data []byte, resp *string) error {
	var (
		err   error
		token proto.AttestationToken
	)

	err = json.Unmarshal(data, &token)
	if err != nil {
		return fmt.Errorf("unmarshaling attestation token: %w", err)
	}

	*resp, err = s.Impl.GetTrustAnchorID(&token)

	return err
}

type ExtractVerifiedClaimsArgs struct {
	Token       []byte
	TrustAnchor string
}

func (s *RPCServer) ExtractVerifiedClaims(args ExtractVerifiedClaimsArgs, resp *[]byte) error {
	var token proto.AttestationToken

	err := json.Unmarshal(args.Token, &token)
	if err != nil {
		return fmt.Errorf("unmarshaling token: %w", err)
	}

	extracted, err := s.Impl.ExtractVerifiedClaims(&token, args.TrustAnchor)
	if err != nil {
		return err
	}

	*resp, err = json.Marshal(extracted)

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
		log.Printf("Plugin.GetName RPC call failed: %v", err) // nolint
		return ""
	}

	return resp
}

func (s *RPCClient) GetFormat() proto.AttestationFormat {
	var (
		err    error
		resp   proto.AttestationFormat
		unused interface{}
	)

	err = s.client.Call("Plugin.GetFormat", &unused, &resp)
	if err != nil {
		log.Printf("Plugin.GetFormat RPC call failed: %v", err)
		return proto.AttestationFormat_UNKNOWN_FORMAT
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
		log.Printf("Plugin.GetSupportedMediaTypes RPC call failed: %v", err)
		return nil
	}

	return resp
}

func (s *RPCClient) SynthKeysFromSwComponent(tenantID string, swComp *proto.Endorsement) ([]string, error) {
	var (
		err  error
		resp []string
		args SynthKeysArgs
	)

	args.TenantID = tenantID

	args.EndorsementJSON, err = json.Marshal(swComp)
	if err != nil {
		return nil, fmt.Errorf("marshaling software component: %w", err)
	}

	err = s.client.Call("Plugin.SynthKeysFromSwComponent", args, &resp)
	if err != nil {
		return nil, fmt.Errorf("Plugin.SynthKeysFromSwComponent RPC call failed: %w", err) // nolint
	}

	return resp, nil
}

func (s *RPCClient) SynthKeysFromTrustAnchor(tenantID string, ta *proto.Endorsement) ([]string, error) {
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
		return nil, fmt.Errorf("Plugin.SynthKeysFromTrustAnchor RPC call failed: %w", err) // nolint
	}

	return resp, nil
}

func (s *RPCClient) GetTrustAnchorID(token *proto.AttestationToken) (string, error) {
	var (
		err  error
		data []byte
		resp string
	)

	data, err = json.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("marshaling token: %w", err)
	}

	err = s.client.Call("Plugin.GetTrustAnchorID", data, &resp)
	if err != nil {
		return "", fmt.Errorf("Plugin.GetTrustAnchorID RPC call failed: %w", err) // nolint
	}

	return resp, nil
}

func (s *RPCClient) ExtractEvidence(token *proto.AttestationToken, trustAnchor string) (*ExtractedClaims, error) {
	var (
		err       error
		args      ExtractVerifiedClaimsArgs
		resp      []byte
		extracted ExtractedClaims
	)

	args.Token, err = json.Marshal(token)
	if err != nil {
		return nil, fmt.Errorf("marshaling token: %w", err)
	}
	args.TrustAnchor = trustAnchor

	err = s.client.Call("Plugin.ExtractEvidence", args, &resp)
	if err != nil {
		return nil, fmt.Errorf("Plugin.ExtractEvidence RPC call failed: %w", err) // nolint
	}

	err = json.Unmarshal(resp, &extracted)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling extracted evidence: %w", err)
	}

	return &extracted, nil
}

func (s *RPCClient) AppraiseEvidence(ec *proto.EvidenceContext, endorsements []string) (*proto.AppraisalContext, error) {
	var (
		args         AppraiseEvidenceArgs
		appraisalCtx proto.AppraisalContext
		err          error
		resp         []byte
	)

	args.Evidence, err = json.Marshal(ec)
	if err != nil {
		return nil, fmt.Errorf("marshaling evidence: %w", err)
	}

	args.Endorsements = endorsements

	err = s.client.Call("Plugin.AppraiseEvidence", args, &resp)
	if err != nil {
		return nil, fmt.Errorf("Plugin.AppraiseEvidence RPC call failed: %w", err) // nolint
	}

	err = json.Unmarshal(resp, &appraisalCtx)

	return &appraisalCtx, err
}

func (s *RPCClient) ExtractVerifiedClaims(token *proto.AttestationToken, trustAnchor string) (*ExtractedClaims, error) {
	var (
		err             error
		args            ExtractVerifiedClaimsArgs
		extractedClaims ExtractedClaims
	)

	args.Token, err = json.Marshal(token)
	if err != nil {
		return nil, fmt.Errorf("marshaling token: %w", err)
	}

	args.TrustAnchor = trustAnchor

	var resp []byte
	err = s.client.Call("Plugin.ExtractVerifiedClaims", args, &resp)
	if err != nil {
		return nil, fmt.Errorf("Plugin.ExtractVerifiedClaims RPC call failed: %w", err) // nolint
	}

	err = json.Unmarshal(resp, &extractedClaims)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling extracted claims: %w", err)
	}

	return &extractedClaims, nil
}
