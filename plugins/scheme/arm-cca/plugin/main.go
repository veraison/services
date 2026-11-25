// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/veraison/services/pkg/plugin/host"
	"github.com/veraison/services/pkg/scheme"
	self "github.com/veraison/services/plugins/scheme/arm-cca"
)

func main() {
	scheme.RegisterImplementation(self.Descriptor, self.NewImplementation())
	host.Serve()
}
