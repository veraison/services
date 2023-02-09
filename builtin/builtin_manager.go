// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package builtin

import (
	"github.com/spf13/viper"
	"github.com/veraison/services/config"
	"github.com/veraison/services/plugin"
	"go.uber.org/zap"
)

type BuiltinManager[I plugin.IPluggable] struct {
	loader *BuiltinLoader
	logger *zap.SugaredLogger
}

func NewBuiltinManager[I plugin.IPluggable](
	loader *BuiltinLoader,
	logger *zap.SugaredLogger,
) *BuiltinManager[I] {
	return &BuiltinManager[I]{loader: loader, logger: logger}
}

func CreateBuiltinManager[I plugin.IPluggable](
	v *viper.Viper,
	logger *zap.SugaredLogger,
	name string,
) (*BuiltinManager[I], error) {
	subs, err := config.GetSubs(v, "*builtin")
	if err != nil {
		return nil, err
	}

	loader, err := CreateBultinLoader(subs["builtin"].AllSettings(), logger)
	if err != nil {
		return nil, err
	}

	return CreateBuiltinManagerWithLoader[I](loader, logger, name)
}

func CreateBuiltinManagerWithLoader[I plugin.IPluggable](
	loader *BuiltinLoader,
	logger *zap.SugaredLogger,
	name string,
) (*BuiltinManager[I], error) {
	manager := NewBuiltinManager[I](loader, logger)
	if err := manager.Init(name, nil); err != nil {
		return nil, err
	}

	return manager, nil
}

func (o *BuiltinManager[I]) Init(name string, ch *plugin.RPCChannel[I]) error {
	return DiscoverBuiltinUsing[I](o.loader)
}

func (o *BuiltinManager[I]) Close() error {
	return nil
}

func (o *BuiltinManager[I]) IsRegisteredMediaType(mediaType string) bool {
	mts := o.GetRegisteredMediaTypes()
	for _, mt := range mts {
		if mt == mediaType {
			return true
		}
	}
	return false
}

func (o *BuiltinManager[I]) GetRegisteredMediaTypes() []string {
	var registeredMediatTypes []string

	for mtName, mt := range o.loader.loadedByMediaType {
		if _, ok := mt.(I); ok {
			registeredMediatTypes = append(registeredMediatTypes, mtName)
		}
	}

	return registeredMediatTypes
}

func (o *BuiltinManager[I]) LookupByName(name string) (I, error) {
	return GetBuiltinHandleByNameUsing[I](o.loader, name)
}

func (o *BuiltinManager[I]) LookupByAttestationScheme(scheme string) (I, error) {
	return GetBuiltinHandleByAttestationSchemeUsing[I](o.loader, scheme)
}

func (o *BuiltinManager[I]) LookupByMediaType(mediaType string) (I, error) {
	return GetBuiltinHandleByMediaTypeUsing[I](o.loader, mediaType)
}
