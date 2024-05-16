// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package cca_realm_provisioning

import (
	"fmt"
	"strings"

	"github.com/veraison/corim/comid"
)

type RealmAttributes struct {
	Rim       []byte
	Rem       [4][]byte
	HashAlgID string
}

func (o *RealmAttributes) FromMeasurement(m comid.Measurement) error {
	if err := o.extractRegisterIndexes(m.Val.IntegrityRegisters); err != nil {
		return fmt.Errorf("extracting measurement: %w", err)
	}

	return nil
}

func (o *RealmAttributes) extractRealmDigest(digests comid.Digests) (algID string, hash []byte, err error) {
	if err := digests.Valid(); err != nil {
		return "", nil, fmt.Errorf("invalid digest: %v", err)
	}
	if len(digests) != 1 {
		return "", nil, fmt.Errorf("expecting 1 digest, got %d", len(digests))
	}

	return digests[0].AlgIDToString(), digests[0].HashValue, nil
}

func (o *RealmAttributes) extractRegisterIndexes(r *comid.IntegrityRegisters) error {
	for k, val := range r.M {
		a, d, err := o.extractRealmDigest(val)
		if err != nil {
			return fmt.Errorf("unable to extract realm digest: %v", err)
		}
		switch t := k.(type) {
		case string:
			key := strings.ToLower(t)
			if !o.isCompatibleAlgID(a) {
				return fmt.Errorf("incompatible AlgID %s for key %s", a, key)
			}
			switch key {
			case "rim":
				o.HashAlgID = a
				o.Rim = d
			case "rem0":
				o.Rem[0] = d
			case "rem1":
				o.Rem[1] = d
			case "rem2":
				o.Rem[2] = d
			case "rem3":
				o.Rem[3] = d
			default:
				return fmt.Errorf("unexpected register index: %s", key)
			}
		default:
			return fmt.Errorf("unexpected type for index: %T", t)
		}
	}
	return nil
}

func (o RealmAttributes) isCompatibleAlgID(hashAlgID string) bool {
	return o.HashAlgID == "" || hashAlgID == o.HashAlgID
}
