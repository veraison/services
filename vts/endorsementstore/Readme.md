# VTS Endorsement Store

The endorsement store used by VTS consists of a list of store plugins. The
implementation iterates over the list for each store operation, until it
succeeds.

## Configuration
The list of store plugins to be used in the store is configured in the
`active-stores` section of the VTS configuration. The plugins are used
in the order they are listed in the configuration.

Example:
```yaml
vts:
  active-stores:
    - corim-store
    - store1
    - store2
```

### Notes

* If the `active-stores` list contains a plugin name which is not loadable,
the service will fail to start.

* If the `active-stores` section is empty or not present, the service will
fail to start.

* For submitting endorsements, only the first store in the list is used,
instead of iterating over the list of active stores.
