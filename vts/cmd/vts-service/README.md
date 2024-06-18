## Configuration

`vts-services` is expecting to find the following top-level entries in
configuration:

- `ta-store`: trust anchor store configuration. See [kvstore config](/kvstore/README.md#Configuration).
- `en-store`: endorsements store configuration. See [kvstore config](/kvstore/README.md#Configuration).
- `po-store`: policy store configuration. See [kvstore config](/kvstore/README.md#Configuration).
- `po-agent` (optional): policy agent configuration. See [policy config](/policy/README.md#Configuration).
- `plugin`: plugin manager configuration. See [plugin config](/vts/pluginmanager/README.md#Configuration).
- `vts` (optional): Veraison Trusted Services backend configuration. See [trustedservices config](/vts/trustedservices/README.md#Configuration).
- `logging` (optional): Logging configuration. See [logging config](/vts/log/README.md#Configuration).
- `ear-signer`: Attestation Result signing configuration. See [signer config](/vts/ear-signer/README.md#Configuration).

### Config files

There are two config files in this directory:

- `config.yaml` is designed to be used when running `vts-service` directly form
  this directory. It assumes that stores have been initialized under `/tmp`
  (running vis `run-vts` script, also in this directory, ensures that). Since
  `config.yaml` is the name the service looks for when loading config, there is
  no need to explicitly specify this file when running from this directory.
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
