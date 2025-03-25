// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package arm

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/services/handler"
	"github.com/veraison/swid"
)

func EndorsementToReferenceValueTriple(e handler.Endorsement) (*comid.ValueTriple, error) {
	var attrs SwAttr

	if err := json.Unmarshal(e.Attributes, &attrs); err != nil {
		return nil, fmt.Errorf("unmarshalling attributes: %w", err)
	}

	// mkey

	rvID, err := comid.NewPSARefValID(attrs.SignerID)
	if err != nil {
		return nil, fmt.Errorf("instantiating PSA reference value ID: %w", err)
	}

	label := attrs.MeasurementType
	if label != "" {
		rvID.SetLabel(label)
	}

	version := attrs.Version
	if version != "" {
		rvID.SetVersion(version)
	}

	// mval

	m, err := comid.NewPSAMeasurement(rvID)
	if err != nil {
		return nil, fmt.Errorf("instantiating PSA measurement: %w", err)
	}

	m.AddDigest(
		swid.AlgIDFromString(attrs.MeasDesc),
		attrs.MeasurementValue,
	)

	measurements := comid.NewMeasurements().Add(m)

	// env

	class := comid.NewClassImplID(comid.ImplID(attrs.ImplID))
	if class == nil {
		return nil, errors.New("class identifier instantiation failed")
	}

	model := attrs.Model
	if model != "" {
		class.SetModel(model)
	}

	vendor := attrs.Vendor
	if vendor != "" {
		class.SetVendor(vendor)
	}

	env := comid.Environment{
		Class: class,
	}

	// rv triple

	return &comid.ValueTriple{
		Environment:  env,
		Measurements: *measurements,
	}, nil
}

func EndorsementToAttestationKeyTriple(e handler.Endorsement) (*comid.KeyTriple, error) {
	var attrs TaAttr

	if err := json.Unmarshal(e.Attributes, &attrs); err != nil {
		return nil, fmt.Errorf("unmarshalling attributes: %w", err)
	}

	// cryptokeys

	k, err := comid.NewPKIXBase64Key(attrs.VerifKey)
	if err != nil {
		return nil, fmt.Errorf("crypto key instantiation failed: %w", err)
	}

	ak := comid.NewCryptoKeys().Add(k)

	// env

	instance, err := comid.NewUEIDInstance(attrs.InstID)
	if err != nil {
		return nil, fmt.Errorf("instance identifier instantiation failed: %w", err)
	}

	class := comid.NewClassImplID(comid.ImplID(attrs.ImplID))
	if class == nil {
		return nil, errors.New("class identifier instantiation failed")
	}

	model := attrs.Model
	if model != "" {
		class.SetModel(model)
	}

	vendor := attrs.Vendor
	if vendor != "" {
		class.SetVendor(vendor)
	}

	env := comid.Environment{
		Class:    class,
		Instance: instance,
	}

	return &comid.KeyTriple{
		Environment: env,
		VerifKeys:   *ak,
	}, nil
}
