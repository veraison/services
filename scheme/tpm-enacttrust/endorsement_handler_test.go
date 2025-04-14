// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package tpm_enacttrust

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestDecoder_Decode_OK(t *testing.T) {
	tvs := [][]byte{
		unsignedCorimComidTpmEnactTrustAKOne,
		unsignedCorimComidTpmEnactTrustGoldenOne,
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
			desc:        "multiple verification keys for an instance",
			input:       unsignedCorimComidTpmEnactTrustAKMult,
			expectedErr: `bad key in CoMID at index 0: expecting exactly one AK public key`,
		},
		{
			desc:        "incorrect instance id in the measurement",
			input:       unsignedCorimComidTpmEnactTrustBadInst,
			expectedErr: `bad software component in CoMID at index 0: could not extract instance attributes: could not extract node-id (UUID) from instance-id`,
		},
		{
			desc:        "no instance id specified in the measurement",
			input:       unsignedCorimComidTpmEnactTrustNoInst,
			expectedErr: `bad software component in CoMID at index 0: could not extract instance attributes: expecting instance in environment`,
		},
		{
			desc:        "multiple digest specified in the measurement",
			input:       unsignedCorimComidTpmEnactTrustMultDigest,
			expectedErr: `bad software component in CoMID at index 0: extracting measurement: expecting exactly one digest`,
		},
		{
			desc:        "multiple measurements in ref value triple",
			input:       unsignedCorimComidTpmEnactTrustGoldenTwo,
			expectedErr: `bad software component in CoMID at index 0: expecting one measurement only`,
		},
		{
			desc:        "no digest specified in the measurement",
			input:       unsignedCorimComidTpmEnactTrustNoDigest,
			expectedErr: `bad software component in CoMID at index 0: extracting measurement: measurement value has no digests`,
		},
		{
			desc:        "incorrect instance id specified in the measurement",
			input:       unsignedCorimComidTpmEnactTrustAKBadInst,
			expectedErr: `bad key in CoMID at index 0: could not extract node id: could not extract node-id (UUID) from instance-id`,
		}}

	for _, tv := range tvs {
		t.Run(tv.desc, func(t *testing.T) {
			d := &EndorsementHandler{}
			_, err := d.Decode(tv.input, "", nil)
			assert.EqualError(t, err, tv.expectedErr)
		})
	}
}
