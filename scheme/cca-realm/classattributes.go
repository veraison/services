// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package cca_realm

import (
	"errors"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/services/log"
)

type ClassAttributes struct {
	UUID   string
	Vendor string
}

// extract class variables from environment
func (o *ClassAttributes) FromEnvironment(e comid.Environment) error {
	class := e.Class

	if class == nil {
		log.Debug("no class in the environment")
		return nil
	}

	classID := class.ClassID

	if classID == nil {
		log.Debug("no classID in the environment")
	}
	if classID != nil {
		uuID, err := classID.GetUUID()
		if err != nil {
			return fmt.Errorf("could not extract uu-id from class-id: %w", err)
		}

		if err := uuID.Valid(); err != nil {
			return fmt.Errorf("no valid uu-id: %w", err)
		}

		o.UUID = uuID.String()
	}

	if class.Vendor != nil {
		o.Vendor = *class.Vendor
	} else {
		return errors.New("class is neither UUID or Vendor Name")
	}

	return nil
}
