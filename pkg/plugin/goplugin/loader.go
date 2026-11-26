//go:build !builtin

// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package goplugin

import (
	"fmt"
	"maps"
	"os/exec"
	"slices"
	"strings"

	goplugin "github.com/hashicorp/go-plugin"
	"github.com/veraison/services/pkg/log"
	"github.com/veraison/services/pkg/plugin"
	"go.uber.org/zap"
)

var (
	pluginRegister = make(goplugin.PluginSet)
	nameRegister   = make(map[string]string)
)

func RegisterPluginType[I plugin.IPluggable](name string, ch *plugin.RPCChannel[I]) error {
	if _, ok := pluginRegister[name]; ok {
		return fmt.Errorf("plugin type %q already registered", name)
	}

	if err := registerRPCChannel(name, ch); err != nil {
		return err
	}

	pluginRegister[name] = &GoPlugin[I]{Name: name}
	nameRegister[plugin.GetTypeName[I]()] = name

	return nil
}

func NewPluginLoader(logger *zap.SugaredLogger) plugin.ILoader {
	return &GoPluginLoader{logger: logger}
}

type GoPluginLoader struct {
	logger  *zap.SugaredLogger
	plugins []goPluginContext
}

func (o *GoPluginLoader) Plugins() []plugin.IPlugin {
	ret := make([]plugin.IPlugin, len(o.plugins))
	for i, plugin := range o.plugins {
		ret[i] = &plugin
	}

	return ret
}

func (o *GoPluginLoader) Close() {
	for _, plugin := range o.plugins {
		o.logger.Debugf("closing plugin %s", plugin.path)
		plugin.client.Kill()
	}
}

func (o *GoPluginLoader) LoadFrom(location string) error {
	o.logger.Debugf("loading plugins from %s", location)
	pluginPaths, err := goplugin.Discover("*.plugin", location)
	if err != nil {
		return err
	}

	for _, path := range pluginPaths {
		if err := o.loadPlugin(path); err != nil {
			o.Close()
			return err
		}

	}

	return nil
}

func (o *GoPluginLoader) Unload(p plugin.IPlugin, typeName string) error {
	ctx, ok := p.(*goPluginContext)
	if !ok {
		return fmt.Errorf("incorrect IPlugin type %T", p)
	}

	if _, ok := ctx.pluggables[typeName]; !ok {
		return fmt.Errorf("%s does not implement %s", ctx.Source(), typeName)
	}

	o.logger.Debugf("unloading %s from %s", typeName, ctx.Source())
	delete(ctx.pluggables, typeName)

	if len(ctx.pluggables) == 0 {
		o.logger.Debugf("all pluggables unloaded from %s; kill the plugin process", ctx.Source())
		ctx.client.Kill()
		o.plugins = slices.DeleteFunc(o.plugins, func(loaded goPluginContext) bool {
			return loaded.Source() == p.Source()
		})
	}

	return nil
}

func (o *GoPluginLoader) loadPlugin(path string) error {
	client := goplugin.NewClient(
		&goplugin.ClientConfig{
			HandshakeConfig: handshakeConfig,
			Plugins:         pluginRegister,
			Cmd:             exec.Command(path),
			Logger:          log.NewInternalLogger(o.logger.Named("rpc")),
		},
	)

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return fmt.Errorf(
			"unable to create the RPC client for %s: %w",
			path, err,
		)
	}

	pluggables := make(map[string]plugin.IPluggable, len(pluginRegister))
	for pluginType := range pluginRegister {
		handle, err := rpcClient.Dispense(pluginType)
		if err != nil {
			if strings.Contains(err.Error(), "unknown plugin") {
				continue
			}

			client.Kill()
			return fmt.Errorf("unable to dispense plugin %s: %w", path, err)
		}

		pluggable, ok := handle.(plugin.IPluggable)
		if !ok {
			client.Kill()
			return fmt.Errorf(
				"protocol client %T does not provide an IPluggable interface",
				handle,
			)
		}

		pluggables[pluginType] = pluggable
	}

	if len(pluggables) == 0 {
		o.logger.Debugf("nothing loaded from %s", path)
		return nil
	}

	pluginContext := goPluginContext{
		path:       path,
		client:     client,
		pluggables: pluggables,
	}

	o.logger.Debugf("loaded %s from %s", strings.Join(pluginContext.PluggableTypes(), ", "), path)
	o.plugins = append(o.plugins, pluginContext)

	return nil
}

type goPluginContext struct {
	path       string
	client     *goplugin.Client
	pluggables map[string]plugin.IPluggable
}

func (o *goPluginContext) PluggableTypes() []string {
	return slices.Collect(maps.Keys(o.pluggables))
}

func (o *goPluginContext) Source() string {
	return o.path
}

func (o *goPluginContext) Pluggables() map[string]plugin.IPluggable {
	return o.pluggables
}
