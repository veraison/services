## Configuration

`management-service` is expecting to find the following top-level entries in
configuration:

- `management`: management service configuration. See [below](#management-service-configuration).
- `po-store`: policy store configuration. See [kvstore config](/kvstore/README.md#Configuration).
- `po-agent` (optional): policy agent configuration. See [policy config](/policy/README.md#Configuration).
- `plugin`: plugin manager configuration. See [plugin config](/vts/pluginmanager/README.md#Configuration).
- `logging` (optional): Logging configuration. See [logging config](/vts/log/README.md#Configuration).
- `auth` (optional): API authentication and authorization mechanism
  configuration. If this is not specified, the `passthrough` backend will be
  used (i.e. no authentication will be performed). With other backends,
  authorization is based on `manager` role. See [auth
  config](/auth/README.md#Configuration).

### Management service configuration

- `listen-addr` (optional): the address, in the form `<host>:<port>` the
  management server will be listening on. If not specified, this defaults to
  `localhost:8088`.

### Example

```yaml
management:
  listen-addr: 0.0.0.0:8088
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
```
