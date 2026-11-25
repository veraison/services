// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package scheme

import (
	"github.com/veraison/services/pkg/plugin/host"
)

const PluginType = "scheme"

func RegisterHandler(i IHandler) {
	err := host.RegisterImplementation(PluginType, i)
	if err != nil {
		panic(err)
	}
}

func RegisterImplementation(desc Descriptor, i Implementation) {
	wrapper, err := NewSchemeImplementationWrapper(desc, i)
	if err != nil {
		panic(err)
	}

	RegisterHandler(wrapper)
}

func init() {
	if err := host.RegisterPluginType(PluginType, SchemeHandlerRPC); err != nil {
		panic(err)
	}
}
