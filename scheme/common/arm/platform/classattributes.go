// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package platform

import (
	"fmt"

	"github.com/veraison/corim/comid"
)

type ClassAttributes struct {
	ImplID []byte
	Vendor string
	Model  string
}

// extract mandatory ImplID and optional vendor & model
func (o *ClassAttributes) FromEnvironment(e comid.Environment) error {
	class := e.Class

	if class == nil {
		return fmt.Errorf("expecting class in environment")
	}

	classID := class.ClassID

	if classID == nil {
		return fmt.Errorf("expecting class-id in class")
	}

	implID, err := classID.GetImplID()
	if err != nil {
		return fmt.Errorf("could not extract implementation-id from class-id: %w", err)
	}

	o.ImplID = implID[:]

	if class.Vendor != nil {
		o.Vendor = *class.Vendor
	}

	if class.Model != nil {
		o.Model = *class.Model
	}

	return nil
}
