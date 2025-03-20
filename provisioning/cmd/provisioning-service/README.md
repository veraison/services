## Configuration

`provisioning-services` is expecting to find the following top-level entries in
configuration:

- `provisioning`: provisioning service configuration. See [below](#provisioning-service-configuration).
- `vts` (optional): Veraison Trusted Services backend configuration. See [trustedservices config](/vts/trustedservices/README.md#Configuration).
- `logging` (optional): Logging configuration. See [logging config](/vts/log/README.md#Configuration).
- `auth` (optional): API authentication and authorization mechanism
  configuration. If this is not specified, the `passthrough` backend will be
  used (i.e. no authentication will be performed). With other backends,
  authorization is based on `provisioner` role. See [auth
  config](/auth/README.md#Configuration).

### `provisioning` configuration

- `listen-addr` (optional): the address, in the form `<host>:<port>` the provisioning
  server will be listening on. If not specified, this defaults to
  `localhost:8443`.
- `protocol` (optional): the protocol that will be used. Must be either "http" or "https". Defaults to "https" if not specified.
- `cert`: path to the x509 certificate to be used. Must be specified if protocol is "https"
- `cert-key`: path to the key associated with the certificate specified in `cert`. Must be specified if protocol is "https"

### Config files

There are two config files in this directory:

- `config.yaml` is designed to be used when running `provisioning-service`
  directly from this directory on the build host (i.e. outside docker). This is
  no longer supported (use the [native
  deployment](../../../deployments/native/README.md) instead). It is kept for
  illustrative purposes only.
- `config-docker.yaml` this is the file that is designed to be used when running
  inside the debug docker container. See [debugging docker
  deployment](/deployments/docker/README.md#Debugging). The `debug` command
  inside the debug container will automatically use it. If running the
  executable directly inside docker shell, rather than via the command, this
  file will need to be specified with `--config` option.

### Example

```yaml
provisioning:
  listen-addr: localhost:9443
  protocol: https
  cert: provisioning.crt
  cert-key: provisioning.key
vts:
  server-addr: vts-service:50051
```
