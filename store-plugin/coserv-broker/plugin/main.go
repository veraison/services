// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/veraison/services/handler"
	"github.com/veraison/services/plugin"
	coservbroker "github.com/veraison/services/store-plugin/coserv-broker"
)

func main() {
	handler.RegisterEndorsementStore(coservbroker.NewBroker())
	plugin.Serve()
}
