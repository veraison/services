// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/veraison/services/pkg/endorsement"
	"github.com/veraison/services/pkg/plugin/host"
	self "github.com/veraison/services/plugins/store/corim-store"
)

func main() {
	endorsement.RegisterStore(self.NewStore(), self.Capabilities...)
	host.Serve()
}
