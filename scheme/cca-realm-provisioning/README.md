# Endorsement Store Interface

## Reference Value

In CCA_Realm scheme the Realm instance is uniquely identified by the values of Realm initial measurements,
used to launch a Realm.

```json
{
    "scheme": "CCA_REALM",
    "type": "REFERENCE_VALUE",
    "attributes": {
        "CCA_REALM.vendor": "Workload Client Ltd",
        "CCA_REALM.class-id": "CD1F0E55-26F9-460D-B9D8-F7FDE171787C",
        "CCA_REALM-realm-initial-measurement": "QoS1aUymwNLPR4mguVrIAlyBjeUjBDZL580pgbLS7caFsyInfsJYGZYkE9jJssH1",
        "CCA_REALM.hash-alg-id": "sha-384",
        "CCA_REALM.measurements": [
            {
                "rim": "QoS1aUymwNLPR4mguVrIAlyBjeUjBDZL580pgbLS7caFsyInfsJYGZYkE9jJssH1"
            },
            {
                "rem0": "IQe752H8pS2VE2oTVNt6TdV7Gya+DT2nHZ6yOYazS6YVq/ZRTPNeWp6lWgMtBop4"
            },
            {
                "rem1": "JQe752H8pS2VE2oTVNt6TdV7Gya+DT2nHZ6yOYazS6YVq/ZRTPNeWp6lWgMtBop4"
            },
            {
                "rem2": "MQe752H8pS2VE2oTVNt6TdV7Gya+DT2nHZ6yOYazS6YVq/ZRTPNeWp6lWgMtBop4"
            },
            {
                "rem3": "NQe752H8pS2VE2oTVNt6TdV7Gya+DT2nHZ6yOYazS6YVq/ZRTPNeWp6lWgMtBop4"
            }
        ]
    }
}
```

## Trust Anchor

CCA_Realm scheme has no explicit Trust Anchor to provision, as they are supplied inline in the Realm attestation token