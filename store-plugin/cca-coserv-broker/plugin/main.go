// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/veraison/services/handler"
	"github.com/veraison/services/plugin"
	ccaendorsementbroker "github.com/veraison/services/store-plugin/cca-coserv-broker"
)

func main() {
	handler.RegisterEndorsementStore(ccaendorsementbroker.NewBroker())
	plugin.Serve()
}
