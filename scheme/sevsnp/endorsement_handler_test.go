// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sevsnp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecoder_GetName(t *testing.T) {
	d := &EndorsementHandler{}

	expected := SchemeName

	actual := d.GetName()

	assert.Equal(t, expected, actual)
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

func TestDecoder_Decode_OK(t *testing.T) {
	d := &EndorsementHandler{}

	_, err := d.Decode(unsignedCorimSevSnp, "", nil)
	assert.NoError(t, err)
}
