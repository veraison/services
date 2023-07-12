This directory contains packages implementing support of specific attestation
schemes. Currently the following schemes are implemented:

- `cca-ssd-platform` Arm Confidential Compute Architecture attestation.
- `psa-iot`: Arm Platform Security Architecture attestation.
- `tcg-dice`: [TCG
  DICE](https://trustedcomputinggroup.org/work-groups/dice-architectures/)-compatible
  attestation (note: this does not implement any specific DICE architecture).
- `tmp-enacttrust`: TPM-based attestation for
  [EnactTrust](https://www.enacttrust.com/) security cloud.
- `parsec-tpm` : Parsec TPM based hardware-backed attestation, details
  [here](https://github.com/CCC-Attestation/attested-tls-poc/blob/main/doc/parsec-evidence-tpm.md)
- `parsec-cca` : Parsec CCA based harware-backed attestation, details
   [here](https://github.com/CCC-Attestation/attested-tls-poc/blob/main/doc/parsec-evidence-cca.md)


## Implementing Attestation Scheme Support

> **Note**: If you already have attestation scheme plugins implemented for an
> earlier version of Veraison, please see the [migration guide](MIGRATING.md)
> for how to convert them to the new framework.

Supporting a new attestation scheme requires defining how to provision
endorsements (if any) and how to process evidence tokens. The former is done by
implementing [`IEndorsementHandler`](../decoder/iendorsementdecoder.go), and the
latter by implementing [`IEvidenceHandler`](../decoder/ievidencedecoder.go).
Finally, an executable should be created that [registers](../decoder/plugin.go)
and serves them.

```
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


func main() {
	handler.RegisterEndorsementHandler(&MyEndorsementHandler{})
	handler.RegisterEvidenceHandler(&MyEvidenceHandler{})

	plugin.Serve()
}
```

## Debugging

Handler code is a lot easier to debug when it runs as part of the service
processes, rather than as a plugin. This can be achieved by using the "builtin"
plugin loader.

Attestation scheme loading method is a build-time configuration. Since `devle`
does its own building, it will ignore the normal build configuration. Instead,
you will have to configure this when invoking `delve`:

```sh
dlv debug --build-flags "-ldflags '-X github.com/veraison/services/config.SchemeLoader=builtin'"
```

This will allow you to step into and set break points inside scheme code.
