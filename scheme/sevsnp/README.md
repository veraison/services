# SEV-SNP scheme

This scheme supports the provisioning of reference values and appraisal of evidence. It is suitable for anyone performing verification of simple SEV-SNP evidence.

## Installation

It doesn't need any specific install instructions, it gets deployed along with other schemes.
```
make really-clean && make native-deploy
```

## Usage example

Following is an example of how to interface with this scheme/plugin. The workflow involves using cocli to submit Reference Values and [ratsd](https://github.com/veraison/ratsd) to submit Evidence.

Generating Reference Values and Evidence is beyond this project's scope. Please see [go-gen-ref](https://github.com/jraman567/go-gen-ref) for creating Reference Values for SEV-SNP; ratsd generates Evidence.

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

### Submitting Evidence
```
git clone https://github.com/veraison/ratsd.git
cd ratsd; go build
# From the following command, note down the nonce (in the body) and
# session-id (location in the header), and save them as NONCE and
# SESSION_ID environment variables respectively
curl -sSk -X POST "https://<verifier-hostname-or-IP>:8443/challenge-response/v1/newSession?nonceSize=64" -H "accept: application/vnd.veraison.challenge-response-session+json" -i
EVIDENCE_TOKEN=$(curl -sS -X POST http://<attester-hostname-or-IP>:8895/ratsd/chares -H "Content-Type: application/vnd.veraison.chares+json" -d "{\"nonce\":\"$NONCE\"}")
ATTESTATION_RESULT=$(curl -sSk -X POST -H "https://localhost:8443/$SESSION_ID \
  -H "accept: application/vnd.veraison.challenge-response-session+json" \
  -H 'Content-Type: application/eat+cwt; eat_profile="tag:github.com,2025:veraison/ratsd/cmw"' \
  -H "Host: localhost:8443" \
  --data-raw "$EVIDENCE_TOKEN")
echo $ATTESTATION_RESULT
```

## Result
The result is in JWT format. We can print and verify the result using the [ARC tool](https://github.com/veraison/ear/tree/main/arc) as follows.
```
go install github.com/veraison/ear/arc@latest
arc print result.jwt
```
The trustworthiness vector, as shown below, summarizes the result of verification.
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