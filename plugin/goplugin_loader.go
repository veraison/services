// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package plugin

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-plugin"
	"go.uber.org/zap"

	"github.com/veraison/services/config"
	"github.com/veraison/services/log"
)

var defaultGoPluginLoader *GoPluginLoader

type unknownPluginErr struct {
	Name string
}

func (o unknownPluginErr) Error() string {
	return fmt.Sprintf("unknown plugin: %s", o.Name)
}

type GoPluginLoaderConfig struct {
	Directory string `mapstructure:"dir"`
}

type GoPluginLoader struct {
	Location string

	logger            *zap.SugaredLogger
	loadedByName      map[string]IPluginContext
	loadedByMediaType map[string]IPluginContext

	// This gets specified as Plugins when creating a new go-plugin client.
	pluginMap map[string]plugin.Plugin

	registeredPluginTypes map[string]string
}

func NewGoPluginLoader(logger *zap.SugaredLogger) *GoPluginLoader {
	return &GoPluginLoader{logger: logger}
}

func CreateGoPluginLoader(
	cfg map[string]interface{},
	logger *zap.SugaredLogger,
) (*GoPluginLoader, error) {
	loader := NewGoPluginLoader(logger)
	err := loader.Init(cfg)
	return loader, err
}

func (o *GoPluginLoader) Init(m map[string]interface{}) error {
	o.pluginMap = make(map[string]plugin.Plugin)
	o.loadedByName = make(map[string]IPluginContext)
	o.loadedByMediaType = make(map[string]IPluginContext)
	o.registeredPluginTypes = make(map[string]string)

	var cfg GoPluginLoaderConfig
	configLoader := config.NewLoader(&cfg)
	if err := configLoader.LoadFromMap(m); err != nil {
		return err
	}

	o.Location = cfg.Directory

	return nil
}

func (o *GoPluginLoader) Close() {
	for _, plugin := range o.loadedByName {
		plugin.Close()
	}
}

func (o *GoPluginLoader) GetRegisteredMediaTypes() []string {
	var mediaTypes []string // nolint:prealloc

	for mt := range o.loadedByMediaType {
		mediaTypes = append(mediaTypes, mt)
	}

	return mediaTypes
}

func (o *GoPluginLoader) GetRegisteredMediaTypesByPluginType(typeName string) []string {
	var mediaTypes []string

	for mt, pc := range o.loadedByMediaType {
		if pc.GetTypeName() == typeName {
			mediaTypes = append(mediaTypes, mt)
		}
	}

	return mediaTypes
}

func Init(m map[string]interface{}) error {
	return defaultGoPluginLoader.Init(m)
}

func RegisterGoPlugin[I IPluggable](name string, ch *RPCChannel[I]) error {
	return RegisterGoPluginUsing(defaultGoPluginLoader, name, ch)
}

func RegisterGoPluginUsing[I IPluggable](
	loader *GoPluginLoader,
	name string,
	ch *RPCChannel[I],
) error {
	if _, ok := loader.pluginMap[name]; ok {
		return fmt.Errorf("plugin for %q is already registred", name)
	}

	if err := registerRPCChannel(name, ch); err != nil {
		return err
	}

	loader.pluginMap[name] = &Plugin[I]{Name: name}
	loader.registeredPluginTypes[GetTypeName[I]()] = name

	return nil
}

func DiscoverGoPlugin[I IPluggable]() error {
	return DiscoverGoPluginUsing[I](defaultGoPluginLoader)
}

func DiscoverGoPluginUsing[I IPluggable](o *GoPluginLoader) error {
	if o.Location == "" {
		return errors.New("plugin manager has not been initialized")
	}

	o.logger.Debugw("discovering plugins", "location", o.Location)
	pluginPaths, err := plugin.Discover("*.plugin", o.Location)
	if err != nil {
		return err
	}

	for _, path := range pluginPaths {
		pluginContext, err := createPluginContext[I](o, path, o.logger)
		if err != nil {
			var upErr unknownPluginErr
			if errors.As(err, &upErr) {
				o.logger.Debugw("plugin not found", "name", upErr.Name, "path", path)
				continue
			} else {
				return err
			}
		}

		pluginName := pluginContext.GetName()
		if existing, ok := o.loadedByName[pluginName]; ok {
			return fmt.Errorf(
				"plugin %q provided by two sources: [%s] and [%s]",
				pluginName,
				existing.GetPath(),
				pluginContext.GetPath(),
			)
		}
		o.loadedByName[pluginName] = pluginContext

		for _, mediaType := range pluginContext.SupportedMediaTypes {
			if existing, ok := o.loadedByMediaType[mediaType]; ok {
				return fmt.Errorf(
					"plugins %q [%s] and %q [%s] both provides support for %q",
					existing.GetName(),
					existing.GetPath(),
					pluginContext.GetName(),
					pluginContext.GetPath(),
					mediaType,
				)
			}
			o.loadedByMediaType[mediaType] = pluginContext
		}
	}

	return nil
}

func GetGoPluginHandleByMediaType[I IPluggable](mediaType string) (I, error) {
	return GetGoPluginHandleByMediaTypeUsing[I](defaultGoPluginLoader, mediaType)
}

func GetGoPluginHandleByMediaTypeUsing[I IPluggable](
	ldr *GoPluginLoader,
	mediaType string,
) (I, error) {
	plugged, ok := ldr.loadedByMediaType[mediaType].(*PluginContext[I])
	if !ok {
		iface := GetTypeName[I]()
		return *new(I), fmt.Errorf( // nolint:gocritic
			"plugin providing %q with interface %s not found",
			mediaType, iface)
	}

	return plugged.Handle, nil
}

func GetGoPluginHandleByNameUsing[I IPluggable](ldr *GoPluginLoader, name string) (I, error) {
	plugged, ok := ldr.loadedByName[name].(*PluginContext[I])
	if !ok {
		iface := GetTypeName[I]()
		return *new(I), fmt.Errorf( // nolint:gocritic
			"plugin named %q with interface %s not found",
			name, iface)
	}

	return plugged.Handle, nil
}

func GetGoPluginLoadedAttestationSchemes[I IPluggable](ldr *GoPluginLoader) []string {
	schemes := make([]string, len(ldr.loadedByName))

	i := 0
	for _, ictx := range ldr.loadedByName {
		if _, ok := ictx.(*PluginContext[I]); !ok {
			continue
		}

		schemes[i] = ictx.GetAttestationScheme()
		i += 1
	}

	return schemes
}

func GetGoPluginHandleByAttestationSchemeUsing[I IPluggable](
	ldr *GoPluginLoader,
	scheme string,
) (I, error) {
	iface := GetTypeName[I]()

	var ctx *PluginContext[I]
	var ok bool

	for name, ictx := range ldr.loadedByName {
		if ictx.GetAttestationScheme() != scheme {
			continue
		}
		ldr.logger.Debugw("found plugin implementing scheme",
			"plugin", name, "scheme", scheme)

		ctx, ok = ictx.(*PluginContext[I])
		if ok {
			break
		}
	}

	if ctx == nil {
		return *new(I), fmt.Errorf( // nolint:gocritic
			"could not find plugin providing scheme %q and implementing interface %s",
			scheme, iface)
	}

	return ctx.Handle, nil
}

func init() {
	defaultGoPluginLoader = NewGoPluginLoader(log.Named("plugin"))
}
