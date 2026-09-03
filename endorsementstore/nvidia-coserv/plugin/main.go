// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	scheme "github.com/veraison/services/endorsementstore/nvidia-coserv"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/plugin"
)

func main() {
	handler.RegisterEndorsementStore(&scheme.CoservProxyHandler{})
	plugin.Serve()
}
