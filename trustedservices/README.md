## Configuration

### Signer Configuration

- `ear-signer`: stanza containing the configuration details about the attestation
  result signing process.  The supported directives are:
  - `alg`: the [JWS algorithm](https://www.iana.org/assignments/jose/jose.xhtml#web-signature-encryption-algorithms)
    used for signing, e.g.: `ES256`, `RS512`.
  - `key`: file containing the private key to be used with `alg`.
    The key is in [JWK format](https://datatracker.ietf.org/doc/rfc7517/).

### Server Configuration

- `server-addr` (optional): address of the VTS server in the form
  `<host>:<port>`. If not specified, this defaults to `127.0.0.1:50051`.
  Unless `listen-addr` is specified (see below), VTS server will extract the
  port to listen on from this setting (but will listen on all local interfaces)
- `listen-addr` (optional): The address the VTS server will listen on in the
  form `<host>:<port>`. Only specify this if you want to restrict the server to
  listen on a particular interface; otherwise, the server will listen on all
  interfaces on the port specified in `server-addr`.
