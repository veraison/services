// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package plugin

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/hashicorp/go-plugin"
	"go.uber.org/zap"

	"github.com/veraison/services/config"
	"github.com/veraison/services/log"
)

var defaultLoader *Loader

type unknownPluginErr struct {
	Name string
}

func (o unknownPluginErr) Error() string {
	return fmt.Sprintf("unknown plugin: %s", o.Name)
}

type LoaderConfig struct {
	Directory string `mapstructure:"dir"`
}

type Loader struct {
	Location string

	logger                *zap.SugaredLogger
	loadedByName          map[string]IPluginContext
	loadedByMediaType     map[string]IPluginContext

	// This gets specified as Plugins when creating a new go-plugin client.
	pluginMap             map[string]plugin.Plugin

	registeredPluginTypes map[string]string
}

func NewLoader(logger *zap.SugaredLogger) *Loader {
	return &Loader{logger: logger}
}

func CreateLoader(cfg map[string]interface{}, logger *zap.SugaredLogger) (*Loader, error) {
	pm := NewLoader(logger)
	err := pm.Init(cfg)
	return pm, err
}

func (o *Loader) Init(m map[string]interface{}) error {
	o.pluginMap = make(map[string]plugin.Plugin)
	o.loadedByName = make(map[string]IPluginContext)
	o.loadedByMediaType = make(map[string]IPluginContext)
	o.registeredPluginTypes = make(map[string]string)

	var cfg LoaderConfig
	configLoader := config.NewLoader(&cfg)
	if err := configLoader.LoadFromMap(m); err != nil {
		return err
	}

	o.Location = cfg.Directory

	return nil
}

func (o *Loader) Close() {
	for _, plugin := range o.loadedByName {
		plugin.Close()
	}
}

func (o *Loader) GetRegisteredMediaTypes() []string {
	var mediaTypes []string

	for mt := range o.loadedByMediaType {
		mediaTypes = append(mediaTypes, mt)
	}

	return mediaTypes
}

func (o *Loader) GetRegisteredMediaTypesByName(name string) []string {
	var mediaTypes []string

	for mt, pc := range o.loadedByMediaType {
		if pc.GetTypeName() == name {
			mediaTypes = append(mediaTypes, mt)
		}
	}

	return mediaTypes
}

func Init(m map[string]interface{}) error {
	return defaultLoader.Init(m)
}

func Register[I IPluggable](name string, ch RPCChannel[I]) error {
	return RegisterUsing(defaultLoader, name, ch)
}

func RegisterUsing[I IPluggable](loader *Loader, name string, ch RPCChannel[I]) error {
	if _, ok := loader.pluginMap[name]; ok {
		return fmt.Errorf("plugin for %q is already registred", name)
	}

	if err := registerRPCChannel(name, ch); err != nil {
		return err
	}

	loader.pluginMap[name] = &Plugin[I]{Name: name}
	loader.registeredPluginTypes[getTypeName[I]()] = name

	return nil
}

func Discover[I IPluggable]() error {
	return DiscoverUsing[I](defaultLoader)
}

func DiscoverUsing[I IPluggable](o *Loader) error {
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

func GetHandleByMediaType[I IPluggable](mediaType string) (I, error) {
	return GetHandleByMediaTypeUsing[I](defaultLoader, mediaType)
}

func GetHandleByMediaTypeUsing[I IPluggable](ldr *Loader, mediaType string) (I, error) {
	plugged, ok := ldr.loadedByMediaType[mediaType].(*PluginContext[I])
	if !ok {
		iface := getTypeName[I]()
		return *new(I), fmt.Errorf("plugin providing %q with interface %s not found", mediaType, iface)
	}

	return plugged.Handle, nil
}

func GetHandleByNameUsing[I IPluggable](ldr *Loader, name string) (I, error) {
	plugged, ok := ldr.loadedByName[name].(*PluginContext[I])
	if !ok {
		iface := getTypeName[I]()
		return *new(I), fmt.Errorf("plugin named %q with interface %s not found", name, iface)
	}

	return plugged.Handle, nil
}

func getTypeName[I IPluggable]() string {
	return reflect.TypeOf((*I)(nil)).Elem().Name()
}

func init() {
	defaultLoader = NewLoader(log.Named("plugin"))
}
