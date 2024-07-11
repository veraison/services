This directory contains specs (or "templates" in `cocli` parlance) for
endorsements to be provisioned to Veraison. There are represented as CoMID's
that associate the endorsements with metadata about the device instance to
which the endorsements pertain, and the endorsing entities.

The CoMID's are not provisioned directly, but are packaged inside a CoRIM (Concise
Reference Integrity Manifest). A single CoRIM can contain multiple CoMID's
(possibly pertaining to multiple modules) along with validity
constraints and metadata pertaining to the provisioning entity.

For more information on CoMID and CoRIM see
[draft-birkholz-rats-corim-03](https://www.ietf.org/archive/id/draft-birkholz-rats-corim-03.html#name-triples)

## Naming convention

All CoMID files must follow the following naming pattern:

    comid-<scheme>-<name>.json

Where `scheme` is the name of the attestation scheme, and `name` is used to
identify this specific set of values (effectively, a test case).

All CoRIM files must follow the following naming pattern:

    corim-<scheme>-<name>.json

Where `scheme` is the name of the attestation scheme, and `name` is used to
identify this specific set of values (effectively, a test case).

See also the endorsements definitions inside the [test
variables](../../tests/common.yaml)
