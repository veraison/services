## Configuration

`verification-services` is expecting to find the following top-level entries in
configuration:

- `verification` (optional): verification service configuration. See [below](#verification-service-configuration).
- `verifier` (optional): verifier configuration. See [below](#verifier-configuration).
- `vts` (optional): Veraison Trusted Services backend configuration. See [trustedservices config](/vts/trustedservices/README.md#Configuration).
- `logging` (optional): Logging configuration. See [logging config](/vts/log/README.md#Configuration).

### Verification service configuration

- `listen-addr` (optional): the address, in the form `<host>:<port>` the provisioning
  server will be listening on. If not specified, this defaults to
  `localhost:8080`.

### Verifier configuration

The verifier currently doesn't support any configuration.

### Example

```yaml
verification:
  listen-addr: localhost:8888
vts:
  server-addr: 127.0.0.1:50051
```
