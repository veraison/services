## Configuration

`provisioning-services` is expecting to find the following top-level entries in
configuration:

- `provisioning`: provisioning service configuration. See [below](#provisioning-service-configuration).
- `vts`: Veraison Trusted Services backend configuration. See [trustedservices config](/vts/trustedservices/README.md#Configuration).

### `provisioning` configuration

- `listen-addr`: the address, in the form `<host>:<port>` the provisioning
  server will be listening on.
- `protocol` (optional): the protocol that will be used. Must be either "http" or "https". Defaults to "https" if not specified.
- `cert`: path to the x509 certificate to be used. Must be specified if protocol is "https"
- `cert-key`: path to the key associated with the certificate specified in `cert`. Must be specified if protocol is "https"

### Example

```yaml
provisioning:
  listen-addr: localhost:8888
  protocol: https
  cert: provisioning.crt
  cert-key: provisioning.key
vts:
  server-addr: 127.0.0.1:50051
```
