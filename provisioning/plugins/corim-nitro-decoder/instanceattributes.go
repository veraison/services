// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/veraison/corim/comid"
)

type NitroInstanceAttributes struct {
//	InstID eat.UEID nothing in here for now
}

func (o *NitroInstanceAttributes) FromEnvironment(e comid.Environment) error {
	// nothing to do here for now
	return nil
}