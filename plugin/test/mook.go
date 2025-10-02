// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package test // nolint:dupl

import (
	"log"
	"net/rpc"

	"github.com/veraison/services/plugin"
)

type IMook interface {
	GetName() string
	GetAttestationScheme() string
	GetSupportedMediaTypes() []string
	GetVersion() string
	Shoot() string
}

type MookRPCClient struct {
	client *rpc.Client
}

func (o *MookRPCClient) GetName() string {
	var (
		resp   string
		unused interface{}
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
		unused interface{}
	)

	err := o.client.Call("Plugin.GetAttestationScheme", &unused, &resp)
	if err != nil {
		log.Printf("Plugin.GetAttestationScheme RPC call failed: %v", err) // nolint
		return ""
	}

	return resp
}

func (o *MookRPCClient) GetSupportedMediaTypes() []string {
	var (
		resp   []string
		unused interface{}
	)

	err := o.client.Call("Plugin.GetSupportedMediaTypes", &unused, &resp)
	if err != nil {
		log.Printf("Plugin.GetSupportedMediaTypes RPC call failed: %v", err) // nolint
		return nil
	}

	return resp
}

func (o *MookRPCClient) GetVersion() string {
	var (
		resp   string
		unused interface{}
	)

	err := o.client.Call("Plugin.GetVersion", &unused, &resp)
	if err != nil {
		log.Printf("Plugin.GetVersion RPC call failed: %v", err) // nolint
		return ""
	}

	return resp
}

func (o *MookRPCClient) Shoot() string {
	var (
		resp   string
		unused interface{}
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

func (o *MookRPCServer) GetName(args interface{}, resp *string) error {
	*resp = o.Impl.GetName()
	return nil
}

func (o *MookRPCServer) GetAttestationScheme(args interface{}, resp *string) error {
	*resp = o.Impl.GetAttestationScheme()
	return nil
}

func (o *MookRPCServer) GetSupportedMediaTypes(args interface{}, resp *[]string) error {
	*resp = o.Impl.GetSupportedMediaTypes()
	return nil
}

func (o *MookRPCServer) GetVersion(args interface{}, resp *string) error {
	*resp = o.Impl.GetVersion()
	return nil
}

func (o *MookRPCServer) Shoot(args interface{}, resp *string) error {
	*resp = o.Impl.Shoot()
	return nil
}

func GetMookClient(c *rpc.Client) interface{} {
	return &MookRPCClient{client: c}
}

func GetMookServer(i IMook) interface{} {
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
