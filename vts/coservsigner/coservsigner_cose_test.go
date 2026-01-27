// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package coservsigner

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testKey = []byte(`{
	"kty":"EC",
	"crv":"P-256",
	"x":"MKBCTNIcKUSDii11ySs3526iDZ8AiTo7Tu6KPAqv7D4",
	"y":"4Etl6SRW2YiLUrN5vfvVHuhp7x8PxltmWWlbbM4IFyM",
	"d":"870MB6gfuTJ4HtUnUvYMyJpr5eUZNP4Bk43bVdj3eAE"
}`)

var testPubKey = []byte(`{
	"kty":"EC",
	"crv":"P-256",
	"x":"MKBCTNIcKUSDii11ySs3526iDZ8AiTo7Tu6KPAqv7D4",
	"y":"4Etl6SRW2YiLUrN5vfvVHuhp7x8PxltmWWlbbM4IFyM"
}`)

var testBadKey = []byte(`{
	"x":"MKBCTNIcKUSDii11ySs3526iDZ8AiTo7Tu6KPAqv7D4",
	"y":"4Etl6SRW2YiLUrN5vfvVHuhp7x8PxltmWWlbbM4IFyM"
}`)

var testSymmetricKey = []byte(`{
	"kty":"oct",
    "alg":"A128KW",
    "k":"GawgguFyGrWKav7AX4VKUg"
}`)

var testRSAKey = []byte(`{
	"kty":"RSA",
    "n":"0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtVT86zwu1RK7aPFFxuhDR1L6tSoc_BJECPebWKRXjBZCiFV4n3oknjhMstn64tZ_2W-5JsGY4Hc5n9yBXArwl93lqt7_RN5w6Cf0h4QyQ5v-65YGjQR0_FDW2QvzqY368QQMicAtaSqzs8KJZgnYb9c7d0zgdAZHzu6qMQvRL5hajrn1n91CbOpbISD08qNLyrdkt-bFTWhAI4vMQFh6WeZu0fM4lFd2NcRwr3XPksINHaQ-G_xBniIqbw0Ls1jF44-csFCur-kEgU8awapJzKnqDKgw",
	"e":"AQAB",
	"d":"X4cTteJY_gn4FYPsXB8rdXix5vwsg1FLN5E3EaG6RJoVH-HLLKD9M7dx5oo7GURknchnrRweUkC7hT5fJLM0WbFAKNLWY2vv7B6NqXSzUvxT0_YSfqijwp3RTzlBaCxWp4doFk5N2o8Gy_nHNKroADIkJ46pRUohsXywbReAdYaMwFs9tv8d_cPVY3i07a3t8MN6TNwm0dSawm9v47UiCl3Sk5ZiG7xojPLu4sbg1U2jx4IBTNBznbJSzFHK66jT8bgkuqsk0GjskDJk19Z4qwjwbsnn4j2WBii3RL-Us2lGVkY8fkFzme1z0HbIkfz0Y6mqnOYtqc0X4jfcKoAC8Q",
	"p":"83i-7IvMGXoMXCskv73TKr8637FiO7Z27zv8oj6pbWUQyLPQBQxtPVnwD20R-60eTDmD2ujnMt5PoqMrm8RfmNhVWDtjjMmCMjOpSXicFHj7XOuVIYQyqVWlWEh6dN36GVZYk93N8Bc9vY41xy8B9RzzOGVQzXvNEvn7O0nVbfs",
	"q":"3dfOR9cuYq-0S-mkFLzgItgMEfFzB2q3hWehMuG0oCuqnb3vobLyumqjVZQO1dIrdwgTnCdpYzBcOfW5r370AFXjiWft_NGEiovonizhKpo9VVS78TzFgxkIdrecRezsZ-1kYd_s1qDbxtkDEgfAITAG9LUnADun4vIcb6yelxk",
	"dp":"G4sPXkc6Ya9y8oJW9_ILj4xuppu0lzi_H7VTkS8xj5SdX3coE0oimYwxIi2emTAue0UOa5dpgFGyBJ4c8tQ2VF402XRugKDTP8akYhFo5tAA77Qe_NmtuYZc3C3m3I24G2GvR5sSDxUyAN2zq8Lfn9EUms6rY3Ob8YeiKkTiBj0",
	"dq":"s9lAH9fggBsoFR8Oac2R_E2gw282rT2kGOAhvIllETE1efrA6huUUvMfBcMpn8lqeW6vzznYY5SSQF7pMdC_agI3nG8Ibp1BUb0JUiraRNqUfLhcQb_d9GF4Dh7e74WbRsobRonujTYN1xCaP6TO61jvWrX-L18txXw494Q_cgk",
	"qi":"GyM_p6JrXySiz1toFgKbWV-JdI3jQ4ypu9rbMWx3rQJBfmt0FoYzgUIZEVFEcOqwemRN81zoDAaa-Bk0KWNGDjJHZDdDmFhW3AN7lI-puxk_mHZGJ11rxyR8O55XLSe3SPmRfKwZI6yU24ZxvQKFYItdldUKGzO6Ia6zTKhAVRU",
	"alg":"RS256",
	"kid":"2011-04-29"
}`)

