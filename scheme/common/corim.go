// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"crypto"
	"errors"
	"fmt"
	"iter"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
)

/// CorimTestCase defines a test case for checking whether the specified CoRIM
// loads as expected.
type CorimTestCase struct {
	Title string
	Input []byte
	Err   string
}

// RunCorimTests run the provided CorimTestCase's.
func RunCorimTests(t *testing.T, tcs []CorimTestCase) {
	for _, tc := range tcs {
		t.Run(tc.Title, func(t *testing.T) {
			_, err := corim.UnmarshalAndValidateUnsignedCorimFromCBOR(tc.Input)
			if tc.Err == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.Err)
			}
		})
	}
}

// ExtractOneVerifKey returns the single CryptoKey contained within provided trust
// anchors. If there is more than one trust anchor supplied, or if it contains
// more than one VerifKey, an error is returned instead.
func ExtractOneVerifKey(trustAnchors []*comid.KeyTriple) (*comid.CryptoKey, error) {
	numTAs := len(trustAnchors) 
	if numTAs != 1 {
		return nil, fmt.Errorf("expected exactly 1 trust anchor; found %d", numTAs)
	}

	numKeys := len(trustAnchors[0].VerifKeys)
	if numKeys != 1 {
		return nil, fmt.Errorf("expected exactly 1 verif. key in triple; found %d", numKeys)
	}

	return trustAnchors[0].VerifKeys[0], nil
}

// ExtractPublicKeyFromTrustAnchors extracts the PublicKey in the common case
// where there is exactly one KeyTriple in the trust anchors, and exactly one
// VerifKey inside the triple.
func ExtractPublicKeyFromTrustAnchors(trustAnchors []*comid.KeyTriple) (crypto.PublicKey, error) {
	vk, err := ExtractOneVerifKey(trustAnchors)
	if err != nil {
		return nil, err
	}

	return vk.PublicKey()
}


type IEnvironmentValidator func(*comid.Environment) error
type ICryptoKeysValid func(keys []*comid.CryptoKey) error
type IMeasurementsValidator func(measurements []comid.Measurement) error

// TriplesValidator  implements generic logic for validating comid.Triples
// based on provided callbacks. This can be registered as a comid.ExtTriples
// extensions for a profile, so that it is automatically invoked when a CoRIM
// with that profile is loaded.
type TriplesValidator struct {
	EnviromentValidator       IEnvironmentValidator `cbor:"-" json:"-"`
	RefValEnviromentValidator IEnvironmentValidator `cbor:"-" json:"-"`
	TAEnviromentValidator     IEnvironmentValidator `cbor:"-" json:"-"`

	CryptoKeysValidator   ICryptoKeysValid       `cbor:"-" json:"-"`
	MeasurementsValidator IMeasurementsValidator `cbor:"-" json:"-"`

	DisallowTAs      bool `cbor:"-" json:"-"`
	DisallowRefVals  bool `cbor:"-" json:"-"`
}

func (o *TriplesValidator) ValidTriples(triples *comid.Triples) error {
	taSeen := false
	for i, kt := range Enumerate(triples.IterAttestVerifKeys()) {
		taSeen = true

		if o.DisallowTAs {
			return errors.New("found trust anchors (disallowed by scheme)")
		}

		if o.TAEnviromentValidator != nil {
			if err := o.TAEnviromentValidator(&kt.Environment); err != nil {
				return fmt.Errorf("trust anchor %d environment: %w", i, err)
			}
		} else if o.EnviromentValidator != nil {
			if err := o.EnviromentValidator(&kt.Environment); err != nil {
				return fmt.Errorf("trust anchor %d environment: %w", i, err)
			}
		}

		if o.CryptoKeysValidator != nil {
			if err := o.CryptoKeysValidator(kt.VerifKeys); err != nil {
				return fmt.Errorf("trust anchor %d verif. keys: %w", i, err)
			}
		}
	}

	refValSeen := false
	for i, vt := range Enumerate(triples.IterRefVals()) {
		refValSeen = true

		if o.DisallowRefVals {
			return errors.New("found reference values (disallowed by scheme)")
		}

		if o.RefValEnviromentValidator != nil {
			if err := o.RefValEnviromentValidator(&vt.Environment); err != nil {
				return fmt.Errorf("reference value %d environment: %w", i, err)
			}
		} else if o.EnviromentValidator != nil {
			if err := o.EnviromentValidator(&vt.Environment); err != nil {
				return fmt.Errorf("reference value %d environment: %w", i, err)
			}
		}

		if o.MeasurementsValidator != nil {
			if err := o.MeasurementsValidator(vt.Measurements.Values); err != nil {
				return fmt.Errorf("trust anchor %d measurements: %w", i, err)
			}
		}
	}

	if !taSeen && !refValSeen {
		return errors.New("no reference values or trust anchors in CoMID triples")
	}

	return nil
}

// Enumerate transforms an iter.Seq[T] into iter.Seq2[int, T], with the
// first value in Seq being the ordinal (starting at 0) of the associated
// T within the sequence.
// (side note: this really ought to be part of iter package...)
func Enumerate[T any](seq iter.Seq[T]) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		i := 0
		for v := range seq {
			if !yield(i, v) {
				return
			}
			i++
		}
	}
}
