## Configuration

Logging configuration is specified under top-level entry `logging`.

### `logging` configuration

- `level` (optional): specify the minimum enabled logging level. Supported
  values (from lowest to highest) are: `debug`, `info`, `warn`, `error`.
  Defaults to `info`.
- `format` (optional): specify the format of the log records (e.g. which
  entries appear in it). Currently supported formats:<br />
  `production`:  default `zap` production config.<br />
  `development`: default `zap` development config.<br />
  `bare`: a relatively bare format, featuring only log level (colored), logger
  name, and its message, suffixed with any fields.
- `development` (optional): set to `true` to put the logger into development
  mode. This changes the behavior of `DPanic` and takes stacktraces more
  liberally.
- `disable-stacktrace` (optional): set to `true` to completely disable
  automatic stacktrace capturing.
- `sampling` (optional): set the sampling policy. There are two possible
  sub-settings: `initial` and `thereafter`, each is an integer value. `initial`
  specifies how many messages, of a given level, will be logged as-is every
  second. After than, only every `thereafter`'th message will be logged within
  that seconds. e.g.
  ```yaml
    sampling:
      initial: 10
      thereafer: 5
  ```
  means, for very level (debug, info, etc) log the first 10 messages that occur
  in 1 second. If more messages occur within the second, log every 5th message
  after the first 10.
- `encoding` (optional): specifies logger encoding. Supported values:
  `console`, `json`.
- `output-paths` (optional): a list of URLs or file paths to write logging
  output to. By default, output is written to `stdout` (which may also be
  specified as a "path" along other locations). In case the same configuration
  is used by multiple services, you can insert `"{{ .service }}"` somewhere in
  paths (part from  `stdout`/`stderr`) to have different services log into
  different files.
- `err-output-paths` (optional): a list of URLs or file paths to write internal
  logger errors to (note: *not* the error-level logging output). Defaults to
  `stderr`. In case the same configuration is used by multiple services, you
  can insert `"{{ .service }}"` somewhere in paths (part from
  `stdout`/`stderr`) to have different services log into different files.
- `initial-fields` (optional): a map of key-value fields to add to
  log records (in addition to those added by specific logging sites).