func makeFS(t *testing.T) afero.Fs {
	fs := afero.NewMemMapFs()

	err := fs.MkdirAll("testdata", 0700)
	require.NoError(t, err)

	err = afero.WriteFile(fs, "testdata/ec256key.jwk", testKey, 0600)
	require.NoError(t, err)

	err = afero.WriteFile(fs, "testdata/ec256key-pub.jwk", testPubKey, 0600)
	require.NoError(t, err)

	err = afero.WriteFile(fs, "testdata/bad-key.json", testBadKey, 0600)
	require.NoError(t, err)

	err = afero.WriteFile(fs, "testdata/symmetric-key.json", testSymmetricKey, 0600)
	require.NoError(t, err)

	err = afero.WriteFile(fs, "testdata/rsa.jwk", testRSAKey, 0600)
	require.NoError(t, err)

	return fs
}

func TestCOSESigner_Init_OK(t *testing.T) {
	cfg := Cfg{
		Key: "testdata/ec256key.jwk",
		Alg: "ES256",
	}
	fs := makeFS(t)

	var o COSESigner
	err := o.Init(cfg, fs)
	assert.NoError(t, err)

	assert.Equal(t, testKey, o.rawkey)
	assert.Equal(t, "ES256", o.Alg.String())
	assert.NotNil(t, o.Signer)
	assert.NotNil(t, o.Key)
}

func TestCOSESigner_Init_KO_not_a_private_key(t *testing.T) {
	cfg := Cfg{
		Key: "testdata/ec256key-pub.jwk",
		Alg: "ES256",
	}
	fs := makeFS(t)

	var o COSESigner
	err := o.Init(cfg, fs)
	assert.EqualError(t, err, "parsing CoSERV signer key: expected private key, got public key")
}

func TestCOSESigner_Init_KO_bad_key(t *testing.T) {
	cfg := Cfg{
		Key: "testdata/bad-key.json",
		Alg: "ES256",
	}
	fs := makeFS(t)

	var o COSESigner
	err := o.Init(cfg, fs)
	assert.EqualError(t, err, "parsing CoSERV signer key: invalid key type from JSON ()")
}

func TestCOSESigner_Init_KO_alg_not_set(t *testing.T) {
	cfg := Cfg{
		Key: "testdata/ec256key.jwk",
	}
	fs := makeFS(t)

	var o COSESigner
	err := o.Init(cfg, fs)
	assert.ErrorContains(t, err, "setting CoSERV signer algorithm")
}

func TestCOSESigner_Init_KO_symmetric_key(t *testing.T) {
	cfg := Cfg{
		Key: "testdata/symmetric-key.json",
		Alg: "A128KW",
	}
	fs := makeFS(t)

	var o COSESigner
	err := o.Init(cfg, fs)
	assert.EqualError(t, err, "parsing CoSERV signer key: expected asymmetric key, got *jwk.symmetricKey")
}

func TestCOSESigner_Init_KO_rsa_key(t *testing.T) {
	cfg := Cfg{
		Key: "testdata/rsa.jwk",
		Alg: "RS256",
	}
	fs := makeFS(t)

	var o COSESigner
	err := o.Init(cfg, fs)
	assert.EqualError(t, err, "creating COSE signer: mapping JWK to crypto.Signer: unsupported key type: RSA")
}
