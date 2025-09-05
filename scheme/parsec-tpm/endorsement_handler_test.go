// Copyright 2023-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecoder_Decode_OK(t *testing.T) {
	tvs := [][]byte{
		unsignedCorimComidParsecTpmKeyGood,
		unsignedCorimComidParsecTpmPcrsGood,
	}

	d := &EndorsementHandler{}

	for _, tv := range tvs {
		_, err := d.Decode(tv, "", nil)
		assert.NoError(t, err)
	}
}

func TestDecoder_Decode_negative_tests(t *testing.T) {
	tvs := []struct {
		desc        string
		input       []byte
		expectedErr string
	}{
		{
			desc:        "key without instance identifier",
			input:       unsignedCorimComidParsecTpmKeyNoInstance,
			expectedErr: `bad key in CoMID at index 0: failed to create trust anchor raw public key: instance not found in ID`,
		},
		{
			desc:        "key with an instance identifier of an unexpected type",
			input:       unsignedCorimComidParsecTpmKeyUnknownInstanceType,
			expectedErr: `bad key in CoMID at index 0: could not extract id from AVK environment: could not extract instance-id (UEID) from instance: instance-id type is: *comid.TaggedUUID`,
		},
		{
			desc:        "key without class",
			input:       unsignedCorimComidParsecTpmKeyNoClass,
			expectedErr: `bad key in CoMID at index 0: could not extract id from AVK environment: class not found in environment`,
		},
		{
			desc:        "key without class id",
			input:       unsignedCorimComidParsecTpmKeyNoClassId,
			expectedErr: `bad key in CoMID at index 0: could not extract id from AVK environment: class-id not found in class`,
		},
		{
			desc:        "key class id of an unexpected type",
			input:       unsignedCorimComidParsecTpmKeyUnknownClassIdType,
			expectedErr: `bad key in CoMID at index 0: could not extract id from AVK environment: class-id not in UUID format`,
		},
		{
			desc:        "key with multiple keys",
			input:       unsignedCorimComidParsecTpmKeyManyKeys,
			expectedErr: `bad key in CoMID at index 0: expecting exactly one AK public key`,
		},
		{
			desc:        "measurement without class",
			input:       unsignedCorimComidParsecTpmPcrsNoClass,
			expectedErr: `bad software component in CoMID at index 0: could not extract id from ref-val environment: class not found in environment`,
		},
		{
			desc:        "measurement without the associated PCR",
			input:       unsignedCorimComidParsecTpmPcrsNoPCR,
			expectedErr: `bad software component in CoMID at index 0: could not extract PCR: measurement key is not present`,
		},
		{
			desc:        "measurement with PCR of an unexpected type",
			input:       unsignedCorimComidParsecTpmPcrsUnknownPCRType,
			expectedErr: `bad software component in CoMID at index 0: could not extract PCR: measurement key is not uint: measurement-key type is: *comid.TaggedUUID`,
		},
		{
			desc:        "measurement with PCR without digests",
			input:       unsignedCorimComidParsecTpmPcrsNoDigests,
			expectedErr: `bad software component in CoMID at index 0: measurement[0]: measurement value does not contain digests`,
		},
	}

	for _, tv := range tvs {
		t.Run(tv.desc, func(t *testing.T) {
			d := &EndorsementHandler{}
			_, err := d.Decode(tv.input, "", nil)
			assert.EqualError(t, err, tv.expectedErr)
		})
	}
}

func TestDecoder_GetAttestationScheme(t *testing.T) {
	d := &EndorsementHandler{}

	expected := SchemeName

	actual := d.GetAttestationScheme()

	assert.Equal(t, expected, actual)
}

func TestDecoder_GetSupportedMediaTypes(t *testing.T) {
	d := &EndorsementHandler{}

	expected := EndorsementMediaTypes

	actual := d.GetSupportedMediaTypes()

	assert.Equal(t, expected, actual)
}

func TestDecoder_Init(t *testing.T) {
	d := &EndorsementHandler{}

	assert.Nil(t, d.Init(nil))
}

func TestDecoder_Close(t *testing.T) {
	d := &EndorsementHandler{}

	assert.Nil(t, d.Close())
}

func TestDecoder_Decode_empty_data(t *testing.T) {
	d := &EndorsementHandler{}

	emptyData := []byte{}

	expectedErr := `empty data`

	_, err := d.Decode(emptyData, "", nil)

	assert.EqualError(t, err, expectedErr)
}

func TestDecoder_Decode_invalid_data(t *testing.T) {
	d := &EndorsementHandler{}

	invalidCbor := []byte("invalid CBOR")

	expectedErr := `CBOR decoding failed: expected map (CBOR Major Type 5), found Major Type 3`

	_, err := d.Decode(invalidCbor, "", nil)

	assert.EqualError(t, err, expectedErr)
}
