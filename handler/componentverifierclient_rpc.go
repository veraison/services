// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"fmt"
	"net/rpc"

	"github.com/veraison/services/log"
	"github.com/veraison/services/plugin"
)

var ComponentVerifierClientHandlerRPC = &plugin.RPCChannel[IComponentVerifierClientHandler]{
	GetClient: getComponentVerifierClientHandler,
	GetServer: getComponentVerifierServer,
}

func getComponentVerifierClientHandler(c *rpc.Client) interface{} {
	return &ComponentVerifierClientHandlerRPCClient{client: c}
}

func getComponentVerifierServer(i IComponentVerifierClientHandler) interface{} {
	return &ComponentVerifierClientHandlerRPCServer{Impl: i}
}

type ComponentVerifierClientHandlerRPCServer struct {
	Impl IComponentVerifierClientHandler
}

func (s *ComponentVerifierClientHandlerRPCServer) GetName(args interface{}, resp *string) error {
	*resp = s.Impl.GetName()
	return nil
}

func (s *ComponentVerifierClientHandlerRPCServer) GetAttestationScheme(args interface{}, resp *string) error {
	*resp = s.Impl.GetAttestationScheme()
	return nil
}

func (s *ComponentVerifierClientHandlerRPCServer) GetSupportedMediaTypes(args interface{}, resp *[]string) error {
	*resp = s.Impl.GetSupportedMediaTypes()
	return nil
}

type AppraiseComponentEvidenceArgs struct {
	Evidence  []byte
	MediaType string
	Nonce     []byte
	ClientCfg []byte
}

func (s *ComponentVerifierClientHandlerRPCServer) AppraiseComponentEvidence(args AppraiseComponentEvidenceArgs, resp *[]byte) error {
	appraisals, err := s.Impl.AppraiseComponentEvidence(args.Evidence, args.MediaType, args.Nonce, args.ClientCfg)

	*resp = appraisals

	return err
}

type ComponentVerifierClientHandlerRPCClient struct {
	client *rpc.Client
}

func (c *ComponentVerifierClientHandlerRPCClient) GetName() string {
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

func (c *ComponentVerifierClientHandlerRPCClient) GetAttestationScheme() string {
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

func (c *ComponentVerifierClientHandlerRPCClient) GetSupportedMediaTypes() []string {
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

func (c *ComponentVerifierClientHandlerRPCClient) AppraiseComponentEvidence(evidence []byte, mediaType string, nonce []byte) ([]byte, error) {
	var (
		resp []byte
		err  error
	)

	args := AppraiseComponentEvidenceArgs{
		Evidence:  evidence,
		MediaType: mediaType,
		Nonce:     nonce,
	}

	err = c.client.Call("Plugin.AppraiseComponentEvidence", args, &resp)
	if err != nil {
		return nil, fmt.Errorf("Plugin.AppraiseComponentEvidence RPC call failed: %w", err)
	}

	return resp, nil
}
