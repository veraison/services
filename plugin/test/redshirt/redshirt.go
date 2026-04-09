// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"

	"github.com/veraison/services/plugin"
	"github.com/veraison/services/plugin/test"
)

type RedShirt struct {
	sound string
}

func (o *RedShirt) Init(params *plugin.Parameters) error {
	var err error
	o.sound, err = params.GetString("sound")
	return err
}

func (o RedShirt) GetName() string {
	return "Federation Starship Officer"
}

func (o RedShirt) GetAttestationScheme() string {
	return "star-trek"
}

func (o RedShirt) GetSupportedMediaTypes() map[string][]string {
	return map[string][]string{"mook": {"phaser"}}
}

func (o RedShirt) Shoot() string {
	return fmt.Sprintf("phaser goes %q", o.sound)
}

func main() {
	test.RegisterMookImplementation(&RedShirt{})
	plugin.Serve()
}
