# Endorsement Store Interface

## Reference Value

```json
{
  "scheme": "PARSEC_TPM",
  "type": "REFERENCE_VALUE",
  "attributes": {
    "parsec-tpm.alg-id": 1,
    "parsec-tpm.class-id": "cd1f0e55-26f9-460d-b9d8-f7fde171787c",
    "parsec-tpm.digest": "h0KPxSKAPTEGXnvOPPA/5HUJZjHl4Hu9eg/eYMTPJcc=",
    "parsec-tpm.pcr": 0
  }
}
```

## Trust Anchor

```json
{
  "scheme": "PARSEC_TPM",
  "type": "VERIFICATION_KEY",
  "attributes": {
    "parsec-tpm.ak-pub": "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAETKRFE_RwSXooI8DdatPOYg_uiKm2XrtT_uEMEvqQZrwJHHcfw0c3WVzGoqL3Y_Q6xkHFfdUVqS2WWkPdKO03uw==",
    "parsec-tpm.class-id": "cd1f0e55-26f9-460d-b9d8-f7fde171787c",
    "parsec-tpm.instance-id": "AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
    "parsec-tpm.vendor": "ACME Corp",
    "parsec-tpm.model": "TPM-2000"
  }
}
```

### Vendor and Model Information

The TPM Trust Anchor can include optional vendor and model information to provide additional context about the TPM manufacturer and specific model. These fields are:

- `parsec-tpm.vendor`: The manufacturer or vendor of the TPM (optional)
- `parsec-tpm.model`: The specific model identifier of the TPM (optional)

Both fields support:
- Unicode characters (e.g., for international vendor names)
- Special characters (allowed in both fields)
- Variable length strings (no artificial length restrictions)

### CORIM Example

To include vendor and model information in your CORIM manifest, add them to the `environment.class` section (following standard CoRIM specification). Here are several examples:

```json
{
  "comid.verification-keys": [
    {
      "environment": {
        "class": {
          "id": {
            "class-id": "cd1f0e55-26f9-460d-b9d8-f7fde171787c"
          },
          "vendor": "ACME Corp",
          "model": "TPM-2000"
        },
        "instance": {
          "instance-id": {
            "type": "ueid",
            "value": "AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
          }
        }
      },
      "key": [{
        "type": "pkix-base64-key",
        "value": "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAETKRFE_RwSXooI8DdatPOYg_uiKm2XrtT_uEMEvqQZrwJHHcfw0c3WVzGoqL3Y_Q6xkHFfdUVqS2WWkPdKO03uw=="
      }]
    }
  ]
}
```

Additional Examples:

1. International Vendor (with Unicode characters):
```json
{
  "comid.verification-keys": [
    {
      "environment": {
        "class": {
          "id": {
            "class-id": "cd1f0e55-26f9-460d-b9d8-f7fde171787c"
          },
          "vendor": "富士通株式会社",
          "model": "FUJITSU TPM 2.0"
        },
        "instance": {
          "instance-id": {
            "type": "ueid",
            "value": "AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
          }
        }
      },
      "key": [{
        "type": "pkix-base64-key",
        "value": "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAETKRFE_RwSXooI8DdatPOYg_uiKm2XrtT_uEMEvqQZrwJHHcfw0c3WVzGoqL3Y_Q6xkHFfdUVqS2WWkPdKO03uw=="
      }]
    }
  ]
}
```

2. Minimal Example (vendor only):
```json
{
  "comid.verification-keys": [
    {
      "environment": {
        "class": {
          "id": {
            "class-id": "cd1f0e55-26f9-460d-b9d8-f7fde171787c"
          },
          "vendor": "Intel Corporation"
        },
        "instance": {
          "instance-id": {
            "type": "ueid",
            "value": "AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
          }
        }
      },
      "key": [{
        "type": "pkix-base64-key",
        "value": "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAETKRFE_RwSXooI8DdatPOYg_uiKm2XrtT_uEMEvqQZrwJHHcfw0c3WVzGoqL3Y_Q6xkHFfdUVqS2WWkPdKO03uw=="
      }]
    }
  ]
}
```

3. Complex Example (with special characters):
```json
{
  "comid.verification-keys": [
    {
      "environment": {
        "class": {
          "id": {
            "class-id": "cd1f0e55-26f9-460d-b9d8-f7fde171787c"
          },
          "vendor": "Company & Co., Ltd.",
          "model": "TPM.v2-Enhanced+"
        },
        "instance": {
          "instance-id": {
            "type": "ueid",
            "value": "AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
          }
        }
      },
      "key": [{
        "type": "pkix-base64-key",
        "value": "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAETKRFE_RwSXooI8DdatPOYg_uiKm2XrtT_uEMEvqQZrwJHHcfw0c3WVzGoqL3Y_Q6xkHFfdUVqS2WWkPdKO03uw=="
      }]
    }
  ]
}
```

### Security Considerations

When using vendor and model fields:

1. **Input Validation**:
   - Maximum length: 1024 characters
   - Strings are trimmed of leading/trailing whitespace
   - Basic sanitization is applied to prevent injection attacks
   - Control characters are removed (except newline and tab)

2. **Storage**:
   - Fields are optional and won't affect TPM validation
   - Unicode characters are preserved for international vendor names
   - Dangerous characters are escaped to prevent injection

3. **Best Practices**:
   - Use consistent vendor/model identifiers
   - Prefer official vendor names and model numbers
   - Keep strings concise and meaningful
   - Test with various character encodings if using international names

Note: The vendor and model fields are always optional and are meant for informational purposes only. The TPM's security validation is based solely on its cryptographic identity and measurements.
```
```
