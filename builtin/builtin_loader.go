// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package builtin

import (
	"fmt"

	"github.com/veraison/services/log"
	"github.com/veraison/services/plugin"
	"go.uber.org/zap"
)

var defaultBuiltinLoader *BuiltinLoader

type BuiltinLoader struct {
	logger *zap.SugaredLogger

	loadedByName      map[string]plugin.IPluggable
	loadedByMediaType map[string]plugin.IPluggable

	registeredPluginTypes map[string]string
}

func NewBuiltinLoader(logger *zap.SugaredLogger) *BuiltinLoader {
	return &BuiltinLoader{logger: logger}
}

func CreateBuiltinLoader(
	cfg map[string]interface{},
	logger *zap.SugaredLogger,
) (*BuiltinLoader, error) {
	loader := NewBuiltinLoader(logger)
	err := loader.Init(cfg)
	return loader, err
}

func (o *BuiltinLoader) Init(m map[string]interface{}) error {
	o.loadedByName = map[string]plugin.IPluggable{}

	o.loadedByMediaType = make(map[string]plugin.IPluggable)
	o.registeredPluginTypes = make(map[string]string)
	return nil
}

func (o *BuiltinLoader) GetRegisteredMediaTypes() []string {
	var mediaTypes []string // nolint:prealloc

	for mt := range o.loadedByMediaType {
		mediaTypes = append(mediaTypes, mt)
	}

	return mediaTypes
}

func DiscoverBuiltin[I plugin.IPluggable]() error {
	return DiscoverBuiltinUsing[I](defaultBuiltinLoader)
}

func DiscoverBuiltinUsing[I plugin.IPluggable](loader *BuiltinLoader) error {
	for _, p := range plugins {
		_, ok := p.(I)
		if !ok {
			continue
		}

		name := p.GetName()
		if _, ok := loader.loadedByName[name]; ok {
			loader.logger.Panicw("duplicate plugin name", "name", name)
		}

		loader.logger.Debugw("found plugin", "name", name)
		loader.loadedByName[name] = p

		for _, mt := range p.GetSupportedMediaTypes() {
			if existing, ok := loader.loadedByMediaType[mt]; ok {
				loader.logger.Panicw("media type handled by two plugins",
					"media-type", mt,
					"plugin1", existing.GetName(), "plugin2", name)
			}

			loader.logger.Debugw("media type handled", "media-type", mt, "name", name)
			loader.loadedByMediaType[mt] = p
		}
	}

	return nil
}

func GetBuiltinHandleByMediaType[I plugin.IPluggable](mediaType string) (I, error) {
	return GetBuiltinHandleByMediaTypeUsing[I](defaultBuiltinLoader, mediaType)
}

func GetBuiltinHandleByMediaTypeUsing[I plugin.IPluggable](
	ldr *BuiltinLoader,
	mediaType string,
) (I, error) {
	handle, ok := ldr.loadedByMediaType[mediaType].(I)
	if !ok {
		iface := plugin.GetTypeName[I]()
		return *new(I), fmt.Errorf( // nolint:gocritic
			"implementation providing %q with interface %s not found",
			mediaType, iface)
	}

	return handle, nil
}

func GetBuiltinHandleByNameUsing[I plugin.IPluggable](ldr *BuiltinLoader, name string) (I, error) {
	handle, ok := ldr.loadedByName[name].(I)
	if !ok {
		iface := plugin.GetTypeName[I]()
		return *new(I), fmt.Errorf( // nolint:gocritic
			"plugin named %q with interface %s not found",
			name, iface)
	}

	return handle, nil
}

func GetBuiltinLoadedAttestationSchemes[I plugin.IPluggable](ldr *BuiltinLoader) []string {
	schemes := make([]string, len(ldr.loadedByName))

	i := 0
	for _, ihandle := range ldr.loadedByName {
		if _, ok := ihandle.(I); !ok {
			continue
		}

		schemes[i] = ihandle.GetAttestationScheme()
		i += 1
	}

	return schemes
}

func GetBuiltinHandleByAttestationSchemeUsing[I plugin.IPluggable](
	ldr *BuiltinLoader,
	scheme string,
) (I, error) {
	iface := plugin.GetTypeName[I]()

	var impl I
	var ok, found bool

	for name, ictx := range ldr.loadedByName {
		if ictx.GetAttestationScheme() != scheme {
			continue
		}
		ldr.logger.Debugw("found plugin implementing scheme",
			"plugin", name, "scheme", scheme)

		impl, ok = ictx.(I)
		if ok {
			found = true
			break
		}
	}

	if !found {
		return *new(I), fmt.Errorf( // nolint:gocritic
			"could not find plugin providing scheme %q and implementing interface %s",
			scheme, iface)
	}

	return impl, nil
}

func init() {
	defaultBuiltinLoader = NewBuiltinLoader(log.Named("builtin"))
}
