// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	corimstore "github.com/veraison/services/endorsementstore/corimstore"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/plugin"
)

func main() {
	handler.RegisterEndorsementStore(corimstore.NewStore())
	plugin.Serve()
}
