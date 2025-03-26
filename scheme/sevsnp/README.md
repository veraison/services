# SEV-SNP scheme

This scheme supports the provisioning of reference values and appraisal of evidence. It is suitable for anyone performing verification of simple SEV-SNP evidence.

## Installation

It doesn't need any specific install instructions, it gets deployed along with other schemes.
```
make really-clean; make native-deploy
```

## Usage example

Following is an example of how to interface with this scheme/plugin. The workflow involves using cocli to submit reference values and ratsd to submit the evidence.

Since ratsd is under construction, please use the following instance of evcli to submit evidence.
https://github.com/jraman567/evcli

Generating reference values and evidence is beyond this project's scope. Please see go-gen-ref for creating reference values for SEV-SNP; RATSd generates evidence.
go-gen-ref: https://github.com/jraman567/go-gen-ref
ratsd: https://github.com/veraison/ratsd

### Provisioning Trust Anchor
```
cocli comid create --template scheme/sevsnp/test/ta-prov.json
cocli corim create -m ta-prov.cbor -t corimMini.json -o ta.cbor
cocli corim submit --corim-file=ta.cbor --api-server="https://localhost:9443/endorsement-provisioning/v1/submit" --media-type="application/corim-unsigned+cbor; profile=\"https://amd.com/ark\""
```

### Provisioning Reference Values
```
cocli corim submit --corim-file=scheme/sevsnp/test/refval-prov.cbor --api-server="https://localhost:9443/endorsement-provisioning/v1/submit" --media-type="application/corim-unsigned+cbor; profile=\"https://amd.com/ark\""
```

### Submitting evidence
```
git clone https://github.com/jraman567/evcli.git
cd evcli; go build
./evcli sev-snp verify-as relying-party --api-server=https://localhost:8443/challenge-response/v1/newSession --token=cmd/sevsnp/sample/SNP-EAT.json
```

## Result
The result is in JWT format. Decoding it using an online tool like https://jwt.io/ reveals formatted results. The trustworthiness vector, as shown below, summarizes the result of verification.
```
    "SEVSNP": {
      "ear.appraisal-policy-id": "policy:SEVSNP",
      "ear.status": "affirming",
      "ear.trustworthiness-vector": {
        "configuration": 0,
        "executables": 0,
        "file-system": 0,
        "hardware": 2,
        "instance-identity": 0,
        "runtime-opaque": 2,
        "sourced-data": 0,
        "storage-opaque": 0
      },
```