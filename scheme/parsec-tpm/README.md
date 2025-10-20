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

### CoRIM Example

To include vendor and model information in your CoRIM manifest, add them to the `environment.class` section (following standard CoRIM specification):

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
