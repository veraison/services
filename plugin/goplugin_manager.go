// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package plugin

import (
	"errors"

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

func CreateGoPluginManager[I IPluggable](
	v *viper.Viper,
	logger *zap.SugaredLogger,
	name string,
	ch *RPCChannel[I],
) (*GoPluginManager[I], error) {

	subs, err := config.GetSubs(v, "go-plugin")
	if err != nil {
		return nil, err
	}

	loader, err := CreateGoPluginLoader(subs["go-plugin"].AllSettings(), logger)
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
	for _, mt := range mts {
		if mt == mediaType {
			return true
		}
	}
	return false
}

func (o *GoPluginManager[I]) GetRegisteredMediaTypes() []string {
	var registeredMediatTypes []string

	for mtName, pc := range o.loader.loadedByMediaType {
		if _, ok := pc.GetHandle().(I); ok {
			registeredMediatTypes = append(registeredMediatTypes, mtName)
		}
	}

	return registeredMediatTypes
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
