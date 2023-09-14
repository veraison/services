# CoRIM Generation

## Preconditions

>>Note: the below assumes both the [evcli](https://github.com/veraison/evcli) and the [cocli](https://github.com/veraison/corim/tree/main/cocli) tools are installed on the system.

## Installing and configuring

To install the `gen-corim` command, do:

```
$ go install github.com/veraison/services/gen-corim@latest
```

## Usage

```
$ gen-corim psa evidence.cbor key.json [--template-dir=templates] [--corim-file=endorsements/output.cbor]
```

On success, you should see something like this printed to stdout:

```
>> generated "endorsements/output.cbor" using "evidence.cbor"
```
### Supplied Arguments
### Attestation Scheme

The attestation scheme to be used. The only attestation schemes supported by this service are `psa` and `cca`.

#### Evidence File

CBOR-encoded evidence token to be used.

### Key File

Public key material needed to verify the evidence. The key file is expected be in [jwk](https://openid.net/specs/draft-jones-json-web-key-03.html) format.

### Template Directory (Optional)

The directory containing the CoMID and CoRIM templates via the `--template-dir` switch (abbrev. `-t`). If this flag is not set the path for the template directory will default to `templates` within the current working directory. The template directory must exist and must contain files named `comid-template.json` and `corim-template.json` which contain the respective templates. Some examples of CoMID and CoRIM JSON templates can be found in the [data/templates](data/templates) folder.

### Output File (Optional)

If you wish to specify the name and path of the produced endorsement then pass this via the `corim-file` switch (abbrev. `-c`). If this flag is not set then the produced endorsement will be saved in the current working directory under the file name `psa-endorsements.cbor` or `cca-endorsements.cbor` depending on the attestation scheme used.
