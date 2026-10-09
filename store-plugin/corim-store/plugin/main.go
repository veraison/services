// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/veraison/services/handler"
	"github.com/veraison/services/plugin"
	corim_store "github.com/veraison/services/store-plugin/corim-store"
)

func main() {
	handler.RegisterEndorsementStore(corim_store.NewStore())
	plugin.Serve()
}
