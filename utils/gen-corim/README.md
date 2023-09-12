# CoRIM Generation

## Installing and configuring

To install the `gen-corim` command, do:

```
$ go install github.com/veraison/services/gen-corim@latest
```

## Usage

```
$ gen-corim --evidence-file=evidence.cbor --attest-scheme=psa --key-file=key.json --template-dir=templates (--corim-file=endorsements/output.cbor)
```

On success, you should see something like this printed to stdout:

```
>> generated "endorsements/output.cbor" using "evidence.cbor"
```
### Supplied Arguments
### Evidence File

Pass the CBOR-encoded evidence token to be used via the `--evidence-file` switch (or equivalently its `-e` shorthand).

### Attestation Scheme

Pass the attestation scheme to be used via the `--attest-scheme` switch (abbrev. `-a`). The only attestation schemes supported by this service are `psa` and `cca`.

### Key File

Pass the -encoded key material needed to verify the evidence via the `--key-file` switch (abbrev. `-k`). The key file is expected be in [jwk](https://openid.net/specs/draft-jones-json-web-key-03.html) format.

### Template Directory

Pass the directory containing the CoMID and CoRIM templates via the `--template-dir` switch (abbrev. `-t`). This directory must exist and must contain files named `comid-template.json` and `corim-template.json` which contain the respective templates. Some examples of CoMID and CoRIM JSON templates can be found in the [data/templates](data/templates) folder.

### Output File (Optional)

If you wish to specify the name and path of the produced endorsement then pass this via the `corim-file` switch (abbrev. `-c`). If this flag is not set then the produced endorsement will be saved in the current working directory under the file name `psa-endorsements.cbor` or `cca-endorsements.cbor` dependig on the attestation scheme used.

