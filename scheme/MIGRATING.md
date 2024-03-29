This is a guide for converting plugins written for an earlier version of
Veraison to the updated plugin framework.

1. Remove `GetFormat()`. Attestation formats are no longer required --  the
   combination of the plugin scheme (see (3) below) and media types is now used
   to select plugins.
2. Update `GetName()`. If this was previously implemented in terms of the
   format. This can now be defined as a string constant.
3. Add `GetAttestationScheme()` method. This should return a `string` naming
   the attestation scheme you're implementing.

   > **Warning**: it is important that both `IEndorsementHandler` and
   > `IEvidinceHandler` implementations for a scheme return the same values for
   > this -- it is the connection that allows the verification flow to
   > correctly identify relevant endorsements provisioned via the provisioning
   > flow.

4. Replace `main()`. Instead of using HashiCorp APIs directly, Veraison now
   provides a couple of methods to handle plugin creation.

   Instead of some thing like

   ```go
   package main

    import "github.com/hashicorp/go-plugin"

    type Scheme struct {}

    // ...
    // Scheme implementation goes here
    // ...

    func main() {
            var handshakeConfig = plugin.HandshakeConfig{
                    ProtocolVersion:  1,
                    MagicCookieKey:   "VERAISON_PLUGIN",
                    MagicCookieValue: "VERAISON",
            }

            var pluginMap = map[string]plugin.Plugin{
                    "scheme": &scheme.Plugin{
                            Impl: &Scheme{},
                    },
            }

            plugin.Serve(&plugin.ServeConfig{
                    HandshakeConfig: handshakeConfig,
                    Plugins:         pluginMap,
            })
    }
   ```

   You now need to do:

   ```go
    package main

    import (
            "github.com/veraison/services/handler"
            "github.com/veraison/services/plugin"
    )

    type Scheme struct {}

    // ...
    // Scheme implementation goes here
    // ...

    func main() {
            handler.RegisterEvidenceHandler(&Scheme{})
            // note the change to the import of "plugin" above
            plugin.Serve()
    }
   ```

The above example is for the verification-side "scheme" plugins (note: these
are now known as "evidence handlers" within Veraison code base, but the only
API change is the removal of `GetFormat()` mentioned above.

Identical changes would also need to be made to the provisioning-side "handler"
plugins (now known as "endorsement handlers" within Veraison code base), with
the only difference that `handler.RegisterEndorsementwHandler()` should be used
instead inside `main()`.

It is now also possible to build both plugins into a single binary, simply by
registering both implementations:

```go
func main() {
        handler.RegisterEvidenceHandler(&Scheme{})
        handler.RegisterEndorsementHandler(&Decoder{})
        plugin.Serve()
}
```
