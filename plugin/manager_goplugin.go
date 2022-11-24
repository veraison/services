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
	loader *Loader
	logger *zap.SugaredLogger
}

func NewGoPluginManager[I IPluggable](
	loader *Loader,
	logger *zap.SugaredLogger,
) *GoPluginManager[I] {
	return &GoPluginManager[I]{loader: loader, logger: logger}
}

func CreateGoPluginManager[I IPluggable](
	v *viper.Viper,
	logger *zap.SugaredLogger,
	name string,
	ch RPCChannel[I],
) (*GoPluginManager[I], error) {

	subs, err := config.GetSubs(v, "go-plugin")
	if err != nil {
		return nil, err
	}

	loader, err := CreateLoader(subs["go-plugin"].AllSettings(), logger)
	if err != nil {
		return nil, err
	}

	return CreateGoPluginManagerWithLoader(loader, name, ch)
}

func CreateGoPluginManagerWithLoader[I IPluggable](
	loader *Loader,
	name string,
	ch RPCChannel[I],
) (*GoPluginManager[I], error) {
	manager := NewGoPluginManager[I](loader, logger)
	if err := manager.Init(name, ch); err != nil {
		return nil, err
	}

	return manager, nil
}

func (o *GoPluginManager[I]) Init(name string, ch RPCChannel[I]) error {
	RegisterUsing(o.loader, name, ch)
	return DiscoverUsing[I](o.loader)
}

func (o *GoPluginManager[I]) Close() error {
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
	typeName := getTypeName[I]()
	return o.loader.GetRegisteredMediaTypesByName(typeName)
}

func (o *GoPluginManager[I]) LookupByName(name string) (I, error) {
	return GetHandleByNameUsing[I](o.loader, name)
}

func (o *GoPluginManager[I]) LookupByMediaType(mediaType string) (I, error) {
	return GetHandleByMediaTypeUsing[I](o.loader, mediaType)
}
