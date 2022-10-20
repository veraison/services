## Configuration

`provisioning-services` is expecting to find the following top-level entries in
configuration:

- `provisoning`: provisioning service configuration. See [below](#provisioning-service-configuration).
- `vts` (optional): Veraison Trusted Services backend configuration. See [trustedservices config](/vts/trustedservices/README.md#Configuration).
- `logging` (optional): Logging configuration. See [logging config](/vts/log/README.md#Configuration).

### Provisioning service configuration

- `plugin-dir`: directory that will be searched for plugin binaries.
- `listen-addr` (optional): the address, in the form `<host>:<port>` the provisioning
  server will be listening on. If not specified, this defaults to
  `localhost:8888`.

### Example

```yaml
provisioning:
  plugin-dir: ../../plugins/bin/
  listen-addr: localhost:8888
vts:
  server-addr: 127.0.0.1:50051
```
