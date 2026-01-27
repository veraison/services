// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"github.com/veraison/services/plugin"
)

func RegisterCoservProxyHandler(i ICoservProxyHandler) {
	err := plugin.RegisterImplementation("coserv-proxy-handler", i, CoservProxyHandlerRPC)
	if err != nil {
		panic(err)
	}
}

func RegisterSchemeHandler(i ISchemeHandler) {
	err := plugin.RegisterImplementation("scheme-handler", i, SchemeHandlerRPC)
	if err != nil {
		panic(err)
	}
}

func RegisterSchemeImplementation(desc SchemeDescriptor, i ISchemeImplementation) {
	wrapper, err := NewSchemeImplementationWrapper(desc, i)
	if err != nil {
		panic(err)
	}

	RegisterSchemeHandler(wrapper)
}
