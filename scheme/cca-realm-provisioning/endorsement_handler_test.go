// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package cca_realm_provisioning

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/veraison/corim/comid"
)

func TestDecoder_Decode_OK(t *testing.T) {
	tvs := []string{
		unsignedCorimcomidCcaRealm,
		unsignedCorimcomidCcaRealmNoClass,
	}

	d := &EndorsementHandler{}

	for _, tv := range tvs {
		data := comid.MustHexDecode(t, tv)
		_, err := d.Decode(data)
		assert.NoError(t, err)
	}
}

func TestDecoder_Decode_negative_tests(t *testing.T) {
	tvs := []struct {
		desc        string
		input       string
		expectedErr string
	}{
		{
			desc:        "no realm instance identity in corim",
			input:       unsignedCorimcomidCcaRealmNoInstance,
			expectedErr: "bad software component in CoMID at index 0: could not extract Realm instance attributes: expecting instance in environment",
		},
		{
			desc:        "invalid instance identity in corim",
			input:       unsignedCorimcomidCcaRealmInvalidInstance,
			expectedErr: "bad software component in CoMID at index 0: could not extract Realm instance attributes: expecting instance as bytes for CCA Realm",
		},
		{
			desc:        "invalid class identity in corim",
			input:       unsignedCorimcomidCcaRealmInvalidClass,
			expectedErr: "bad software component in CoMID at index 0: could not extract Realm class attributes: could not extract uu-id from class-id: class-id type is: *comid.TaggedImplID",
		},
	}

	for _, tv := range tvs {
		data := comid.MustHexDecode(t, tv.input)
		d := &EndorsementHandler{}
		_, err := d.Decode(data)
		assert.EqualError(t, err, tv.expectedErr)
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

func TestDecoder_GetName_ok(t *testing.T) {
	d := &EndorsementHandler{}
	expectedName := "unsigned-corim (CCA realm profile)"
	name := d.GetName()
	assert.Equal(t, name, expectedName)
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
