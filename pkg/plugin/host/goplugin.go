//go:build !builtin

// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package host

import (
	"github.com/veraison/services/pkg/plugin"
	"github.com/veraison/services/pkg/plugin/goplugin"
)

func RegisterImplementation[I plugin.IPluggable](name string, i I) error {
	return goplugin.RegisterImplementation(name, i)
}

func RegisterPluginType[I plugin.IPluggable](name string, ch *plugin.RPCChannel[I]) error {
	return goplugin.RegisterPluginType(name, ch)
}

func Serve() {
	goplugin.Serve()
}
