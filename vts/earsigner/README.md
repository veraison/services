## Configuration

- `ear-signer`: stanza containing the configuration details about the attestation
  result signing process.  The supported directives are:
  - `alg`: the [JWS algorithm](https://www.iana.org/assignments/jose/jose.xhtml#web-signature-encryption-algorithms)
    used for signing, e.g.: `ES256`, `RS512`.
  - `key`: file containing the private key to be used with `alg`.
    The key is in [JWK format](https://datatracker.ietf.org/doc/rfc7517/).
