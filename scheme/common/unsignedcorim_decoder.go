// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/services/handler"
)

func UnsignedCorimDecoder(
	data []byte,
	xtr IExtractor,
	signerThumbprint ...string,
) (*handler.EndorsementHandlerResponse, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}

	var uc corim.UnsignedCorim
	if err := uc.FromCBOR(data); err != nil {
		return nil, fmt.Errorf("CBOR decoding failed: %w", err)
	}

	if err := uc.Valid(); err != nil {
		return nil, fmt.Errorf("invalid unsigned corim: %w", err)
	}

	if uc.Profile != nil {
		profile, err := uc.Profile.Get()
		if err != nil {
			return nil, fmt.Errorf("failed to get the profile information: %w", err)
		}
		xtr.SetProfile(profile)
	} else {
		return nil, fmt.Errorf("no profile information set in CoRIM")
	}

	rsp := handler.EndorsementHandlerResponse{}

	for i, tag := range uc.Tags {
		// need at least 3 bytes for the tag and 1 for the smallest bstr
		if len(tag) < 3+1 {
			return nil, fmt.Errorf("malformed tag at index %d", i)
		}

		// split tag from data
		cborTag, cborData := tag[:3], tag[3:]

		// The EnactTrust profile only knows about CoMIDs
		if !bytes.Equal(cborTag, corim.ComidTag) {
			return nil, fmt.Errorf("unknown CBOR tag %x detected at index %d", cborTag, i)
		}

		var c comid.Comid

		err := c.FromCBOR(cborData)
		if err != nil {
			return nil, fmt.Errorf("decoding failed for CoMID at index %d: %w", i, err)
		}

		if err := c.Valid(); err != nil {
			return nil, fmt.Errorf("decoding failed for CoMID at index %d: %w", i, err)
		}

		if c.Triples.ReferenceValues != nil {
			refVals, err := xtr.RefValExtractor(*c.Triples.ReferenceValues)
			if err != nil {
				return nil, fmt.Errorf(
					"bad software component in CoMID at index %d: %w",
					i,
					err,
				)
			}

			for _, refVal := range refVals {
				rsp.ReferenceValues = append(rsp.ReferenceValues, *refVal)
			}
		}

		if c.Triples.AttestVerifKeys != nil {
			for _, avk := range *c.Triples.AttestVerifKeys {
				k, err := xtr.TaExtractor(avk)
				if err != nil {
					return nil, fmt.Errorf(
						"bad key in CoMID at index %d: %w",
						i,
						err,
					)
				}

				rsp.TrustAnchors = append(rsp.TrustAnchors, *k)
			}
		}

		// silently ignore any other triples
	}

	return &rsp, nil
}
