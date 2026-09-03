// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"github.com/veraison/services/plugin"
)

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

func RegisterEndorsementStore(i IEndorsementStorePlugin) {
	err := plugin.RegisterImplementation("endorsement-store", i, EndorsementStoreRPC)
	if err != nil {
		panic(err)
	}
}
