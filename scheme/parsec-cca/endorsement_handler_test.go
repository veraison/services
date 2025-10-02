// Copyright 2023-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_cca

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecoder_Decode_OK(t *testing.T) {
	tvs := [][]byte{
		unsignedCorimComidParsecCcaRefValOne,
		unsignedCorimComidParsecCcaMultRefVal,
	}

	d := &EndorsementHandler{}

	for _, tv := range tvs {
		_, err := d.Decode(tv, "", nil)
		assert.NoError(t, err)
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
	expectedName := EndorsementHandlerName
	name := d.GetName()
	assert.Equal(t, name, expectedName)
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
