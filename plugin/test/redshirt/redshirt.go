// Copyright 2022-2026 Contributors to the Veraison project.
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

func (o RedShirt) GetSupportedMediaTypes() map[string][]string {
	return map[string][]string{"mook": {"phaser"}}
}

func (o RedShirt) Shoot() string {
	return `phaser goes "zap"`
}

func main() {
	test.RegisterMookImplementation(&RedShirt{})
	plugin.Serve()
}
