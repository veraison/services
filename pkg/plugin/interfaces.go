// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package plugin

// IPluggable respresents a "pluggable" point within Veraison services. It is
// the common interfaces shared by all Veraison plugins loaded through this
// framework.
type IPluggable interface {
	// Init performs plugin initialization using parameters provided by
	// configuration. This method is always invoked by the IPluginLoader
	// implementation when the plugin is loaded, and should not, in general,
	// be invoked by the plugin's user or anyone else.
	Init(params *Parameters) error

	// GetName returns a string containing the name of the
	// implementation of this IPluggable interface. It is the plugin name.
	GetName() string

	// Releases the resources used by the plugin. This method should be
	// called before closing connection to the plugin.
	Fini() error
}

type IPlugin interface {
	Source() string
	Pluggables() map[string]IPluggable
}

type ILoader interface {
	LoadFrom(string) error
	Unload(IPlugin, string) error
	Plugins() []IPlugin
	Close()
}
