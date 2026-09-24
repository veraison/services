// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package lifecycle

import (
	"errors"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/eat"
	"github.com/veraison/swid"
)

type Query struct {
	coserv.Query

	Profile *eat.Profile `cbor:"265,keyasint,omitempty" json:"profile,omitempty"`
}

// NewEnvironmentQuery creates a new environment Query instance.
// An error is returned if the supplied environment selector is invalid.
func NewEnvironmentQuery(
	profile eat.Profile,
	artifactType coserv.ArtifactType,
	envSelector coserv.EnvironmentSelector,
) (*Query, error) {
	if err := envSelector.Valid(); err != nil {
		return nil, fmt.Errorf("invalid environment selector: %w", err)
	}

	return &Query{
		Query: coserv.Query{
			ArtifactType:        &artifactType,
			EnvironmentSelector: &envSelector,
		},
		Profile: &profile,
	}, nil
}

// NewRimQuery creates a new RIM Query instance. An error is returned if the
// supplied arguments are invalid.
func NewRimQuery(typ coserv.RimSelectorType, tagID swid.TagID) (*Query, error) {
	selector, err := coserv.NewRimSelectorID(typ, tagID)
	if err != nil {
		return nil, err
	}

	return &Query{
		Query: coserv.Query{
			RimSelector: coserv.NewRimSelectorIDs().Add(selector),
		},
	}, nil
}

// Valid ensures that the Query target is correctly populated.
func (o Query) Valid() error {
	if o.ResultType != nil {
		return errors.New("result type cannot be specified in an endorsement lifecycle management (ELM) query")
	}

	if o.EnvironmentSelector != nil { // nolint:gocritic
		if o.RimSelector != nil {
			return errors.New("environment and RIM selectors cannot be specified at the same time")
		}

		if o.Profile == nil {
			return errors.New("profile must be specified with an environment selector")
		}

		if o.ArtifactType == nil {
			return errors.New("artifact type must be specified with an environment selector")
		}

		if err := o.EnvironmentSelector.Valid(); err != nil {
			return fmt.Errorf("invalid environment selector: %w", err)
		}

		return nil
	} else if o.RimSelector != nil {
		if o.ArtifactType != nil {
			return errors.New("artifact type cannot be specified with a RIM selector")
		}

		if o.Profile != nil {
			return errors.New("profile cannot be specified with a RIM selector")
		}

		if err := o.RimSelector.Valid(); err != nil {
			return fmt.Errorf("invalid RIM selector: %w", err)
		}

		return nil
	} else {
		return errors.New("no selector specified")
	}
}

// ToEDN encodes the target Query to CBOR Extended Diagnostic Notation (EDN)
func (o Query) ToEDN() (string, error) {
	b, err := o.ToCBOR()
	if err != nil {
		return "", fmt.Errorf("failed encoding ELM Query object: %w", err)
	}
	return cbor.Diagnose(b)
}

// ToCBOR validates and serializes the target Query to CBOR.
// An error is returned if either validation or encoding of the Query target fails.
func (o Query) ToCBOR() ([]byte, error) {
	if err := o.Valid(); err != nil {
		return nil, fmt.Errorf("validating ELM Query: %w", err)
	}

	opts := cbor.CoreDetEncOptions()
	opts.Time = cbor.TimeRFC3339
	opts.TimeTag = 1
	em, err := opts.EncMode()
	if err != nil {
		return nil, fmt.Errorf("CBOR encoding setup failed: %w", err)
	}

	data, err := em.Marshal(o)
	if err != nil {
		return nil, fmt.Errorf("encoding ELM Query to CBOR: %w", err)
	}

	return data, nil
}

// FromCBOR deserializes from CBOR into the target Query.
// An error is returned if either decoding or validation of the Query payload fails.
func (o *Query) FromCBOR(data []byte) error {
	if err := cbor.Unmarshal(data, o); err != nil {
		return fmt.Errorf("decoding ELM Query from CBOR: %w", err)
	}

	if err := o.Valid(); err != nil {
		return fmt.Errorf("validating ELM Query: %w", err)
	}

	return nil
}

/************************************************************
JSON serialization and deserialization depends on coserv.Query
having JSON support. Commenting out this section until a decision
is made.
See: https://github.com/veraison/corim/issues/282
*************************************************************

// ToJSON validates and serializes the target Query to JSON.
// An error is returned if either validation or encoding of the Query target fails.
func (o Query) ToJSON() ([]byte, error) {
	if err := o.Valid(); err != nil {
		return nil, fmt.Errorf("validating ELM Query: %w", err)
	}

	return encoding.SerializeStructToJSON(o)
}

// FromJSON deserializes from JSON into the target Query.
// An error is returned if either decoding or validation of the Query payload fails.
func (o *Query) FromJSON(data []byte) error {
	if err := encoding.PopulateStructFromJSON(data, o); err != nil {
		return fmt.Errorf("decoding ELM query from JSON: %w", err)
	}

	if err := o.Valid(); err != nil {
		return fmt.Errorf("validating ELM query: %w", err)
	}

	return nil
}

************************************************************/
