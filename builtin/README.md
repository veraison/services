This package implements "built in" version of plugin loader and manager (see
[Veraison plugin framework](../plugin). This removes runtime plugin discovery
and loading (and, thus, any potential security issues associated with running
external executables).

Instead, plugins are "discovered" at build time by iterating over an array of
plugin implementations defined inside [schemes.go](schemes.go).

> **Note**: When a new plugin is added, [schemes.go](schemes.go) must be
> manually updated.
