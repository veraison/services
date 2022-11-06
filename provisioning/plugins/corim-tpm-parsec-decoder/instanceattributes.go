// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"encoding/base64"
	"fmt"

	"github.com/veraison/corim/comid"
)

type InstanceAttributes struct {
	KeyID string
}

func (o *InstanceAttributes) FromEnvironment(e comid.Environment) error {
	inst := e.Instance

	if inst == nil {
		return fmt.Errorf("expecting instance in environment")
	}

	keyID, err := e.Instance.GetUEID()
	if err != nil {
		return fmt.Errorf("could not extract (UEID) from instance-id")
	}

	o.KeyID = base64.RawURLEncoding.EncodeToString(keyID)

	return nil
}
