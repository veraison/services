## Configuration

`ear-signer`is the stanza containing the configuration details about the
attestation result signing process.  The supported directives are:

### `ear-signer` configuration

- `alg`: the [JWS algorithm](https://www.iana.org/assignments/jose/jose.xhtml#web-signature-encryption-algorithms)
used for signing, e.g.: `ES256`, `RS512`.
- `key`: URL with the location of the private key to be used with `alg`. The
following URL schems are supported:
  - `file`: URL path is the path to a local file
  - `aws`: URL path is in the form `<region>:<secret-name>` where `<region>`
    is an AWS region, and `<secrete-name>` is the name under which the key is
    stored in AWS Secrets Manager.
  If a scheme is not specified, `file` is assumed.
  The key is in [JWK format](https://datatracker.ietf.org/doc/rfc7517/).
