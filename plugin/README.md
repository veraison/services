# Veraison Plugin Framework

Veraison services functionality can be extended via plugins. Extension points
are defined via interfaces that embed `plugin.IPluggable` interface (which defines
functionality common to all plugins. `plugin.IPluggable` implementations are
managed via a `plugin.IManager[I IPluggable]`, which allows obtaining plugin
implementations either by name or MIME type.

## Defining new extensions points

In order to define new extension points (plugin types) for Veraison, one needs
only to define a new interface embedding `pluign.IPluggable`

```go
```

