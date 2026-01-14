# Configuration

The verifier's configuration now includes an optional new key, `dispatch-table`.

```yaml
dispatch-table: ./dispatch-table.json
```

This variable points to a local JSON file containing one object for each client type.

```json
{
  "vrsn-local": {
    "type": "veraison-client",
    "url": "https://localhost:8443",
    "insecure": true,
    "ca-certs": [ "../../../deployments/docker/src/certs/rootCA.crt" ],
    "hints": [ "application/vnd.veraison.tsm-report+cbor; provider=arm_cca" ]
  }
}
```

Currently, only the `veraison` client is supported.
Other verifiers (e.g. Intel-TA, NV and Trustee) will be incorporated in future work.
The optional `"hints"` contains a list of evidence formats (as media types) that the downstream verifier is supposed to known how to handle.
If the downstream verifier exposes a discovery interface, the discovered evidence formats (as media types) will be integrated with the hints.
In the event of a conflict between clients, an evidence format introduced via `"hints"` takes precedence over a discovered evidence format.

The configuration stanza is passed down to the composite evidence client plugin with the same name as the value of `"type"`.

# Discovery

The lead verifier adds a new `"composite-evidence-media-types"` array listing all the supported "composition" base[^1] media types to the discovery object.

```json
{
  "ear-verification-key": { /* ... */ },
  "media-types": [ /* ... */ ],
  "version": "...",
  "service-state": "...",
  "api-endpoints": { /* ... */ },

  "composite-evidence-media-types": [
    "application/cmw+cbor",
    "application/cmw+json",
    "application/eat+cwt",
    "application/eat+jwt"
  ]
}
```

The list is discovered through a VTS endpoint.

Note that if the same media type is listed in the `"media-types"` array (likely including some parameters), the dispatch function will prioritise the `"media-types"` entry over the one in `"composite-evidence-media-types"`.

[^1]: By "base", we mean that they do not include any parameters.

## Discussion

It's probably worth exposing the supported media types of the downstream verifier(s), so that the composite attester (or its RP) can see whether their verification request has any chance of succeeding.
