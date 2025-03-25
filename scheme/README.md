This directory contains packages implementing support of specific attestation
schemes.

> [!NOTE]
> When adding (or removing) a scheme, please update `../builtin/scheme.gen.go`
> to include the appropriate entries. This can be done automatically using
> `../scripts/gen-schemes` script (see `../builtin/Makefile`) or by manually
> editing the file. The script takes a long time to execute, so unless multiple
> schemes are being added/moved/deleted, manual editing may be easier.

## Current Schemes

Currently the following schemes are implemented:

- `arm-cca` Arm Confidential Compute Architecture attestation.
- `psa-iot`: Arm Platform Security Architecture attestation.
- `riot`: [RIoT based DICE](https://trustedcomputinggroup.org/work-groups/dice-architectures/)-compatible
  attestation (note: this does not implement any specific DICE architecture).
- `tmp-enacttrust`: TPM-based attestation for
  [EnactTrust](https://www.enacttrust.com/) security cloud.
- `parsec-tpm` : Parsec TPM based hardware-backed attestation, details
  [here](https://github.com/CCC-Attestation/attested-tls-poc/blob/main/doc/parsec-evidence-tpm.md)
- `parsec-cca` : Parsec CCA based hardware-backed attestation, details
   [here](https://github.com/CCC-Attestation/attested-tls-poc/blob/main/doc/parsec-evidence-cca.md)

## Implementing Attestation Scheme Support

> [!NOTE]
> If you already have attestation scheme plugins implemented for an
> earlier version of Veraison, please see the [migration guide](MIGRATING.md)
> for how to convert them to the new framework.

Supporting a new attestation scheme requires defining how to provision
endorsements (if any) by implementing
[`IEndorsementHandler`](../handler/iendorsementhandler.go), how to process
evidence tokens by implementing
[`IEvidenceHandler`](../handler/ievidencehandler.go), how to create and obtain
scheme-specific keys used to store and retrieve endorsements and trust anchors
by implementing [`IStoreHandler`](../handler/istorehandler.go), and how to
handle CoSERV queries by implementing
[`ICoservProxyHandler`](../handler/icoservproxyhandler.go).

Finally, an executable should be created that [registers](../handler/plugin.go)
and serves them.

```go
package main

import (
	"github.com/veraison/services/decoder"
	"github.com/veraison/services/plugin"
)

type MyEvidenceHandler struct {}

// ...
// Implementation of IEvidenceHandler for MyEvidenceHandler
// ...

type MyEndrosementHandler struct {}

// ...
// Implementation of IEndrosementHandler for MyEndrosementHandler
// ...

type MyStoreHandler struct {}

// ...
// Implementation of IStoreHandler for MyStoreHandler
// ...

type MyCoservProxyHandler struct {}

// ...
// Implementation of ICoservProxyHandler for MyCoservProxyHandler
// ...

func main() {
	handler.RegisterEndorsementHandler(&MyEndorsementHandler{})
	handler.RegisterEvidenceHandler(&MyEvidenceHandler{})
	handler.RegisterStoreHandler(&MyStoreHandler{})
	handler.RegisterCoservProxyHandler(&MyCoservProxyHandler{})

	plugin.Serve()
}
```

## Debugging

Handler code is a lot easier to debug when it runs as part of the service
processes, rather than as a plugin. This can be achieved by using the "builtin"
plugin loader.

Attestation scheme loading method is a build-time configuration. Since `delve`
does its own building, it will ignore the normal build configuration. Instead,
you will have to configure this when invoking `delve`:

```sh
dlv debug --build-flags "-ldflags '-X github.com/veraison/services/config.SchemeLoader=builtin'"
```

This will allow you to step into and set break points inside scheme code.
