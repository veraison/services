// Copyright 2022-2023 Contributors to the Veraison project.
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

func (o GasCartridge) GetSupportedMediaTypes() []string {
	return []string{"tibanna gas"}
}

func (o GasCartridge) GetVersion() string {
	return "1.0.0"
}

func (o GasCartridge) GetCapacity() int {
	return 500
}

func main() {
	test.RegisterAmmoImplementation(&GasCartridge{})
	plugin.Serve()
}
