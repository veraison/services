// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package arm_cca

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

	_, err := d.Decode(emptyData)

	assert.EqualError(t, err, expectedErr)
}

func TestDecoder_Decode_invalid_data(t *testing.T) {
	d := &EndorsementHandler{}

	invalidCbor := []byte("invalid CBOR")

	expectedErr := `CBOR decoding failed: expected map (CBOR Major Type 5), found Major Type 3`

	_, err := d.Decode(invalidCbor)

	assert.EqualError(t, err, expectedErr)
}

func TestDecoder_Decode_CcaSsdRefVal_OK(t *testing.T) {
	tvs := [][]byte{
		unsignedCorimCcaComidCcaRefValOne,
		unsignedCorimCcaComidCcaRefValFour,
	}

	d := &EndorsementHandler{}

	for _, tv := range tvs {
		_, err := d.Decode(tv)
		assert.NoError(t, err)
	}
}

func TestDecoder_Decode_CCaSsdRefVal_NOK(t *testing.T) {
	tvs := []struct {
		desc        string
		input       []byte
		expectedErr string
	}{
		{
			desc:        "missing profile inside corim containing one CCA platform config measurement",
			input:       unsignedCorimCcaNoProfileComidCcaRefValOne,
			expectedErr: "no profile information set in CoRIM",
		},
		{
			desc:        "missing profile inside corim containing multiple reference value measurements",
			input:       unsignedCorimCcaNoProfileComidCcaRefValFour,
			expectedErr: "no profile information set in CoRIM",
		},
	}

	for _, tv := range tvs {
		d := &EndorsementHandler{}
		_, err := d.Decode(tv.input)
		assert.EqualError(t, err, tv.expectedErr)
	}
}

func TestDecoder_DecodeCcaRealm_OK(t *testing.T) {
	tvs := [][]byte{
		unsignedCorimCcaRealmComidCcaRealm,
		unsignedCorimCcaRealmComidCcaRealmNoClass,
	}

	d := &EndorsementHandler{}

	for _, tv := range tvs {
		_, err := d.Decode(tv)
		assert.NoError(t, err)
	}
}

func TestDecoder_DecodeCcaRealm_negative_tests(t *testing.T) {
	tvs := []struct {
		desc        string
		input       []byte
		expectedErr string
	}{
		{
			desc:        "no realm instance identity in corim",
			input:       unsignedCorimCcaRealmComidCcaRealmNoInstance,
			expectedErr: "bad software component in CoMID at index 0: could not extract Realm instance attributes: expecting instance in environment",
		},
		{
			desc:        "invalid instance identity in corim",
			input:       unsignedCorimCcaRealmComidCcaRealmInvalidInstance,
			expectedErr: "bad software component in CoMID at index 0: could not extract Realm instance attributes: expecting instance as bytes for CCA Realm",
		},
		{
			desc:        "invalid class identity in corim",
			input:       unsignedCorimCcaRealmComidCcaRealmInvalidClass,
			expectedErr: "bad software component in CoMID at index 0: could not extract Realm class attributes: could not extract uuid from class-id: class-id type is: *comid.TaggedImplID",
		},
	}

	for _, tv := range tvs {
		t.Run(tv.desc, func (t *testing.T) {
			d := &EndorsementHandler{}
			_, err := d.Decode(tv.input)
			assert.EqualError(t, err, tv.expectedErr)
		})
	}
}
