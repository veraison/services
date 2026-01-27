// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/veraison/services/handler"
	"github.com/veraison/services/plugin"
	scheme "github.com/veraison/services/scheme/sevsnp"
)

func main() {
	handler.RegisterSchemeImplementation(scheme.Descriptor, scheme.NewImplementation())
	plugin.Serve()
}
