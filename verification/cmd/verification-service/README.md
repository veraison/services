## Configuration

`verification-services` is expecting to find the following top-level entries in
configuration:

- `verification` (optional): verification service configuration. See [below](#verification-service-configuration).
- `verifier` (optional): verifier configuration. See [below](#verifier-configuration).
- `vts` (optional): Veraison Trusted Services backend configuration. See [trustedservices config](/vts/trustedservices/README.md#Configuration).
- `logging` (optional): Logging configuration. See [logging config](/vts/log/README.md#Configuration).
- `sessionmanager` (optional): Session manager backend configuration. See [below](#session-manager-configuration)

### Verification service configuration

- `listen-addr` (optional): the address, in the form `<host>:<port>` the verification
  server will be listening on. If not specified, this defaults to
  `localhost:8080`.
- `protocol` (optional): the protocol that will be used. Defaults to "https" if not specified. Must be either "http" or "https".
- `cert`: path to the x509 certificate to be used. Must be specified if protocol is "https"
- `cert-key`: path to the key associated with the certificate specified in `cert`. Must be specified if protocol is "https"

### Verifier configuration

The verifier currently doesn't support any configuration.

### Session manager configuration

Session manager has a single configuration point: `backend`. This specifies
which `ISessionManager` implementation will be used. The following backends are
supported:

- `ttlcache`: the default; this creates the session cache in memory of the
  `verification-service` process.
- `memcached`: uses an external [memcached](https://www.memcached.org/) server.

All other entries under `sessionmanager` must be backend names (i.e. `ttlcache`
or `memcached`), providing backend-specific configuration. Only configuration
for the backend selected by `backend` entry will actually be used.

#### `ttlcache` backend

`ttlcache` backend does not have any configuration points.

#### `memcached` backend

`memcached` backend has the following configuration points:

- `servers` (optional): a list of servers in "<host>:<port>" format that will
  be used by the backend. The servers will be used with equal weight. Adding
  the same entry multiple times increases its weight. If this is not specified,
  it will default to `["localhost:11211"]`.

### Config files

There are two config files in this directory:

- `config.yaml` is designed to be used when running `verification-service`
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
verification:
  listen-addr: localhost:8443
  protocol: https
  cert: verification.crt
  cert-key: verification.key
vts:
  server-addr: 127.0.0.1:50051
sessionmanager:
  backend: memcached
  memcached:
    servers:
      - localhost:11211
```
