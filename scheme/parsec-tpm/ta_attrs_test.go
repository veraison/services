package parsec_tpm

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/corim/comid"
)

func TestMakeTaAttrs_WithVendorAndModel(t *testing.T) {
	id := ID{
		class: "test-class",
		instance: "test-instance",
	}

	key := &comid.CryptoKey{
		Parameters: map[string]interface{}{
			"vendor": "Vendor Inc.",
			"model": "TPM-2000",
		},
	}

	attrs, err := makeTaAttrs(id, key)
	assert.NoError(t, err)

	var taAttr TaAttr
	err = json.Unmarshal(attrs, &taAttr)
	assert.NoError(t, err)

	assert.Equal(t, "test-class", *taAttr.ClassID)
	assert.Equal(t, "test-instance", *taAttr.InstID)
	assert.Equal(t, "Vendor Inc.", *taAttr.Vendor)
	assert.Equal(t, "TPM-2000", *taAttr.Model)
}

func TestMakeTaAttrs_WithoutVendorAndModel(t *testing.T) {
	id := ID{
		class: "test-class",
		instance: "test-instance",
	}

	key := &comid.CryptoKey{
		Parameters: map[string]interface{}{},
	}

	attrs, err := makeTaAttrs(id, key)
	assert.NoError(t, err)

	var taAttr TaAttr
	err = json.Unmarshal(attrs, &taAttr)
	assert.NoError(t, err)

	assert.Equal(t, "test-class", *taAttr.ClassID)
	assert.Equal(t, "test-instance", *taAttr.InstID)
	assert.Nil(t, taAttr.Vendor)
	assert.Nil(t, taAttr.Model)
}

func TestMakeTaAttrs_WithInvalidTypes(t *testing.T) {
	id := ID{
		class: "test-class",
		instance: "test-instance",
	}

	testCases := []struct {
		name       string
		vendorVal  interface{}
		modelVal   interface{}
		wantVendor bool
		wantModel  bool
	}{
		{
			name:       "number values",
			vendorVal:  123,
			modelVal:   456,
			wantVendor: false,
			wantModel:  false,
		},
		{
			name:       "array values",
			vendorVal:  []string{"vendor1", "vendor2"},
			modelVal:   []string{"model1", "model2"},
			wantVendor: false,
			wantModel:  false,
		},
		{
			name:       "mixed valid and invalid",
			vendorVal:  "Valid Vendor",
			modelVal:   []int{1, 2, 3},
			wantVendor: true,
			wantModel:  false,
		},
		{
			name:       "nil values",
			vendorVal:  nil,
			modelVal:   nil,
			wantVendor: false,
			wantModel:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key := &comid.CryptoKey{
				Parameters: map[string]interface{}{},
				Value:     &comid.TaggedPKIXBase64Key{PK: "test-key-data"},
			}
			if tc.vendorVal != nil {
				key.Parameters["vendor"] = tc.vendorVal
			}
			if tc.modelVal != nil {
				key.Parameters["model"] = tc.modelVal
			}

			attrs, err := makeTaAttrs(id, key)
			require.NoError(t, err)

			var taAttr TaAttr
			err = json.Unmarshal(attrs, &taAttr)
			require.NoError(t, err)

			// Check mandatory fields
			assert.Equal(t, "test-class", *taAttr.ClassID)
			assert.Equal(t, "test-instance", *taAttr.InstID)
			assert.Equal(t, "test-key-data", *taAttr.VerifKey)

			// Check optional fields
			if tc.wantVendor {
				assert.NotNil(t, taAttr.Vendor)
				assert.Equal(t, tc.vendorVal.(string), *taAttr.Vendor)
			} else {
				assert.Nil(t, taAttr.Vendor)
			}
			if tc.wantModel {
				assert.NotNil(t, taAttr.Model)
				assert.Equal(t, tc.modelVal.(string), *taAttr.Model)
			} else {
				assert.Nil(t, taAttr.Model)
			}
		})
	}
}

func TestMakeTaAttrs_WithSpecialCases(t *testing.T) {
	id := ID{
		class:    "test-class",
		instance: "test-instance",
	}

	testCases := []struct {
		name       string
		vendor    string
		model     string
		wantError bool
	}{
		{
			name:    "special characters",
			vendor: "Vendor & Co., Ltd.",
			model:  "TPM-2000!@#$%",
			wantError: false,
		},
		{
			name:    "empty strings",
			vendor: "",
			model:  "",
			wantError: false,
		},
		{
			name:    "very long strings",
			vendor: "Very Long Vendor Name That Could Potentially Cause Issues " + 
			        "In Some Systems That Have Length Limitations For These Fields " +
			        "We Should Handle This Gracefully",
			model:  "TPM-" + strings.Repeat("0", 1000),
			wantError: false,
		},
		{
			name:    "unicode characters",
			vendor: "製造元株式会社",  // Japanese for "Manufacturer Co., Ltd."
			model:  "TPM-モデル2000", // Japanese for "TPM-Model2000"
			wantError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key := &comid.CryptoKey{
				Parameters: map[string]interface{}{
					"vendor": tc.vendor,
					"model":  tc.model,
					"extra1": "should be ignored",
					"extra2": 123,
				},
				Value: &comid.TaggedPKIXBase64Key{PK: "test-key-data"},
			}

			attrs, err := makeTaAttrs(id, key)
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			var taAttr TaAttr
			err = json.Unmarshal(attrs, &taAttr)
			require.NoError(t, err)

			// Verify fields
			assert.Equal(t, "test-class", *taAttr.ClassID)
			assert.Equal(t, "test-instance", *taAttr.InstID)
			assert.Equal(t, tc.vendor, *taAttr.Vendor)
			assert.Equal(t, tc.model, *taAttr.Model)
		})
	}
}

func TestMakeTaAttrs_WithPEMKey(t *testing.T) {
	// Sample RSA public key in PEM format
	const pemKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA3Tz2mr7SZiAMfQyuvBjM
9OiJjRazXBZ1BUiFOkWM83wGv2PkPA/mPfH9jqrLdQJLxG/Nkw3F0w5nHVPljNB1
X1YzQR9G7qYPqhXFOZE/m2tgL1qzrijZZrFmNR94EJj1N/7UHqkZZaZWprgq7kHx
HnDM8JD/4kZ0QPQv0tHU6paG0nfxlRmeyiWx7zHZZz7QXqlQzG65KOA0XgZc0PO6
TiwgQ8CFPLd2yOAKR1eVXnUZ
-----END PUBLIC KEY-----`
	
	id := ID{
		class: "test-class",
		instance: "test-instance",
	}

	key := &comid.CryptoKey{
		Parameters: map[string]interface{}{
			"vendor": "Vendor Inc.",
			"model": "TPM-2000",
		},
		Value: &comid.TaggedPKIXBase64Key{PK: base64.StdEncoding.EncodeToString([]byte(pemKey))},
	}

	attrs, err := makeTaAttrs(id, key)
	require.NoError(t, err)

	var taAttr TaAttr
	err = json.Unmarshal(attrs, &taAttr)
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, "test-class", *taAttr.ClassID)
	assert.Equal(t, "test-instance", *taAttr.InstID)
	assert.Equal(t, "Vendor Inc.", *taAttr.Vendor)
	assert.Equal(t, "TPM-2000", *taAttr.Model)
	assert.NotEmpty(t, *taAttr.VerifKey)

	// Verify the key is properly encoded
	_, err = base64.StdEncoding.DecodeString(*taAttr.VerifKey)
	require.NoError(t, err)
}