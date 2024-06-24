// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package realm

import (
	"errors"
	"fmt"

	"github.com/veraison/corim/comid"
)

type RealmClassAttributes struct {
	UUID   *string
	Vendor *string
}

// extract class variables from environment
func (o *RealmClassAttributes) FromEnvironment(e comid.Environment) error {
	class := e.Class

	if class == nil {
		return nil
	}

	classID := class.ClassID
	if classID != nil {
		UUID, err := classID.GetUUID()
		if err != nil {
			return fmt.Errorf("could not extract uuid from class-id: %w", err)
		}

		if err := UUID.Valid(); err != nil {
			return fmt.Errorf("no valid uuid: %w", err)
		}
		uuid := UUID.String()
		o.UUID = &uuid
	} else {
		if class.Vendor != nil {
			o.Vendor = class.Vendor
		} else {
			return errors.New("class is neither a classID or a vendor name")
		}
	}

	return nil
}
