// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/plugin/test"
)

type PowerCell struct {
}

func (o PowerCell) GetName() string {
	return "power cell"
}

func (o PowerCell) GetAttestationScheme() string {
	return "star-trek"
}

func (o PowerCell) GetSupportedMediaTypes() []string {
	return []string{"plasma"}
}

func (o PowerCell) GetVersion() string {
	return "1.0.0"
}

func (o PowerCell) GetCapacity() int {
	return 12000000
}

func main() {
	test.RegisterAmmoImplementation(&PowerCell{})
	plugin.Serve()
}
