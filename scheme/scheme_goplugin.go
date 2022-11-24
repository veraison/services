// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package scheme

import (
	"github.com/veraison/services/plugin"
)

func RegisterImplementation(i IScheme) {
	plugin.RegisterImplementation("scheme", i, SchemeRPC)
}
