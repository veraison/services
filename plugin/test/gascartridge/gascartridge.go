// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/plugin/test"
)

type GasCartridge struct {
}

func (o GasCartridge) GetName() string {
	return "gas cartridge"
}

func (o GasCartridge) GetAttestationScheme() string {
	return "star-wars"
}

func (o GasCartridge) GetSupportedMediaTypes() map[string][]string {
	return map[string][]string{"ammo": {"tibanna gas"}}
}

func (o GasCartridge) GetCapacity() int {
	return 500
}

func main() {
	test.RegisterAmmoImplementation(&GasCartridge{})
	plugin.Serve()
}
