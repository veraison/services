// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package realm

import (
	"errors"
	"fmt"
	"strings"

	"github.com/veraison/corim/comid"
	"github.com/veraison/services/log"
)

type RealmAttributes struct {
	RIM       *[]byte
	REM       [4]*[]byte
	HashAlgID string
	RPV       *[]byte
}

func (o *RealmAttributes) FromMeasurement(m comid.Measurement) error {
	if err := o.extractRealmPersonalizationValue(m.Val.RawValue); err != nil {
		return fmt.Errorf("extracting rpv: %w", err)
	}
	if err := o.extractRegisterIndexes(m.Val.IntegrityRegisters); err != nil {
		return fmt.Errorf("extracting measurement: %w", err)
	}

	if err := o.Valid(); err != nil {
		return fmt.Errorf("extracting realm attributes: %w", err)
	}
	return nil
}

func (o *RealmAttributes) GetRefValType() string {
	return "realm.reference-value"
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
	for k, val := range r.IndexMap {
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
				o.RIM = &d
			case "rem0":
				o.REM[0] = &d
			case "rem1":
				o.REM[1] = &d
			case "rem2":
				o.REM[2] = &d
			case "rem3":
				o.REM[3] = &d
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

func (o *RealmAttributes) extractRealmPersonalizationValue(r *comid.RawValue) error {
	var err error
	if r == nil {
		log.Debug("realm personalization value not present")
		return nil
	}
	rpv, err := r.GetBytes()
	if err != nil {
		return err
	} else if len(rpv) != 64 {
		return fmt.Errorf("invalid length %d for realm personalization value", len(rpv))
	}
	o.RPV = &rpv
	return nil
}

func (o *RealmAttributes) Valid() error {
	if o.RIM == nil {
		return errors.New("no realm initial measurements")
	}
	return nil
}
