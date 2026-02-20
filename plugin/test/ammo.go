// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package test // nolint:dupl

import (
	"encoding/json"
	"log"
	"net/rpc"

	"github.com/veraison/services/plugin"
)

type IAmmo interface {
	GetName() string
	GetAttestationScheme() string
	GetSupportedMediaTypes() map[string][]string
	GetCapacity() int
}

type AmmoRPCClient struct {
	client *rpc.Client
}

func (o *AmmoRPCClient) GetName() string {
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

func (o *AmmoRPCClient) GetAttestationScheme() string {
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

func (o *AmmoRPCClient) GetSupportedMediaTypes() map[string][]string {
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

func (o *AmmoRPCClient) GetCapacity() int {
	var (
		resp   int
		unused any
	)

	err := o.client.Call("Plugin.GetCapacity", &unused, &resp)
	if err != nil {
		log.Printf("Plugin.GetCapacity RPC call failed: %v", err) // nolint
		return 0
	}

	return resp
}

type AmmoRPCServer struct {
	Impl IAmmo
}

func (o *AmmoRPCServer) GetName(args any, resp *string) error {
	*resp = o.Impl.GetName()
	return nil
}

func (o *AmmoRPCServer) GetAttestationScheme(args any, resp *string) error {
	*resp = o.Impl.GetAttestationScheme()
	return nil
}

func (o *AmmoRPCServer) GetSupportedMediaTypes(args any, resp *[]byte) error {
	var err error

	*resp, err = json.Marshal(o.Impl.GetSupportedMediaTypes())

	return err
}

func (o *AmmoRPCServer) GetCapacity(args any, resp *int) error {
	*resp = o.Impl.GetCapacity()
	return nil
}

func GetAmmoClient(c *rpc.Client) any {
	return &AmmoRPCClient{client: c}
}

func GetAmmoServer(i IAmmo) any {
	return &AmmoRPCServer{Impl: i}
}

var AmmoRPC = &plugin.RPCChannel[IAmmo]{
	GetClient: GetAmmoClient,
	GetServer: GetAmmoServer,
}

func RegisterAmmoImplementation(v IAmmo) {
	err := plugin.RegisterImplementation("ammo", v, AmmoRPC)
	if err != nil {
		panic(err)
	}
}
