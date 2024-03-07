// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/veraison/services/handler"
	"github.com/veraison/services/plugin"
	scheme "github.com/veraison/services/scheme/tpm-enacttrust"
)

func main() {
	handler.RegisterStoreHandler(&scheme.StoreHandler{})
	plugin.Serve()
}
