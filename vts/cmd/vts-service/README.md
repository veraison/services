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
    datasource: en-store.sql
po-agent:
    backend: opa
plugin:
  backend: go-plugin
  go-plugin:
    folder: ../../plugins/bin/
vts:
  server-addr: 127.0.0.1:50051
```
