// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package builtin

import (
	"fmt"
	"maps"
	"slices"

	"github.com/veraison/services/pkg/plugin"
	"go.uber.org/zap"
)

var (
	// builtinPluggables is populated at init() time by packages that
	// register their implementations (see pkg/plugin/builtin/all). This
	// package must NOT import any plugin implementations (or pkg/scheme,
	// pkg/endorsement) itself, as those depend on pkg/plugin/host, which
	// depends on this package -- doing so creates an import cycle.
	builtinPluggables = make(map[string][]plugin.IPluggable)
)

// Serve is a no-op for builtin plugins: there is no separate plugin process.
func Serve() {}

func RegisterPluginType[I plugin.IPluggable](name string, ch *plugin.RPCChannel[I]) error {
	return nil
}

// RegisterImplementation adds an implementation of the specified plugin type
// to the builtin registry. The RPC channel is unused for builtin plugins.
func RegisterImplementation[I plugin.IPluggable](name string, i I) error {
	return Register(name, i)
}

// Register adds a pluggable of the specified plugin type to the builtin
// registry. Must be called before NewPluginLoader (i.e. from an init()).
func Register(pluginType string, p plugin.IPluggable) error {
	if p == nil {
		return fmt.Errorf("nil %s pluggable", pluginType)
	}

	builtinPluggables[pluginType] = append(builtinPluggables[pluginType], p)
	return nil
}

// MustRegister is like Register but panics on error.
func MustRegister(pluginType string, p plugin.IPluggable) {
	if err := Register(pluginType, p); err != nil {
		panic(err)
	}
}

func NewPluginLoader(logger *zap.SugaredLogger) plugin.ILoader {
	logger.Debug("loading built-in plugins")
	var plugins []*builtinPluginContext
	for pluginType, pluggables := range builtinPluggables {
		for _, pluggable := range pluggables {
			plugins = append(plugins, &builtinPluginContext{
				pluggables: map[string]plugin.IPluggable{
					pluginType: pluggable,
				},
			})
			logger.Debugf("loaded %s %s", pluginType, pluggable.GetName())
		}
	}

	return &BuiltinPluginLoader{logger: logger, plugins: plugins}
}

type BuiltinPluginLoader struct {
	logger  *zap.SugaredLogger
	plugins []*builtinPluginContext
}

func (o *BuiltinPluginLoader) Plugins() []plugin.IPlugin {
	ret := make([]plugin.IPlugin, len(o.plugins))
	for i, plugin := range o.plugins {
		ret[i] = plugin
	}

	return ret
}

func (o *BuiltinPluginLoader) Close() {
}

func (o *BuiltinPluginLoader) LoadFrom(location string) error {
	o.logger.Debugf("ignoring location %s", location)
	return nil
}

func (o *BuiltinPluginLoader) Unload(p plugin.IPlugin, typeName string) error {
	impl, ok := p.Pluggables()[typeName]
	if !ok {
		return fmt.Errorf("%s does not implement %s", p.Source(), typeName)
	}

	o.logger.Debugf("unloading %s %q", typeName, impl.GetName())
	o.plugins = slices.DeleteFunc(o.plugins, func(e *builtinPluginContext) bool {
		return plugin.IPlugin(e) == p
	})

	return nil
}

type builtinPluginContext struct {
	pluggables map[string]plugin.IPluggable
}

func (o *builtinPluginContext) PluggableTypes() []string {
	return slices.Collect(maps.Keys(o.pluggables))
}

func (o *builtinPluginContext) Source() string {
	return "<builtin>"
}

func (o *builtinPluginContext) Pluggables() map[string]plugin.IPluggable {
	return o.pluggables
}
