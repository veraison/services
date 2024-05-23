// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package arm

import (
	"errors"

	"github.com/veraison/corim/comid"
)

type RealmInstanceAttributes struct {
	InstID string
}

func (o *RealmInstanceAttributes) FromEnvironment(e comid.Environment) error {
	var err error

	if e.Instance == nil {
		return errors.New("expecting instance in environment")
	}

	if e.Instance.Type() != "bytes" {
		return errors.New("expecting instance as bytes for CCA Realm")
	}
	b := e.Instance.Bytes()

	o.InstID = string(b)
	return err
}
