// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package compositeevidenceparser

import (
	"encoding/hex"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Parse_CBORCollection_OK(t *testing.T) {
	expected := []ComponentEvidence{
		{
			label:       "bretwaldadom",
			data:        []byte{0xa1, 0x0a},
			mediaType:   "application/eat-ucs+cbor",
			parentLabel: "",
			depth:       0,
		},
		{
			label:       "polyscopic",
			data:        []byte(`{"eat_nonce": ...}`),
			mediaType:   "application/eat-ucs+json",
			parentLabel: "murmurless",
			depth:       1,
		},
		{
			label:       "photoelectrograph",
			data:        []byte{0x82, 0x78, 0x18},
			mediaType:   "application/eat-ucs+cbor",
			parentLabel: "",
			depth:       1,
		},
	}

	b := mustReadFile(t, "testdata/collection-1.cbor")

	var cmwParser cmwParser
	ev, err := cmwParser.Parse(b)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, ev)
}

func Test_ParseJSONCollection_OK(t *testing.T) {
	expected := []ComponentEvidence{
		{
			label:       "bretwaldadom",
			data:        []byte{0xa1, 0x0a},
			mediaType:   "application/eat-ucs+cbor",
			parentLabel: "",
			depth:       0,
		},
		{
			label:       "polyscopic",
			data:        []byte(`{"eat_nonce": ...}`),
			mediaType:   "application/eat-ucs+json",
			parentLabel: "murmurless",
			depth:       1,
		},
		{
			label:       "photoelectrograph",
			data:        []byte{0x82, 0x78, 0x18},
			mediaType:   "application/eat-ucs+cbor",
			parentLabel: "",
			depth:       1,
		},
	}

	b := mustReadFile(t, "testdata/collection-1.json")

	var cmwParser cmwParser
	ev, err := cmwParser.Parse(b)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, ev)
}

func Test_Parse_MixedKey_Collection_OK(t *testing.T) {
	expected := []ComponentEvidence{
		{
			label:       "1024",
			data:        []byte{0xaa},
			mediaType:   "text/plain; charset=utf-8",
			parentLabel: "",
			depth:       0,
		},
		{
			label:       "string",
			data:        []byte{0xff},
			mediaType:   "text/plain; charset=utf-8",
			parentLabel: "",
			depth:       0,
		},
	}

	tv := mustReadFile(t, "testdata/collection-cbor-mixed-keys.cbor")

	var cmwParser cmwParser
	ev, err := cmwParser.Parse(tv)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, ev)
}

func Test_Parse_Collection_NOK(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		expectedErr string
	}{
		{
			"CMW monad instead of collection",
			`83781d6170706c69636174696f6e2f7369676e65642d636f72696d2b63626f724dd901f6d28440a044d901f5a04003`,
			`evidence is not a CMW collection`,
		},
		{
			"Invalid CMW CBOR header",
			`63781d6170706c69636174696f6e2f7369676e65642d636f72696d2b63626f724dd901f6d28440a044d901f5a04003`,
			"unable to unmarshal CMW collection: unknown start symbol for CMW",
		},
	}

	for _, tt := range tests {
		var cmwParser cmwParser
		cmw, err := hexDecode(tt.data)
		require.Nil(t, err)

		_, err = cmwParser.Parse(cmw)
		assert.ErrorContains(t, err, tt.expectedErr)
	}
}

func Test_SupportedMediaTypes(t *testing.T) {
	expected := []string{"application/cmw+cbor", "application/cmw+json"}

	var cmwParser cmwParser
	mtList := cmwParser.SupportedMediaTypes()

	assert.ElementsMatch(t, expected, mtList)
}

func hexDecode(s string) ([]byte, error) {
	// allow a long hex string to be split over multiple lines (with soft or
	// hard tab indentation)
	m := regexp.MustCompile("[ \t\n]")
	s = m.ReplaceAllString(s, "")

	data, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func mustReadFile(t *testing.T, fname string) []byte {
	b, err := os.ReadFile(fname)
	require.NoError(t, err)
	return b
}
