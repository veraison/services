// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package endorsement

import (
	"github.com/veraison/services/pkg/plugin/host"
)

const StorePluginType = "endorsement-store"

func RegisterStore(i INamed, capabilities ...Capability) {
	if err := host.RegisterImplementation(
		StorePluginType,
		MustWrapHandler(i, capabilities...),
	); err != nil {
		panic(err)
	}
}

func init() {
	if err := host.RegisterPluginType(StorePluginType, EndorsementStoreRPC); err != nil {
		panic(err)
	}
}
