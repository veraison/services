// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package da_spdm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/veraison/corim/comid"
	"github.com/veraison/da"
	"github.com/veraison/ear"
	"github.com/veraison/services/vts/appraisal"
)

var (
	hwConfigName = "hardware-config"
	mutableFwName = "mutable-firmware"

	digest = comid.MustHexDecode(&testing.T{}, "07060504030201000f0e0c0b0817161514131211101f1e1d1c1b1a1918")

	testTriplesMap = map[string]*comid.ValueTriple{
		"spdm:ACME:WIDGET-A": &comid.ValueTriple{
			Environment: comid.Environment{
				Class: &comid.Class{
					ClassID: comid.MustNewBytesClassID(
						[]byte("spdm:ACME:WIDGET-A"),
					),
				},
			},
			Measurements: *comid.NewMeasurements().Add(
				&comid.Measurement{
					Key: comid.MustNewMkey(uint(1), comid.UintType),
					Val: comid.Mval{
						Name: &hwConfigName,
						RawValue: comid.NewRawValueFromBytes(
							comid.MustHexDecode(&testing.T{}, "4f6d616861"),
						),
					},
				},
			),
		},
		"spdm:C=CA,O=ACME,OU=Widget-B": &comid.ValueTriple{
			Environment: comid.Environment{
				Class: &comid.Class{
					ClassID: comid.MustNewBytesClassID(
						[]byte("spdm:C=CA,O=ACME,OU=Widget-B"),
					),
				},
			},
			Measurements: *comid.NewMeasurements().Add(
				&comid.Measurement{
					Key: comid.MustNewMkey(uint(1), comid.UintType),
					Val: comid.Mval{
						Name: &mutableFwName,
						Digests: comid.NewDigests().
							AddDigest(comid.Sha256, digest),
					},
				},
			).Add(
				&comid.Measurement{
					Key: comid.MustNewMkey(uint(6), comid.UintType),
					Val: comid.Mval{
						Name: &hwConfigName,
						Digests: comid.NewDigests().
							AddDigest(comid.Sha256, digest),
					},
				},
			),
		},
	}
)

