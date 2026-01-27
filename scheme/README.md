This directory contains packages implementing support of specific attestation
schemes.

> [!NOTE]
> When adding (or removing) a scheme, please update `../builtin/schemes.go`
> to include the appropriate entries.

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
> `example` sub-directory contains a "template" for new scheme boiler plate.
> You can start implementing a new scheme by copying it and filling in the 
> TODO's.

An implementation of an attestation scheme needs to provide two things:

- A [`SchemeDescriptor`](../handler/schemedescriptor.go) providing basic
  information about the scheme.
- An implementation of
  [`ISchemeImplementation`](../handler/ischemeimplementation.go) interface.

Both of these need to then be registered

- as a plugin by calling `handler.RegisterSchemeImpalementation`, and
- for in-line implementations (i.e. those that are part of this code base), by
  updating [builtin/schemes.go](../builtin/schemes.go)


```go
package main

import (
	"github.com/veraison/services/handler"
	"github.com/veraison/services/plugin"
)

var Descriptor = handler.SchemeDescriptor{
	Name: "MY_SCHEME",
	VersionMajor: 1,
	VersionMinor: 0,
	CorimProfiles: []string{
		"http://my-org/my-scheme",
	},
	EvidenceMediaTypes: []string{
		"application/my-scheme-evidence",
	},
}

type Implementation struct {}

// ...
// Implementation of ISchemeImplementation for Implementation
// ...

func main() {
	handler.RegisterSchemeImplementation(Descriptor,  &Implementation{})
	plugin.Serve()
}
```

### Validating endorsements and trust anchors

Endorsements and trust anchors are provisioned as CoRIMs into the store using
a standard provisioning flow. A scheme implementation does not need to worry
about that except to make sure that the CoRIM provisioned for it contains the
information it expects. 

The recommended way to do that is to utilize the validation support provided by
the CoRIM libraries [profile extensions
mechanism](https://github.com/veraison/corim/blob/main/extensions/README.md).

```go
type TriplesValidator struct {}

func (o *TriplesValidator) ValidTriples(triples *comid.Triples) error {
    // make sure the triples contain the right data for the scheme, returning
    // an error if they don't.

	return nil
}

func init() {
	extMap := extensions.NewMap().Add(comid.ExtTriples, &TriplesValidator{})

    for _, profileString := range Descriptor.CorimProfiles {
        profileID, err := eat.NewProfile(profileString)
        if err != nil {
            panic(err)
        }

        if err := corim.RegisterProfile(profileID, extMap); err != nil {
            panic(err)
        }
    }
}

```

To make this easier, we provide standard `TriplesValidator` implementation that
only needs to be given callback validators for individual elements (such as
environments). E.g.

```go
import (
	"github.com/veraison/corim/comid"
    "github.com/veraison/services/scheme/common"
)

func revValEnvValidator(e *comid.Environment) error {
    // validate reference value environment
    return nil
}

func taEnvValidator(e *comid.Environment) error {
    // validate trust anchor environment
    return nil
}

func measurementsValidator(meas []comid.Measurement) error {
    // validate reference value measurements environment
    return nil
}

func cryptoKeysValidator(keys []*comid.CryptoKeys) error {
    // validate attestation verification keys
    return  nil
}

var validator = &common.TriplesValidator{
    TAEnviromentValidator: taEnvValidator,
    RefValEnviromentValidator: refValEnvValidator,
    // Alternatively to the above, EnvironmentValidator may specified, which
    // will be invoked for both trust anchors and reference values.
    CryptoKeysValidator: measurementsValidator,
    MeasurementsValidator: cryptoKeysValidator,
}

func init() {
	extMap := extensions.NewMap().Add(comid.ExtTriples, validator)

    // ...
}
```
Yo do not need to specify all the validator functions listed above -- just the
ones for the elements you care about for your scheme.

### `ISchemeHander`

VTS actually expects [`ISchemeHander`](../handler/ischemehandler.go) to be
implemented by plugins. [A wrapper](../handler/schemeimplementationwrapper.go)
provides this based on the `SchemeDescriptor` and `ISchemeImplementation`. This
results in a simpler, more declarative scheme definition.

In general, it is recommended that new schemes are defined using
`SchemeDescriptor` and `ISchemeImplementation` as described above. If, for
whatever reason, this is not sufficient, it is possible to implement
`ISchemeHandler` directly:

```go
package main

import (
	"github.com/veraison/services/handler"
	"github.com/veraison/services/plugin"
)

type Handler struct {}

// ...
// Implementation of ISchemeHandler for Handler
// ...

func main() {
	handler.RegisterSchemeHandler(&Handler{})
	plugin.Serve()
}
```

(Note that `ISchemeHandler` embeds `ISchemeImplementation`, so you would still
need to implement it.)

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
