# Notes

* Using the [AWS Nitro Enclaves Developer](https://aws.amazon.com/marketplace/pp/prodview-37z6ersmwouq2) AMI

Things to remember:
* add an ssh key (`nitraison-dev.pem`)
* "enable enclave" in the advanced options
* add a public IP from the elastic IP pool
* the preset user for the AMI is `ec2-user`

* SSH into the EC2 instance:
```shell
ssh -i "~/.ssh/nitraison-dev.pem" ec2-user@ec2-3-248-250-248.eu-west-1.compute.amazonaws.com
```

* install git
```shell
sudo yum install git -y
```

* Build the Rust vsock example following the instructions at []()

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

* install go
```shell
sudo yum install golang -y
```
the installed version is 1.18
```
go version
go version go1.18.9 linux/amd64
```

