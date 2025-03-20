This directory contains packages implementing `arm-cca` (Arm Confidential Compute Architecture) attestation scheme.

Arm CCA attestation scheme is a composite attestation scheme which comprises a CCA Platform Attestation & a Realm Attestation.

Endorsement Store Interface for the CCA Platform and Realm Attestation Scheme is given below.

## Endorsement Store Interface

### Arm CCA Platform 

#### Reference Value
```json
{
  "scheme": "ARM_CCA",
  "type": "reference value",
  "subType": "platform.sw-component",
  "attributes": {
    "hw-model": "RoadRunner",
    "hw-vendor": "ACME",
    "impl-id": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
    "measurement-desc": "sha-256",
    "measurement-type": "BL",
    "measurement-value": "BwYFBAMCAQAPDg0MCwoJCBcWFRQTEhEQHx4dHBsaGRg=",
    "signer-id": "BwYFBAMCAQAPDg0MCwoJCBcWFRQTEhEQHx4dHBsaGRg=",
    "version": "3.4.2"
  }
}
{
  "scheme": "ARM_CCA",
  "type": "reference value",
  "subType": "platform.config",
  "attributes": {
    "hw-model": "RoadRunner",
    "hw-vendor": "ACME",
    "impl-id": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
    "platform-config-id": "AQID",
    "platform-config-label": "cfg v1.0.0"
  }
}
```

#### Trust Anchor
```json
{
  "scheme": "ARM_CCA",
  "type": "trust anchor",
  "attributes": {
    "hw-model": "RoadRunner",
    "hw-vendor": "ACME",
    "iak-pub": "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEMKBCTNIcKUSDii11ySs3526iDZ8A\niTo7Tu6KPAqv7D7gS2XpJFbZiItSs3m9+9Ue6GnvHw/GW2ZZaVtszggXIw==\n-----END PUBLIC KEY-----",
    "impl-id": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
    "inst-id": "AQICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIC"
  }
}
```

### Arm CCA Realm  

#### Reference Value

A Realm instance is uniquely identified by the values of Realm initial measurements and Realm Personalization Value (if provided) used to launch a Realm.

```json
{
  "scheme": "ARM_CCA",
  "type": "REFERENCE_VALUE",
  "subType": "realm.reference-value",
  "attributes": {
    "vendor": "Workload Client Ltd",
    "class-id": "CD1F0E55-26F9-460D-B9D8-F7FDE171787C",
    "realm-initial-measurement": "QoS1aUymwNLPR4mguVrIAlyBjeUjBDZL580pgbLS7caFsyInfsJYGZYkE9jJssH1",
    "hash-alg-id": "sha-384",
    "realm-personalization-value": "5Fty9cDAtXLbTY06t+l/No/3TmI0eoJN7LZ6hOUiTXXkW3L1wMC1cttNjTq36X82j/dOYjR6gk3stnqE5SJNdQ==",
    "rem0": "IQe752H8pS2VE2oTVNt6TdV7Gya+DT2nHZ6yOYazS6YVq/ZRTPNeWp6lWgMtBop4",
    "rem1": "JQe752H8pS2VE2oTVNt6TdV7Gya+DT2nHZ6yOYazS6YVq/ZRTPNeWp6lWgMtBop4",
    "rem2": "MQe752H8pS2VE2oTVNt6TdV7Gya+DT2nHZ6yOYazS6YVq/ZRTPNeWp6lWgMtBop4",
    "rem3": "NQe752H8pS2VE2oTVNt6TdV7Gya+DT2nHZ6yOYazS6YVq/ZRTPNeWp6lWgMtBop4"
  }
}
```

#### Trust Anchor

Realms have no explicit Trust Anchor to provision, as they are supplied inline in the Realm attestation token.
