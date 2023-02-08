// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/services/handler"
)

func UnsignedCorimDecoder(
	data []byte,
	xtr IExtractor,
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

	if uc.Profiles != nil {
		// get the profile
		if len(*uc.Profiles) > 1 {
			var profiles []string
			for _, p := range *uc.Profiles {
				name, _ := p.Get()
				profiles = append(profiles, name)
			}
			return nil, fmt.Errorf("found multiple profiles (expected exactly one): %s", strings.Join(profiles, ", "))
		}
		p := (*uc.Profiles)[0]
		_, err := p.Get()
		if err != nil {
			return nil, fmt.Errorf("failed to get the profile information: %w", err)
		}
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
			for _, rv := range *c.Triples.ReferenceValues {
				refVal, err := xtr.RefValExtractor(rv)
				if err != nil {
					return nil, fmt.Errorf("bad software component in CoMID at index %d: %w", i, err)
				}

				for i := range refVal {
					rsp.ReferenceValues = append(rsp.ReferenceValues, refVal[i])

				}
			}
		}

		if c.Triples.AttestVerifKeys != nil {
			for _, avk := range *c.Triples.AttestVerifKeys {
				k, err := xtr.TaExtractor(avk)
				if err != nil {
					return nil, fmt.Errorf("bad key in CoMID at index %d: %w", i, err)
				}

				rsp.TrustAnchors = append(rsp.TrustAnchors, k)
			}
		}

		// silently ignore any other triples
	}

	return &rsp, nil
}
