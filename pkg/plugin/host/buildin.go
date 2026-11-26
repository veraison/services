//go:build builtin

// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package host

import (
	"github.com/veraison/services/pkg/plugin"
	"github.com/veraison/services/pkg/plugin/builtin"
)

func RegisterImplementation[I plugin.IPluggable](name string, i I) error {
	return builtin.RegisterImplementation(name, i)
}

func RegisterPluginType[I plugin.IPluggable](name string, ch *plugin.RPCChannel[I]) error {
	return builtin.RegisterPluginType(name, ch)
}

func Serve() {
	builtin.Serve()
}
