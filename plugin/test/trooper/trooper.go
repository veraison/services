// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"

	"github.com/veraison/services/plugin"
	"github.com/veraison/services/plugin/test"
)

type ImperialTrooper struct {
	sound string
}

func (o *ImperialTrooper) Init(params *plugin.Parameters) error {
	var err error

	o.sound, err = params.GetString("sound")
	return err
}

func (o ImperialTrooper) GetName() string {
	return "Galactic Imperial Trooper"
}

func (o ImperialTrooper) GetAttestationScheme() string {
	return "star-wars"
}

func (o ImperialTrooper) GetSupportedMediaTypes() map[string][]string {
	return map[string][]string{"mook": {"blaster"}}
}

func (o ImperialTrooper) Shoot() string {
	return fmt.Sprintf("blaster goes %q", o.sound)
}

func main() {
	test.RegisterMookImplementation(&ImperialTrooper{})
	plugin.Serve()
}
