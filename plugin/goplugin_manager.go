// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package plugin

import (
	"errors"
	"slices"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/veraison/services/config"
)

var ErrNotFound = errors.New("plugin not found")

type GoPluginManager[I IPluggable] struct {
	loader *GoPluginLoader
	logger *zap.SugaredLogger
}

func NewGoPluginManager[I IPluggable](
	loader *GoPluginLoader,
	logger *zap.SugaredLogger,
) *GoPluginManager[I] {
	return &GoPluginManager[I]{loader: loader, logger: logger}
}

// CreateGoPluginManager create a new GoPluginManager for the provided plugin
// RPCChannel ch. In addition to the RPCChannel, it takes the the following
// additional inputs to find and load the plugins:
//
//   - v - plugin configuration (section plugin in the config)
//   - pluginClass - the class of plugins for which to create the manager.
//     The parameters for loading this class of plugins will be part of
//     'pluginClass' section of the plugin configuration
//   - pluginParams - parameters to be passed to this class of plugins
//   - logger - logger that will be used by the plugin manager
//   - name - the plugin implementation name that was registered
//   - ch - the plugin RPC channel
func CreateGoPluginManager[I IPluggable](
	v *viper.Viper,
	pluginClass string,
	pluginParams map[string]*Parameters,
	logger *zap.SugaredLogger,
	name string,
	ch *RPCChannel[I],
) (*GoPluginManager[I], error) {

	subs, err := config.GetSubs(v, pluginClass)
	if err != nil {
		return nil, err
	}

	loader, err := CreateGoPluginLoader(subs[pluginClass].AllSettings(), pluginParams, logger)
	if err != nil {
		return nil, err
	}

	return CreateGoPluginManagerWithLoader(loader, name, logger, ch)
}

func CreateGoPluginManagerWithLoader[I IPluggable](
	loader *GoPluginLoader,
	name string,
	logger *zap.SugaredLogger,
	ch *RPCChannel[I],
) (*GoPluginManager[I], error) {
	manager := NewGoPluginManager[I](loader, logger)
	if err := manager.Init(name, ch); err != nil {
		return nil, err
	}

	return manager, nil
}

func (o *GoPluginManager[I]) Init(name string, ch *RPCChannel[I]) error {
	err := RegisterGoPluginUsing(o.loader, name, ch)
	if err != nil {
		return err
	}
	return DiscoverGoPluginUsing[I](o.loader)
}

func (o *GoPluginManager[I]) Close() error {
	o.loader.Close()
	return nil
}

func (o *GoPluginManager[I]) IsRegisteredMediaType(mediaType string) bool {
	mts := o.GetRegisteredMediaTypes()
	return slices.Contains(mts, mediaType)
}

func (o *GoPluginManager[I]) GetRegisteredMediaTypes() []string {
	var registeredMediaTypes []string

	for mtName, pc := range o.loader.loadedByMediaType {
		if _, ok := pc.GetHandle().(I); ok {
			registeredMediaTypes = append(registeredMediaTypes, mtName)
		}
	}

	return registeredMediaTypes
}

func (o *GoPluginManager[I]) GetRegisteredMediaTypesByCategory(category string) []string {
	var registeredMediaTypes []string

	for _, pc := range o.loader.loadedByName {
		if pluggable, ok := pc.GetHandle().(I); ok {
			mts, ok := pluggable.GetSupportedMediaTypes()[category]
			if ok {
				registeredMediaTypes = append(registeredMediaTypes, mts...)
			}
		}
	}

	return registeredMediaTypes
}

func (o *GoPluginManager[I]) GetRegisteredAttestationSchemes() []string {
	return GetGoPluginLoadedAttestationSchemes[I](o.loader)
}

func (o *GoPluginManager[I]) LookupByName(name string) (I, error) {
	return GetGoPluginHandleByNameUsing[I](o.loader, name)
}

func (o *GoPluginManager[I]) LookupByAttestationScheme(name string) (I, error) {
	return GetGoPluginHandleByAttestationSchemeUsing[I](o.loader, name)
}

func (o *GoPluginManager[I]) LookupByMediaType(mediaType string) (I, error) {
	return GetGoPluginHandleByMediaTypeUsing[I](o.loader, mediaType)
}
