## Configuration

`vts-services` is expecting to find the following top-level entries in
configuration:

- `ta-store`: trust anchor store configuration. See [kvstore config](/kvstore/README.md#Configuration).
- `en-store`: endorsements store configuration. See [kvstore config](/kvstore/README.md#Configuration).
- `po-store`: policy store configuration. See [kvstore config](/kvstore/README.md#Configuration).
- `po-agent` (optional): policy agent configuration. See [policy config](/policy/README.md#Configuration).
- `plugin`: plugin manager configuration. See below.
- `vts` (optional): Veraison Trusted Services backend configuration. See [trustedservices config](/vts/trustedservices/README.md#Configuration).
- `logging` (optional): Logging configuration. See [logging config](/vts/log/README.md#Configuration).
- `ear-signer`: Attestation Result signing configuration. See [signer config](/vts/ear-signer/README.md#Configuration).
- `scheme` (optional): Scheme-specific configuration. See below.

### `plugin` configuration

The following directives are currently supported:

- `backend`: specifies which plugin mechanism will be used. Currently the only
  supported plugin backend is `go-plugin`.
- `<backend name>` (in practice, just `go-plugin`): configuration specific to
  the backend.
- `builtin`: configuration used when plugins are disabled and scheme-specific
  functionally is compiled into the VTS service executable. (Currently just a
  place-holder as there is no configuration for the built-in loading mechanism.)
  Note: enabling or disabling of plugins is a build-time option. It is not
  possible to do so via configuration.

#### `go-plugin` backend configuration

- `dir`: path to the directory that will be scanned for plugin executables.

#### `scheme` configuration

`scheme` must be a map of (case insensitive) scheme names to configuration
entries for those schemes. A configuration entry is a map of parameter names to
corresponding parameter values. For example:

```yaml
scheme:
  my_scheme_name:
    foo: 1
    bar: "baz"
```

Parameter values must be either strings, integers or booleans. Bytes values are
also supported as either raw base64-encoded strings, `my_param: "3q2+7w=="`, or
as YAML `!!binary` directives, `my_param: !!binary "3q2+7w=="`.

Valid parameter names, and corresponding expected value types, are
scheme-specific. Please refer to scheme documentation if it exists, otherwise
assume that no parameters are expected (which is the case for most existing
schemes).

Any `scheme` sub-entry that doesn't correspond to a known scheme name will be
ignored.

### Config files

There are two config files in this directory:

- `config.yaml` is designed to be used when running `vts-service` directly form
  this directory. This is no longer supported (use the [native
  deployment](../../../deployments/native/README.md) instead). It is kept for
  illustrative purposes only.
- `config-docker.yaml` this is the file that is designed to be used when running
  inside the debug docker container. See [debugging docker
  deployment](/deployments/docker/README.md#Debugging). The `debug` command
  inside the debug container will automatically use it. If running the
  executable directly, this file will need to be specified with `--config`
  option.

### Example

```yaml
ta-store:
  backend: sql
  sql:
    driver: sqlite3
    datasource: ta-store.sql
en-store:
  backend: sql
  sql:
    driver: sqlite3
    datasource: en-store.sql
po-store:
  backend: sql
  sql:
    driver: sqlite3
    datasource: po-store.sql
po-agent:
    backend: opa
plugin:
  backend: go-plugin
  go-plugin:
    folder: ../../plugins/bin/
vts:
  server-addr: 127.0.0.1:50051
ear-signer:
  alg: ES256
  key: ./skey.jwk
```
