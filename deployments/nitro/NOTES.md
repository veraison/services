# Notes

* Using the [AWS Nitro Enclaves Developer](https://aws.amazon.com/marketplace/pp/prodview-37z6ersmwouq2) AMI

Things to remember:
* add an ssh key (`nitraison-dev.pem`)
* "enable enclave" in the advanced options
* add a public IP from the elastic IP pool
* the preset user for the AMI is `ec2-user`

* SSH into the EC2 instance:
```shell
ssh -i "~/.ssh/nitraison-dev.pem" ec2-user@<EC2_INSTANCE_NAME>
```

* install git
```shell
sudo yum install git -y
```

* Build the Rust vsock example following the [instructions](https://docs.aws.amazon.com/enclaves/latest/user/developing-applications-linux.html)

* Create the enclave image file:
```shell
nitro-cli build-enclave --docker-dir ./ --docker-uri vsock-sample-server --output-file vsock_sample.eif
```
which returns the set of measurements:
```
Start building the Enclave Image...
Enclave Image successfully created.
{
  "Measurements": {
    "HashAlgorithm": "Sha384 { ... }",
    "PCR0": "9f2825885724e9998bbe9b07660f0650978526d6acc962e996c8de1de95bf44ad5075d9ade8ac78900c2536463acd8c3",
    "PCR1": "bcdf05fefccaa8e55bf2c8d6dee9e79bbff31e34bf28a99aa19e6b29c37ee80b214a414b7607236edf26fcb78654e63f",
    "PCR2": "6d2d37f1d87da28618ccec0120c0b554314fa2f537d9a0e797a39ad2f3c3f14dc848c8ee13a9f758f830e4b8c65d1724"
  }
}
```

* Launch the enclave:
```shell
nitro-cli run-enclave --eif-path vsock_sample.eif --cpu-count 2 --enclave-cid 6 --memory 256 --debug-mode
```
which returns
```
Start allocating memory...
Started enclave with enclave-cid: 6, memory: 256 MiB, cpu-ids: [1, 5]
{
  "EnclaveName": "vsock_sample",
  "EnclaveID": "i-099db943bfa3ad1d9-enc1879f5d1481d8f0",
  "ProcessID": 25973,
  "EnclaveCID": 6,
  "NumberOfCPUs": 2,
  "CPUIDs": [
    1,
    5
  ],
  "MemoryMiB": 256
}
```

* Use the enclave ID to monitor it:
```shell
nitro-cli console --enclave-id i-099db943bfa3ad1d9-enc1879f5d1481d8f0
```

* In another shell, run the client:
```shell
./aws-nitro-enclaves-samples/vsock_sample/rs/target/x86_64-unknown-linux-musl/release/vsock-sample client --cid 6 --port 5005
```

* To terminate all the running enclaves:
```shell
nitro-cli terminate-enclave --all
```

## Veraison

### Dependencies

* install go 1.19
```shell
wget https://go.dev/dl/go1.19.linux-amd64.tar.gz
tar -xzf go1.19.linux-amd64.tar.gz
sudo mv go /usr/local
export PATH=$PATH:/usr/local/go/bin
```

* install go deps
```shell
go install golang.org/x/tools/gopls@latest
go install golang.org/x/tools/cmd/guru@latest
go install github.com/golang/mock/mockgen@v1.7.0-rc.1
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
go install github.com/mitchellh/protoc-gen-go-json@v1.1.0
go install github.com/veraison/corim/cocli@latest
go install github.com/veraison/evcli@latest
go install github.com/go-delve/delve/cmd/dlv@latest
export PATH=$PATH:${HOME}/go/bin
```
* also:
```shell
curl -L -O https://github.com/protocolbuffers/protobuf/releases/download/v22.3/protoc-22.3-linux-x86_64.zip
sudo unzip protoc-22.3-linux-x86_64.zip -d /usr/local
```

### Veraison Proper

* checkout code
```shell
git checkout https://github.com/veraison/services
cd services
```

* compile versaison services
```shell
make SCHEME_LOADER=builtin
```

#### Nitro management scripts

* nitro-build.sh
```shell
#!/bin/bash

set -eux
set -o pipefail

. nitro.env

docker build -t "${ENCLAVE}" .

nitro-cli build-enclave \
	--docker-dir ./ \
	--docker-uri "${ENCLAVE}" \
	--output-file "${ENCLAVE_IMG}"
```

* nitro-kill.sh
```shell
#!/bin/bash

set -eux
set -o pipefail

ENCLAVE_ID=$(nitro-cli describe-enclaves | jq -r ".[0].EnclaveID")
[ "$ENCLAVE_ID" != "null" ] && nitro-cli terminate-enclave --enclave-id ${ENCLAVE_ID}
```

* nitro-run.sh
```shell
#!/bin/bash

set -eux
set -o pipefail

. nitro.env

nitro-cli run-enclave \
	--eif-path "${ENCLAVE_IMG}" \
	--cpu-count ${ENCLAVE_CPUS} \
	--enclave-cid ${ENCLAVE_CID} \
	--memory ${ENCLAVE_MEM} \
	--debug-mode

ENCLAVE_ID=$(nitro-cli describe-enclaves | jq -r ".[0].EnclaveID")

[ "$ENCLAVE_ID" != "null" ] && nitro-cli console --enclave-id ${ENCLAVE_ID}
```

* nitro.env
```shell
ENCLAVE_CID=5
ENCLAVE="vts-nitro"
ENCLAVE_IMG="${ENCLAVE}.eif"
ENCLAVE_CPUS=2
ENCLAVE_MEM=512
```

# Demo commands

## Setup

```shell
cd integration-tests/__generated__
```

```shell
export AWS_HOST="ec2-54-228-160-194.eu-west-1.compute.amazonaws.com"
```

## Provisioning

```shell
cocli corim submit \
	-f endorsements/endorsements.cbor \
	-s "http://${AWS_HOST}:8888/endorsement-provisioning/v1/submit" \
	-m 'application/corim-unsigned+cbor; profile=http://arm.com/psa/iot/1'
```

## Verification

```shell
evcli psa verify-as attester \
	-s "http://${AWS_HOST}:8080/challenge-response/v1/newSession" \
	-c claims/psa.good.json \
	-k ../data/keys/ec.p256.jwk \
    | tr -d \" \
    | step crypto jwt inspect --insecure
```

* new Link header

```shell
curl -v -X POST "http://${AWS_HOST}:8080/challenge-response/v1/newSession"
```

## Discovery

```shell
curl "http://${AWS_HOST}:8080/.well-known/veraison/verification" \
    | jq .
```


