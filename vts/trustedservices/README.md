## Configuration

Configuration for Veraison Trusted Services is specified under `vts` top-level
entry.

### `vts` configuration

- `server-addr` (optional): address of the VTS server in the form
  `<host>:<port>`. If not specified, this defaults to `127.0.0.1:50051`. Unless
  `listen-addr` is specified (see below), VTS server will extract the port to
  listen on from this setting (but will listen on all local interfaces)
- `listen-addr` (optional): The address the VTS server will listen on in the
  form `<host>:<port>`. Only specify this if you want to restrict the server to
  listen on a particular interface; otherwise, the server will listen on all
  interfaces on the port specified in `server-addr`.
- `tls` (optional): specifies whether TLS should be used for client
  connections. Defaults to `true`.
- `cert`: path to the file containing the certificate that should be
  used by the server if `tls` (see above) is `true`.
- `cert-key`: path to the file containing the key associated with the
  certificate specified by `server-cert` (see above).
- `ca-certs` (optional): a list of paths to certificates that will be used
  in addition to system certs during mutual validation with the client when
  `tls` (see above) is `true`.
