## Configuration

`verification-services` is expecting to find the following top-level entries in
configuration:

- `verification` (optional): verification service configuration. See [below](#verification-service-configuration).
- `verifier` (optional): verifier configuration. See [below](#verifier-configuration).
- `vts` (optional): Veraison Trusted Services backend configuration. See [trustedservices config](/vts/trustedservices/README.md#Configuration).
- `logging` (optional): Logging configuration. See [logging config](/vts/log/README.md#Configuration).

### Verification service configuration

- `listen-addr` (optional): the address, in the form `<host>:<port>` the verification
  server will be listening on. If not specified, this defaults to
  `localhost:8080`.
- `protocol` (optional): the protocol that will be used. Defaults to "https" if not specified. Must be either "http" or "https".
- `cert`: path to the x509 certificate to be used. Must be specified if protocol is "https"
- `cert-key`: path to the key associated with the certificate specified in `cert`. Must be specified if protocol is "https"

### Verifier configuration

The verifier currently doesn't support any configuration.

### Config files

There are two config files in this directory:

- `config.yaml` is designed to be used when running `verification-service`
  directly from this directory on the build host (i.e. outside docker).
- `config-docker.yaml` this is the file that is designed to be used when running
  inside the debug docker container. See [debugging docker
  deployment](/deployments/docker/README.md#Debugging). The `debug` command
  inside the debug container will automatically use it. If running the
  executable directly inside docker shell, rather than via the command, this
  file will need to be specified with `--config` option.

### Example

```yaml
verification:
  listen-addr: localhost:8888
  protocol: https
  cert: verification.crt
  cert-key: verification.key
vts:
  server-addr: 127.0.0.1:50051
```
