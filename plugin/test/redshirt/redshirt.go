// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/plugin/test"
)

type RedShirt struct {
}

func (o RedShirt) GetName() string {
	return "Federation Starship Officer"
}

func (o RedShirt) GetAttestationScheme() string {
	return "star-trek"
}

func (o RedShirt) GetSupportedMediaTypes() []string {
	return []string{"phaser"}
}

func (o RedShirt) GetVersion() string {
	return "1.0.0"
}

func (o RedShirt) Shoot() string {
	return `phaser goes "zap"`
}

func main() {
	test.RegisterMookImplementation(&RedShirt{})
	plugin.Serve()
}
