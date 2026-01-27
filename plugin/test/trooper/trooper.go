// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/plugin/test"
)

type ImperialTrooper struct {
}

func (o ImperialTrooper) GetName() string {
	return "Galactic Imperial trooper"
}

func (o ImperialTrooper) GetAttestationScheme() string {
	return "star-wars"
}

func (o ImperialTrooper) GetSupportedMediaTypes() map[string][]string {
	return map[string][]string{"mook": {"blaster"}}
}

func (o ImperialTrooper) Shoot() string {
	return `blaster goes "pew, pew"`
}

func main() {
	test.RegisterMookImplementation(&ImperialTrooper{})
	plugin.Serve()
}
