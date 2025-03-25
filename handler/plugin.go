// Copyright 2022-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"github.com/veraison/services/plugin"
)

func RegisterEndorsementHandler(i IEndorsementHandler) {
	err := plugin.RegisterImplementation("endorsement-handler", i, EndorsementHandlerRPC)
	if err != nil {
		panic(err)
	}
}

func RegisterEvidenceHandler(i IEvidenceHandler) {
	err := plugin.RegisterImplementation("evidence-handler", i, EvidenceHandlerRPC)
	if err != nil {
		panic(err)
	}
}

func RegisterStoreHandler(i IStoreHandler) {
	err := plugin.RegisterImplementation("store-handler", i, StoreHandlerRPC)
	if err != nil {
		panic(err)
	}
}

func RegisterCoservProxyHandler(i ICoservProxyHandler) {
	err := plugin.RegisterImplementation("coserv-proxy-handler", i, CoservProxyHandlerRPC)
	if err != nil {
		panic(err)
	}
}
