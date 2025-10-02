// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package plugin

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-plugin"
	"go.uber.org/zap"

	"github.com/veraison/services/log"
)

// IPluginContext is the common interace for handling all PluginContext[I] type
// instances of the generic PluginContext[].
type IPluginContext interface {
	GetName() string
	GetAttestationScheme() string
	GetTypeName() string
	GetPath() string
	GetHandle() interface{}
	GetVersion() string
	Close()
}

// PluginConntext is a generic for handling Veraison services plugins. It is
// parameterised on the IPluggale interface it handles.
type PluginContext[I IPluggable] struct {
	// Path to the exectable binary containing the plugin implementation
	Path string
	// Name of this plugin
	Name string
	// Name of the attestatin scheme implemented by this plugin
	Scheme string
	// Version of this plugin implementation
	Version string
	// SupportedMediaTypes are the types of input this plugin can process.
	// This is is the method by which a plugin is selected.
	SupportedMediaTypes []string
	// Handle is actual RPC interface to the plugin implementation.
	Handle I

	// go-plugin client
	client *plugin.Client
}

func (o PluginContext[I]) GetName() string {
	return o.Name
}

func (o PluginContext[I]) GetAttestationScheme() string {
	return o.Scheme
}

func (o PluginContext[I]) GetTypeName() string {
	return GetTypeName[I]()
}

func (o PluginContext[I]) GetPath() string {
	return o.Path
}

func (o PluginContext[I]) GetHandle() interface{} {
	return o.Handle
}

func (o PluginContext[I]) GetVersion() string {
	return o.Version
}

func (o PluginContext[I]) Close() {
	if o.client != nil {
		o.client.Kill()
	}
}

func createPluginContext[I IPluggable](
	loader *GoPluginLoader,
	path string,
	logger *zap.SugaredLogger,
) (*PluginContext[I], error) {
	client := plugin.NewClient(
		&plugin.ClientConfig{
			HandshakeConfig: handshakeConfig,
			Plugins:         loader.pluginMap,
			Cmd:             exec.Command(path),
			Logger:          log.NewInternalLogger(logger),
		},
	)

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf(
			"unable to create the RPC client for %s: %w",
			path, err,
		)
	}

	typeName := GetTypeName[I]()
	name, ok := loader.registeredPluginTypes[typeName]
	if !ok {
		return nil, fmt.Errorf("unknown plugin type %q", name)
	}

	protocolClient, err := rpcClient.Dispense(name)
	if err != nil {
		client.Kill()
		if strings.Contains(err.Error(), "unknown plugin") {
			return nil, unknownPluginErr{Name: name}
		}
		return nil, fmt.Errorf("unable to dispense plugin %s: %w", path, err)
	}

	handle, ok := protocolClient.(I)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf(
			"protocol client %T does not provide an implementation of %s",
			protocolClient,
			GetTypeName[I](),
		)
	}

	return &PluginContext[I]{
		Path:                path,
		Name:                handle.GetName(),
		Scheme:              handle.GetAttestationScheme(),
		Version:             handle.GetVersion(),
		SupportedMediaTypes: handle.GetSupportedMediaTypes(),
		Handle:              handle,
		client:              client,
	}, nil
}
