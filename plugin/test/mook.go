// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package test // nolint:dupl

import (
	"encoding/json"
	"log"
	"net/rpc"

	"github.com/veraison/services/plugin"
)

type IMook interface {
	GetName() string
	GetAttestationScheme() string
	GetSupportedMediaTypes() map[string][]string
	Shoot() string
}

type MookRPCClient struct {
	client *rpc.Client
}

func (o *MookRPCClient) GetName() string {
	var (
		resp   string
		unused any
	)

	err := o.client.Call("Plugin.GetName", &unused, &resp)
	if err != nil {
		log.Printf("Plugin.GetName RPC call failed: %v", err) // nolint
		return ""
	}

	return resp
}

func (o *MookRPCClient) GetAttestationScheme() string {
	var (
		resp   string
		unused any
	)

	err := o.client.Call("Plugin.GetAttestationScheme", &unused, &resp)
	if err != nil {
		log.Printf("Plugin.GetAttestationScheme RPC call failed: %v", err) // nolint
		return ""
	}

	return resp
}

func (o *MookRPCClient) GetSupportedMediaTypes() map[string][]string {
	var (
		resp   []byte
		unused any
	)

	err := o.client.Call("Plugin.GetSupportedMediaTypes", &unused, &resp)
	if err != nil {
		log.Printf("Plugin.GetSupportedMediaTypes RPC call failed: %v", err) // nolint
		return nil
	}

	var res map[string][]string
	if err := json.Unmarshal(resp, &res); err != nil {
		panic(err)
	}

	return res
}

func (o *MookRPCClient) Shoot() string {
	var (
		resp   string
		unused any
	)

	err := o.client.Call("Plugin.Shoot", &unused, &resp)
	if err != nil {
		log.Printf("Plugin.Shoot RPC call failed: %v", err) // nolint
		return ""
	}

	return resp
}

type MookRPCServer struct {
	Impl IMook
}

func (o *MookRPCServer) GetName(args any, resp *string) error {
	*resp = o.Impl.GetName()
	return nil
}

func (o *MookRPCServer) GetAttestationScheme(args any, resp *string) error {
	*resp = o.Impl.GetAttestationScheme()
	return nil
}

func (o *MookRPCServer) GetSupportedMediaTypes(args any, resp *[]byte) error {
	var err error

	*resp, err = json.Marshal(o.Impl.GetSupportedMediaTypes())

	return err
}

func (o *MookRPCServer) Shoot(args any, resp *string) error {
	*resp = o.Impl.Shoot()
	return nil
}

func GetMookClient(c *rpc.Client) any {
	return &MookRPCClient{client: c}
}

func GetMookServer(i IMook) any {
	return &MookRPCServer{Impl: i}
}

var MookRPC = &plugin.RPCChannel[IMook]{
	GetClient: GetMookClient,
	GetServer: GetMookServer,
}

func RegisterMookImplementation(v IMook) {
	err := plugin.RegisterImplementation("mook", v, MookRPC)
	if err != nil {
		panic(err)
	}
}
