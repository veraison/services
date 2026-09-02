# VTS Endorsement Store

The endorsement store used by VTS consists of a list of store plugins. The
implementation iterates over the list for each store operation, until it
succeeds.

## Configuration
The endorsement store configuration is passed in the `endorsement-store` stanza
of the configuration yaml file.

Example:
```yaml
endorsement-store:
  coserv:
    signer:
      alg: ES256
      key: ./skey.jwk
    max-expiry: 5 mins
  active-plugins:
    - corim-store
  plugin-parameters:
    corim-store:
      dbms: sqlite3
      dsn: file::memory:?cache=shared
      trace-sql: false
```

* `coserv`: contains the CoSERV configuration parameters of the store. These
parameters are passed along with the plugin parameters to each of the plugins
during plugin initialization.

* `active-plugins`: contains the list of store plugins that will be used to
construct the endorsement-store. The plugins will be iterated over in the
order they are given in the config.

* `plugin-parameters`: contains parameters for store plugins.

### Notes

* If the `active-stores` list contains a plugin name which is not loadable,
the service will fail to start.

* If the `active-stores` section is empty or not present, the service will
fail to start.

* **Important**: For submitting endorsements, only the first store in the list
is used, instead of iterating over the list of active stores. If a read-only
store is used as the first store in the list, endorsement provisioning will not
work.
