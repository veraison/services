// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package cca_ssd_platform

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/veraison/corim/comid"
)

func TestDecoder_GetAttestationScheme(t *testing.T) {
	d := &EndorsementDecoder{}

	expected := SchemeName

	actual := d.GetAttestationScheme()

	assert.Equal(t, expected, actual)
}

func TestDecoder_GetSupportedMediaTypes(t *testing.T) {
	d := &EndorsementDecoder{}

	expected := EndorsementMediaTypes

	actual := d.GetSupportedMediaTypes()

	assert.Equal(t, expected, actual)
}

func TestDecoder_Init(t *testing.T) {
	d := &EndorsementDecoder{}

	assert.Nil(t, d.Init(nil))
}

func TestDecoder_Close(t *testing.T) {
	d := &EndorsementDecoder{}

	assert.Nil(t, d.Close())
}

func TestDecoder_Decode_empty_data(t *testing.T) {
	d := &EndorsementDecoder{}

	emptyData := []byte{}

	expectedErr := `empty data`

	_, err := d.Decode(emptyData)

	assert.EqualError(t, err, expectedErr)
}

func TestDecoder_Decode_invalid_data(t *testing.T) {
	d := &EndorsementDecoder{}

	invalidCbor := []byte("invalid CBOR")

	expectedErr := `CBOR decoding failed: cbor: cannot unmarshal UTF-8 text string into Go value of type corim.UnsignedCorim`

	_, err := d.Decode(invalidCbor)

	assert.EqualError(t, err, expectedErr)
}

func TestDecoder_Decode_CcaRefVal_OK(t *testing.T) {
	tvs := []string{
		unsignedCorimComidCcaRefValOne,
		unsignedCorimComidCcaRefValFour,
	}

	d := &EndorsementDecoder{}

	for _, tv := range tvs {
		data := comid.MustHexDecode(t, tv)
		_, err := d.Decode(data)
		assert.NoError(t, err)
	}
}

func TestDecoder_Decode_CCaRefVal_NOK(t *testing.T) {
	tvs := []struct {
		desc        string
		input       string
		expectedErr string
	}{
		{
			desc:        "missing profile inside corim containing one CCA platform config measurement",
			input:       unsignedCorimNoProfileComidCcaRefValOne,
			expectedErr: "no profile information set in CoRIM",
		},
		{
			desc:        "missing profile inside corim containing multiple reference value measurements",
			input:       unsignedCorimNoProfileComidCcaRefValFour,
			expectedErr: "no profile information set in CoRIM",
		},
	}

	for _, tv := range tvs {
		data := comid.MustHexDecode(t, tv.input)
		d := &EndorsementDecoder{}
		_, err := d.Decode(data)
		assert.EqualError(t, err, tv.expectedErr)
	}
}
