// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"

	"github.com/veraison/corim/comid"
)

type NitroClassAttributes struct {
	//ImplID []byte
	Vendor string
	Model  string
}

// extract mandatory ImplID and optional vendor & model
func (o *NitroClassAttributes) FromEnvironment(e comid.Environment) error {
	class := e.Class

	if class == nil {
		return fmt.Errorf("expecting class in environment")
	}

	classID := class.ClassID

	if classID == nil {
		return fmt.Errorf("expecting class-id in class")
	}

	if class.Vendor != nil {
		o.Vendor = *class.Vendor
	}

	if class.Model != nil {
		o.Model = *class.Model
	}

	return nil
}