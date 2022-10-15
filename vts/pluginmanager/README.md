## Configuration

Plugin manager supports the following configuration:

- `backend`: specifies the plugin implementation to be used. Currently
  supported backends: `go-plugin`.
- `<backend name>`: an entry with the name of a backend is used to specify
  configuration for that backend. Multiple such entries may exist in a single
  config, but only the one for the backend specified by the `backend` directive
  will be used.

### `go-plugin` backend configuration

- `folder`: The path to the directory that will be searched for the plugin
  binaries.
