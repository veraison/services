# Endorsement Plugin

Endorsement Plugin is an umbrella term for the custom code that handles
extraction and normalisation of reference values and trust anchors for a
specific Attestation Scheme.

The values are extracted from a reference integrity manifest (RIM), e.g., CoRIM,
and normalised into a key-value representation ready to be stored in the
appropriate KVStore.

When designing an Endorsement Plugin, the first thing to decide is the shape of
your KVStore records.  Specifically, the following two things must be defined
for both reference values and trust anchors:

1. The attributes of each record type
1. The lookup keys for locating the records

These design choices are dictated not only by the Endorsement format, but also
by the Evidence format, i.e., they depend on the Attestation Scheme as a whole.
In particular, the lookup keys should be structured so that they can be
assembled entirely from data available in Evidence.

Each KVStore _key_ is expressed as a custom URI (see
[KVStore#uri-format](../kvstore/README.md#uri-format) for the details).

Each KVStore _value_ is a JSON object with some fixed keys and a customisable
payload.  The common structure is as follows:

```json
{
  "scheme": <attestation scheme>
  "type": <record type>
  "attributes": {
    <any defined by the combo scheme/type>
  }
}
```

These _key_ and _value_ effectively represent the _internal interface_ between
Endorsement and Verification plugins of the same Attestation Scheme.

## A Fictitious TPM-based Attestation Scheme

It may be easier to look at a concrete use case to see how these abstract
principles are put into practice.  To this end we will use an imaginary
TPM-based Attestation Scheme called `MY_TPM`, and assume that CoRIM is the RIM
format.  More specifically, the assumption is that trust anchors and reference
values are conveyed as CoMID attestation verification key and reference value
triples respectively.

In `MY_TPM` trust anchors are per-device raw public keys, whilst reference
values are digests associated with specific PCRs that must be the same for all
devices in a _class_ (e.g., the deployed fleet).

### Trust Anchor Design

Trust anchors are unique to a given device - note that this is an arbitrary
choice we made for our `MY_TPM` example, it is not generally true.

The natural way this is expressed in CoMID is by means of an _instance
environment_ made of an _instance_ unique identifier and, optionally, a _class_
identifier.

Each record will contain exactly one key.

#### Normalised Record Layout

The `scheme` is `MY_TPM` and `type` is `VERIFICATION_KEY`.

The `attributes` bag contains the raw public key (`my-tpm.pkey`) alongside any
identification metadata needed to synthesise the lookup key for this record.  In
this case, `my-tpm.instance-id` and `my-tpm.class-id`.  Note that the types of
the attributes depend on their native CoMID representation.

```json
{
  "scheme": "MY_TPM",
  "type": "VERIFICATION_KEY",
  "attributes": {
    "my-tpm.class-id": "<e.g., UUID>",
    "my-tpm.instance-id": "<e.g., UUID>",
    "my-tpm.pkey": "<Base64 URL-safe encoded SPKI>"
  }
}
```

#### Lookup Keys

The lookup key for trust anchor follows the KVStore URI convention (see above),
using the device instance identifier as path:

```
MY_TPM://<tenant-id>/<instance-id>
```

To allow `IEvidenceHandler`'s `SynthKeysFromTrustAnchor()` to correctly
synthesize the KVStore URI for the trust anchor record, the instance identifier
needs to be stashed alongside the endorsement data in the `attributes`.

`SynthKeysFromTrustAnchor()` will use the value of `my-tpm.instance-id` (and
`scheme`) to do that.

### Reference Values Design

Reference Values are shared by all devices of the same class.  Again, while
common this is not necessarily true for all Attestation Schemes.

The natural way this is expressed in CoMID is by means of a _class environment_,
i.e., one with only a _class_ identifier.

Each record will contain exactly one reference value.  Note that, for algorithm
agility reasons, the same PCR may be associated with more than one reference
value.  Each of these will have have their own record.

#### Normalised Record Layout

The `scheme` is `MY_TPM` and `type` is `REFERENCE_VALUE`.

The `attributes` bag contains the PCR index (`my-tpm.pcr`) and the associated
measurement comprising the hash algorithm (`my-tpm.alg-id`) and the actual value
(`my-tpm.digest`).

Any identification metadata needed to synthesise the lookup key for this record
are stashed in the `attributes`.  In this case, `my-tpm.class-id` that, for a
matching record, is normally the same as the one found in the trust anchor case.

Again, the exact types of the attributes depend on their native CoMID
representation.

```json
{
  "scheme": "MY_TPM",
  "type": "REFERENCE_VALUE",
  "attributes": {
    "my-tpm.pcr": "<PCR index>",
    "my-tpm.alg-id": "<hash algorithm identifier>",
    "my-tpm.digest": "<hash value>",
    "my-tpm.class-id": "<must match the one in the trust anchor record>"
  }
}
```

#### Lookup Keys

The lookup key for reference values follows the KVStore URI convention (see
above) using the device class identifier as path:

```
MY_TPM://<tenant-id>/<class-id>
```

To allow `IEvidenceHandler`'s `SynthKeysFromRefValue()` to correctly synthesize
the KVStore URI for the reference value record, the class identifier needs to be
stashed alongside the endorsement data in the `attributes`.

`SynthKeysFromRefValue()` will use the value of `my-tpm.class-id` (and `scheme`)
to do that.

### Endorsement Handler Code

Once the records and their lookup keys are defined one can start implementing
the actual handler code.

As described in [README.md](README.md#implementing-attestation-scheme-support),
the handler code must implement all the methods defined by the
`IEndorsementHandler` interface.

In particular, we want:

* `GetAttestationScheme()` to return the string `"MY_TPM"`
* `GetSupportedMediaTypes` to return the media type(s) that the plugin
  understands
* `GetName()` to return `"unsigned-corim (my TPM profile)"` since we assume RIM
  data to be conveyed in a CoRIM format suitably profiled to precisely describe
  our attester
* `Init()` and `Close()` to just return `nil` - we don't need any special
  initialisation / termination code

The core extraction and normalisation work is carried out by the `Decode()`
method.  Since we assume CoRIM as input format, the
`common.UnsignedCorimDecoder()` method can be used to take care of the base
CoRIM decoding and validation.  Alongside the RIM data, this method takes input
a `common.IExtractor` implementation that provides the profile-specific
extraction and normalisation logics.

We define a `MyTPMExtractor` object to implement the interface.  This amounts to
providing an implementation of the `RefValExtractor()` and `TaExtractor()`
signatures.

#### Reference Value Extractor

The reference value extractor is fed by the `UnsignedCorimDecoder()` with a
Reference Value triple at a time.

The extractor business logics consists of:

* Extracting the class identifier from the target environment
* Then, for each measurement:
  * Extracting the PCR index (assuming it's found in the measurement `Key` of
    type `uint`), and
  * For each digest (there may be one or more):
    * Extracting the algorithm identifier and the actual digest value
    * Creating a `structpb.Struct` map corresponding to the reference value
      "Normalised Record Layout" as described above using the current PCR index
    * Assembling a `proto.Endorsement` record with `Type` set to
      `proto.EndorsementType_REFERENCE_VALUE`, scheme set to `MY_TPM` and
      attributes set to the `structpb.Struct` map just created
    * Appending the record to the list
* Returning the records list to the caller

#### Verification Key Extractor

The trust anchor extractor is fed by the `UnsignedCorimDecoder()` with a
Verification Key triple at a time.

We expect to find exactly one Verification Key in the supplied triple.

The extractor business logics consists of:

* Extracting the instance identifier from the target environment
* Extracting the raw public key from the first (and only) Verification Key
* Creating a `structpb.Struct` map corresponding to the trust anchor "Normalised
  Record Layout" as described above
* Assembling a `proto.Endorsement` record with `Type` set to
  `proto.EndorsementType_VERIFICATION_KEY`, scheme set to `MY_TPM` and
  attributes set to the `structpb.Struct` map just created
* Returning the record