func Test_evidenceToTriplesMap(t *testing.T) {
	testCases := []struct{
		title string
		evidence appraisal.Evidence
		expected map[string]*comid.ValueTriple
		err string
	}{
		{
			title: "ok",
			evidence: appraisal.Evidence{
				Nonce: comid.MustHexDecode(t, "f9efc3341597f75f8d94432ad39566a8c5704b2004ba001c094f475bfc057f9f25d7aa40cd86cd30ebaae746fb19f008c1e6a1f23ad6a178e18dceda918f7f6e"),
				Data: daSpdmGood,
			},
			expected: testTriplesMap,
		},
		{
			title: "bad nonce",
			evidence: appraisal.Evidence{
				Nonce: comid.MustHexDecode(t, "deadbeef"),
				Data: daSpdmGood,
			},
			err: "nonce in evidence does not match session nonce",
		},
		{
			title: "bad CBOR",
			evidence: appraisal.Evidence{
				Nonce: comid.MustHexDecode(t, "deadbeef"),
				Data: comid.MustHexDecode(t, "deadbeef"),
			},
			err: "cbor: invalid additional information",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			ret, err := evidenceToTriplesMap(&tc.evidence)
			if tc.err == "" {
				assert.NoError(t, err)
				assert.EqualValues(t, tc.expected, ret)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func Test_daTokenToTriplesMap_bad(t *testing.T) {
	testCases := []struct{
		title string
		token da.Token
		err string
	}{
		{
			title: "no submods",
			token: da.Token{},
			err: "no submods in DA token",
		},
		{
			title: "bad submod profile",
			token: da.Token{
				EatSubmods: map[string]da.SPDMClaims{
					"foo": da.SPDMClaims{
						EatProfile: "bar",
					},
				},
			},
			err: `submod["foo"]: unsupported profile "bar"`,
		},
		{
			title: "no measurements",
			token: da.Token{
				EatSubmods: map[string]da.SPDMClaims{
					"foo": da.SPDMClaims{
						EatProfile: DaDeviceSpdmProfile,
					},
				},
			},
			err: `submod["foo"]: no measurements`,
		},
		{
			title: "bad component type",
			token: da.Token{
				EatSubmods: map[string]da.SPDMClaims{
					"foo": da.SPDMClaims{
						EatProfile: DaDeviceSpdmProfile,
						Measurements: map[uint8]da.SPDMMeasurement{
							1: da.SPDMMeasurement{
							},
						},
					},
				},
			},
			err: `submod["foo"]: measurement[1]: unexpected component type 0`,
		},
		{
			title: "no measurement value",
			token: da.Token{
				EatSubmods: map[string]da.SPDMClaims{
					"foo": da.SPDMClaims{
						EatProfile: DaDeviceSpdmProfile,
						Measurements: map[uint8]da.SPDMMeasurement{
							1: da.SPDMMeasurement{
								ComponentType: 1,
							},
						},
					},
				},
			},
			err: `submod["foo"]: measurement[1]: neither raw measurement nor digest are set`,
		},
		{
			title: "bad digest algorithm",
			token: da.Token{
				EatSubmods: map[string]da.SPDMClaims{
					"foo": da.SPDMClaims{
						EatProfile: DaDeviceSpdmProfile,
						Measurements: map[uint8]da.SPDMMeasurement{
							1: da.SPDMMeasurement{
								ComponentType: 1,
								DigestedMeasurement: &da.Digest{
									Algorithm: 255,
									Value: comid.MustHexDecode(t, "deadbeef"),
								},
							},
						},
					},
				},
			},
			err: `submod["foo"]: measurement[1]: unknown algorithm 255`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			_, err := daTokenToTriplesMap(&tc.token)
			assert.ErrorContains(t, err, tc.err)
		})
	}
}

func TestImplementation_GetReferenceValueIDs(t *testing.T) {
	claims := make(map[string]any)
	for submodName, triple := range testTriplesMap {
		claims[submodName] = triple
	}

	var impl Implementation

	ret, err := impl.GetReferenceValueIDs(nil, claims)
	assert.NoError(t, err)
	assert.Len(t, ret, 2)

	for _, environment := range ret {
		classIdString := string(environment.Class.ClassID.Bytes()) // nolint:staticcheck
		_, ok := claims[classIdString]
		assert.True(t, ok)
	}
}

func Test_matchMeasurement(t *testing.T) {
	testCases := []struct{
		title string
		reference []comid.Measurement
		evidence []comid.Measurement
		expected *MatchResult
		err string
	}{
		{
			title: "ok digests matched",
			reference: []comid.Measurement{
				comid.Measurement{
					Key: comid.MustNewMkey(uint(1), comid.UintType),
					Val: comid.Mval{
						Name: &hwConfigName,
						Digests: comid.NewDigests().
							AddDigest( 1, comid.MustHexDecode(t,"4f6d616861")),
					},
				},
			},
			evidence: []comid.Measurement{
				comid.Measurement{
					Key: comid.MustNewMkey(uint(1), comid.UintType),
					Val: comid.Mval{
						Name: &hwConfigName,
						Digests: comid.NewDigests().
							AddDigest( 1, comid.MustHexDecode(t,"4f6d616861")),
					},
				},
			},
			expected: &MatchResult{
				Matched: true,
			},
		},
		{
			title: "ok raw-value matched",
			reference: []comid.Measurement{
				comid.Measurement{
					Key: comid.MustNewMkey(uint(1), comid.UintType),
					Val: comid.Mval{
						Name: &hwConfigName,
						RawValue: comid.NewRawValueFromBytes(
							comid.MustHexDecode(t, "4f6d616861"),
						),
					},
				},
			},
			evidence: []comid.Measurement{
				comid.Measurement{
					Key: comid.MustNewMkey(uint(1), comid.UintType),
					Val: comid.Mval{
						Name: &hwConfigName,
						RawValue: comid.NewRawValueFromBytes(
							comid.MustHexDecode(t, "4f6d616861"),
						),
					},
				},
			},
			expected: &MatchResult{
				Matched: true,
			},
		},
		{
			title: "nok different raw-values",
			reference: []comid.Measurement{
				comid.Measurement{
					Key: comid.MustNewMkey(uint(1), comid.UintType),
					Val: comid.Mval{
						Name: &hwConfigName,
						RawValue: comid.NewRawValueFromBytes(
							comid.MustHexDecode(t, "4f6d616861"),
						),
					},
				},
			},
			evidence: []comid.Measurement{
				comid.Measurement{
					Key: comid.MustNewMkey(uint(1), comid.UintType),
					Val: comid.Mval{
						Name: &hwConfigName,
						RawValue: comid.NewRawValueFromBytes(
							comid.MustHexDecode(t, "deadbeef"),
						),
					},
				},
			},
			expected: &MatchResult{
				Matched: false,
				Reason: "mkey 1: raw-value (deadbeef) did not match reference (4f6d616861)",
			},
		},
		{
			title: "nok different digests",
			reference: []comid.Measurement{
				comid.Measurement{
					Key: comid.MustNewMkey(uint(1), comid.UintType),
					Val: comid.Mval{
						Name: &hwConfigName,
						Digests: comid.NewDigests().
							AddDigest( 1, comid.MustHexDecode(t,"4f6d616861")),
					},
				},
			},
			evidence: []comid.Measurement{
				comid.Measurement{
					Key: comid.MustNewMkey(uint(1), comid.UintType),
					Val: comid.Mval{
						Name: &hwConfigName,
						Digests: comid.NewDigests().
							AddDigest( 1, comid.MustHexDecode(t,"deadbeef")),
					},
				},
			},
			expected: &MatchResult{
				Matched: false,
				Reason: "mkey 1: digests did not match reference",
			},
		},
		{
			title: "nok wrong measurement type",
			reference: []comid.Measurement{
				comid.Measurement{
					Key: comid.MustNewMkey(uint(1), comid.UintType),
					Val: comid.Mval{
						Name: &hwConfigName,
						RawValue: comid.NewRawValueFromBytes(
							comid.MustHexDecode(t, "4f6d616861"),
						),
					},
				},
			},
			evidence: []comid.Measurement{
				comid.Measurement{
					Key: comid.MustNewMkey(uint(1), comid.UintType),
					Val: comid.Mval{
						Name: &hwConfigName,
						Digests: comid.NewDigests().
							AddDigest( 1, comid.MustHexDecode(t,"deadbeef")),
					},
				},
			},
			expected: &MatchResult{
				Matched: false,
				Reason: "mkey 1: evidence does not contain raw-value",
			},
		},
		{
			title: "nok wrong mkey",
			reference: []comid.Measurement{
				comid.Measurement{
					Key: comid.MustNewMkey(uint(1), comid.UintType),
					Val: comid.Mval{
						Name: &hwConfigName,
						RawValue: comid.NewRawValueFromBytes(
							comid.MustHexDecode(t, "4f6d616861"),
						),
					},
				},
			},
			evidence: []comid.Measurement{
				comid.Measurement{
					Key: comid.MustNewMkey(uint(2), comid.UintType),
					Val: comid.Mval{
						Name: &hwConfigName,
						RawValue: comid.NewRawValueFromBytes(
							comid.MustHexDecode(t, "4f6d616861"),
						),
					},
				},
			},
			expected: &MatchResult{
				Matched: false,
				Reason: "failed to match mkey 1",
			},
		},
		{
			title: "nok wrong name",
			reference: []comid.Measurement{
				comid.Measurement{
					Key: comid.MustNewMkey(uint(1), comid.UintType),
					Val: comid.Mval{
						Name: &hwConfigName,
						RawValue: comid.NewRawValueFromBytes(
							comid.MustHexDecode(t, "4f6d616861"),
						),
					},
				},
			},
			evidence: []comid.Measurement{
				comid.Measurement{
					Key: comid.MustNewMkey(uint(1), comid.UintType),
					Val: comid.Mval{
						Name: &mutableFwName,
						RawValue: comid.NewRawValueFromBytes(
							comid.MustHexDecode(t, "4f6d616861"),
						),
					},
				},
			},
			expected: &MatchResult{
				Matched: false,
				Reason: `mkey 1: reference ("hardware-config") and evidence ("mutable-firmware") names don't match`,
			},
		},
		{
			title: "err bad reference mkey",
			reference: []comid.Measurement{
				comid.Measurement{
					Key: comid.MustNewMkey("foo", comid.StringType),
				},
			},
			evidence: []comid.Measurement{
				comid.Measurement{
					Key: comid.MustNewMkey(uint(1), comid.UintType),
				},
			},
			err: "reference measurement[0]: measurement-key type is: *comid.StringMkey",
		},
		{
			title: "err bad evidence mkey",
			reference: []comid.Measurement{
				comid.Measurement{
					Key: comid.MustNewMkey(uint(1), comid.UintType),
				},
			},
			evidence: []comid.Measurement{
				comid.Measurement{
					Key: comid.MustNewMkey("foo", comid.StringType),
				},
			},
			err: "evidence measurement[0]: measurement-key type is: *comid.StringMkey",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			result, err := matchMeasurements(tc.reference, tc.evidence)
			if tc.expected != nil {
				assert.NoError(t, err)
				assert.EqualValues(t, *tc.expected, result)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestImplementation_AppraiseClaims(t *testing.T) {
	claims := make(map[string]any)
	endorsements := make([]*comid.ValueTriple, 0, len(testTriplesMap))
	for submodName, triple := range testTriplesMap {
		claims[submodName] = triple
		endorsements = append(endorsements, triple)
	}
	implementation := &Implementation{}

	result, err := implementation.AppraiseClaims(claims, endorsements)
	assert.NoError(t, err)

	for _, appraisal := range result.Submods {
		assert.Equal(t, ear.TrustTierAffirming, *appraisal.Status)
	}

	result, err = implementation.AppraiseClaims(claims, nil)
	assert.NoError(t, err)

	for _, appraisal := range result.Submods {
		assert.Equal(t, ear.TrustTierContraindicated, *appraisal.Status)
	}
}

func Test_classNameFromSubmodName(t *testing.T) {
	testCases := []struct{
		title string
		submodName string
		expected string
		err string
	}{
		{
			title: "ok SAN with CN",
			submodName: "spdm:C=CA,O=ACME,OU=Widget-B,CN=9876543210",
			expected: "spdm:C=CA,O=ACME,OU=Widget-B",
		},
		{
			title: "ok SAN without CN",
			submodName: "spdm:C=CA,O=ACME,OU=Widget-B",
			expected: "spdm:C=CA,O=ACME,OU=Widget-B",
		},
		{
			title: "ok ub-DMTF-device-info",
			submodName: "spdm:ACME:WIDGET-A:0123456789",
			expected: "spdm:ACME:WIDGET-A",
		},
		{
			title: "err not namespace",
			submodName: "foo",
			err: "submod name must start with spdm namespace",
		},
		{
			title: "err bad fromat",
			submodName: "spdm:foo",
			err: `could not derive class from "spdm:foo"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			result, err := classNameFromSubmodName(tc.submodName)
			if tc.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}
