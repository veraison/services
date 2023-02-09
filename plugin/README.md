# Veraison Plugin Framework

> **Note**: This section is primarily of interest to core Veraison
> developers. [see here](../scheme/README.md) if you wish to learn how to
> implement plugins to support new attestation schemes.

Veraison services functionality can be extended via plugins. Extension points
are defined via interfaces that embed `plugin.IPluggable` interface (which defines
functionality common to all plugins. `plugin.IPluggable` implementations are
managed via a `plugin.IManager[I IPluggable]`, which allows obtaining plugin
implementations either by name or media type.

Veraison plugin handling is implemented on top of HashiCorp [go-plugin
framework](https://github.com/hashicorp/go-plugin).

## Defining new plugin types

All Veraison plugins must implement the `IPluggable` interface:

```go

type IPluggable interface {
	GetName() string
	GetAttestationScheme() string
	GetSupportedMediaTypes() []string
}

```

A plugin implementation must have a unique name, and it may also advertise the
media types it can handle.

> **Warning**: Currently, Veraison doesn't support multiple plugins advertising
> the same media type. If a manager (see below) discovers a plugin that
> advertises a media type that has already been registered, it will return an
> error.

In order to define new extension points (plugin types) for Veraison, one needs
to define a new interface embedding [`plugin.IPluggable`](./ipluggable.go), and
provide an RPC channel for that interface.

```go
package myplugin

import (
    "net/rpc"
    "github.com/veraison/plugin"
)

type IMyPlugin interface {
    plugin.IPluggable

    Method1()
}

var MyPluginRPC = &plugin.RPCChannel[IMyPlugin]{
    GetClient: func(c *rpc.Client) interface{} { return &PRCClient{client: c } },
    GetServer: func(i IMyPlugin) interface{} { return &PRCServer{Impl: i } },
}

type PRCClient {
    client *rpc.Client
}

type RPCServer struct {
    Impl IMyPlugin
}

// ---8<----------------------------------------------------------
// IMyPlugin implementations for RPCClient and RPCServer go below.
// ...
```

HashiCorp framework supports implementing RPC using `net/rpc` or
[gRPC](https://grpc.io/). [See
here](https://pkg.go.dev/github.com/hashicorp/go-plugin) for further details.

## Creating plugins

> :information_source: For the specifics of how to create implementations of the already
> defined `IEndorsmentHandler` and `IEvidinceHandler` Veraison plugins, [see
> here](../handler/README.md).

Plugins are created by implementing the corresponding interface, registering
this implementation under an appropriate name (matching the name the manager
will look for on discovery -- see below), and serving the plugin inside
`main()`:


```go
package main

import (
    "github.com/veraison/plugin"

    "myplugin" // see above
)

type Impl struct {}

func (o *Impl) GetName() string { return "my-implementation" }
func (o *Impl) GetAttestationScheme() string { return "my-scheme" }
func (o *Impl) GetSupportedMediaTypes() []string { return []string{"text/html"} }
func (o *Impl) Method1() {}

func main() {
    // "my-plugin" should match what the manager is looking for -- see below
    plugin.RegisterImplementation("my-plugin", &Impl{}, myplugin.MyPluginRPC)
    plugin.Serve()
}
```

## Discovering and using plugins

In short, you create a manager using `plugin.CreateGoPluginManager`, specifying
which plugins you want it to manage, and where to look for them, and then
look up your plugin either via it's name or one of the media types it supports.

```go
package main

import (
    "github.com/veraison/config"
    "github.com/veraison/log"
    "github.com/veraison/plugin"

    "myplugin" // see above
)

func main() {

    // Read configuration form a file using github.com/spf13/viper
    v, err := config.ReadRawConfig("", false)
    if err != nil {
            log.Fatalf("Could not read config sources: %v", err)
    }

    // Extract the sub-secition of config related to plugin management
    subs, err := config.GetSubs(v, "plugin")
    if err != nil {
            log.Fatal(err)
    }

    pluginManager, err := plugin.CreateGoPluginManager(
            subs["plugin"], log.Named("plugin"),
            // plugins must register themselves with type "my-plugin" -- see
            // above.
            "my-plugins", myplugin.MyPluginRPC)

    var impl myplugin.IMyPlugin

    // look up by name...
    imp, err = pluginManager.LookupByName("my-implementation")
    if err != nil {
            log.Fatal(err)
    }

    // ...or, alternatively, by media type
    imp, err = pluginManager.LookupByMediaType("test/html")
    if err != nil {
            log.Fatal(err)
    }

    // use your plugin
    impl.Method1()
}
```
