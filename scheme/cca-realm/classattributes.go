// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package cca_realm

import (
	"fmt"

	"github.com/veraison/corim/comid"
)

type ClassAttributes struct {
	UUID   string
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

	uuID, err := classID.GetUUID()
	if err != nil {
		return fmt.Errorf("could not extract uu-id from class-id: %w", err)
	}

	if err := uuID.Valid(); err != nil {
		return fmt.Errorf("no valid uu-id: %w", err)
	}

	o.UUID = uuID.String()

	if class.Vendor != nil {
		o.Vendor = *class.Vendor
	}

	if class.Model != nil {
		o.Model = *class.Model
	}

	return nil
}
