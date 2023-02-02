This directory contains packages implementing support of specific attestation
schemes. Currently the following schemes are implemented:

- `cca-ssd-platform` Arm Confidential Compute Architecture attestation.
- `psa-iot`: Arm Platform Security Architecture attestation.
- `tcg-dice`: [TCG
  DICE](https://trustedcomputinggroup.org/work-groups/dice-architectures/)-compatible
  attestation (note: this does not implement any specific DICE architecture).
- `tmp-enacttrust`: TPM-based attestation for
  [EnactTrust](https://www.enacttrust.com/) security cloud.


## Implementing Attestation Scheme Support

> **Note**: If you already have attestation scheme plugins implemented for an
> earlier version of Veraison, please see the [migration guide](MIGRATING.md)
> for how to convert them to the new framework.

Supporting a new attestation scheme requires defining how to provision
endorsements (if any) and how to process evidence tokens. The former is done by
implementing [`IEndorsementDecoder`](../decoder/iendorsementdecoder.go), and the
latter by implementing [`IEvidenceDecoder`](../decoder/ievidencedecoder.go).
Finally, an executable should be created that [registers](../decoder/plugin.go)
and serves them.

```
package main

import (
	"github.com/veraison/services/decoder"
	"github.com/veraison/services/plugin"
)

type MyEvidenceDecoder struct {}

// ...
// Implementation of IEvidenceDecoder for MyEvidenceDecoder
// ...

type MyEndrosementDecoder struct {}

// ...
// Implementation of IEndrosementDecoder for MyEndrosementDecoder
// ...


func main() {
	decoder.RegisterEndorsementDecoder(&MyEndorsementDecoder{})
	decoder.RegisterEvidenceDecoder(&MyEvidenceDecoder{})

	plugin.Serve()
}
```
