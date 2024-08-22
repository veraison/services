# Attestation Walkthough

This guide explains how to setup a verifier and an attester. To produce a workable example setup, it is necessary to 

- create attestation keys on the attester,
- provisioning trust anchors, reference values and endorsements on the verifier, and
- to create evidence on the attester for consumption by the verifier.

For this example setup evidence is provided in form of an PSA attestation token.

## Software Dependencies

The software depends at least on 

- Docker X, and
- Go

Veraison also has additional dependencies, which are described at https://github.com/veraison/services

> **Important**: Do not install Docker Desktop (at least on Ubuntu) since otherwise Veraison build will fail.

Use the following command to install the dependencies for Veraison on an Ubuntu OS:

~~~bash
sudo apt install git docker.io jq tmux docker-buildx
~~~

## Veraison

This section provides details about the use of the Veraison docker image, which is meant to be used for demo purposes. The repository of Veraison can be found at https://github.com/veraison/services

In summary, the following steps need to be executed. Note that these steps assume the use of the bash shell.

~~~bash
git clone https://github.com/veraison/services.git
cd services
make docker-deploy

source deployments/docker/env.bash
veraison status
~~~

If Veraison services are not running, then they need to be started with the following command:

~~~bash
veraison start
~~~

To check the status of the database use the following command:

~~~bash
veraison check-stores
~~~

Of course, after the initial installation, the database will be empty. The provisioning of reference values, trust anchors and endosements will
be explained below and those will be stored in the database. If you haven't created attestation keys yet, first look at the next section otherwise skip it. 

In case of problems, use the following command (in a separate command line window) to see the log output.

~~~bash
veraison follow <service-name>
~~~

Replace `<service-name>` with one of the services available with Veraison, namely `vts`, `provisioning`, `verification`, `management`, or `keycloak`.

Debug output can be enabled by changing the log level inside `deployments/docker/src/config.yaml.template` before creating the deployment (i.e. running `make docker-deploy`; if a deployment already exists `make really-clean` should be done first). See also https://github.com/veraison/services/tree/main/deployments/docker#debugging for more detailed instructions on how to debug a specific service.

Sometimes it is necessary to remove the content of the database. In this case, run the following command.

~~~bash
veraison clear-stores
~~~

To completely remove the deployment and remove the build artifacts use the following command. Note that this requires you to re-run `make docker-deploy` if you want to use Veraison again.

~~~bash
make -C ../deployments/docker really-clean
~~~

## Creating an Attestation Key for the Attester

Attesters need to have an attestation key to sign attestation Evidence. If this key is not yet available, here are instructions for creating it.
It is also necessary to convert the keys into the JSON Web Key (JWK) format and a separate tool is used for this transcoding. 

To start, we create a private key in PEM format using the following command:

~~~bash
openssl genpkey -algorithm ec -pkeyopt ec_paramgen_curve:prime256v1 -out key.pem
~~~

We use the `go-pem2jwk` tool to convert the PEM-encoded private key into a JWK. First, we need to obtain the tool.

~~~bash
go install github.com/thomas-fossati/go-pem2jwk@latest
~~~

The go-pem2jwk tool can convert a key in PEM format into a JWK using the following command:

~~~bash
go-pem2jwk < key.pem > jwk.json
~~~

GitHub will complain if we show the PEM-encoded private key. Hence, we only print the JWT containing the private key for this example.
You are supposed to create your private key for your example setup.

~~~json
{
  "crv": "P-256",
  "d": "-fVWtEiKGbVk1J92WRwCefa78RohjMUBVKRfKARMFSQ",
  "kty": "EC",
  "x": "Oq7AxYePubv1b3bhcszgwycyDKDBvIRL400LA4xoJWA",
  "y": "eXVOAz4k28xU4ylyQszt6CorQ7_EQdutFjDYSLRiUG4"
}
~~~

Next, we create a public key based on the private key using the following command:

~~~bash
openssl ec -in key.pem -pubout -outform pem -out cert.pub
~~~

The resulting public key in PEM format is:

~~~
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEOq7AxYePubv1b3bhcszgwycyDKDB
vIRL400LA4xoJWB5dU4DPiTbzFTjKXJCzO3oKitDv8RB260WMNhItGJQbg==
-----END PUBLIC KEY-----
~~~

We also convert the PEM-encoded public key to a JWK:

~~~
go-pem2jwk < cert.pub > pub.json
~~~

The result is shown below:

~~~json
{
  "crv": "P-256",
  "kty": "EC",
  "x": "Oq7AxYePubv1b3bhcszgwycyDKDBvIRL400LA4xoJWA",
  "y": "eXVOAz4k28xU4ylyQszt6CorQ7_EQdutFjDYSLRiUG4"
}
~~~

## Provisioning Trust Anchors, Reference Values and Endorsements

The Verifier must be provisioned with trust anchors, reference values, and endorsements before it can be used. The `cocli` tool offers this functionality and supports several subcommands:

- `comid` to create, display and validate Concise Module Identifiers (CoMIDs)
- `cots` to create, display and validate Concise TA Stores (CoTSs)
- `corim` to create, display, sign, and verify CoRIMs 

See also the [detailed documentation](https://github.com/veraison/corim/tree/main/cocli#readme) on the `veraison/corim` repository.

~~~bash
go install github.com/veraison/corim/cocli@latest
~~~

The concepts are described in separate IETF drafts, which are still work in progress.

- CoTSs are explained in https://datatracker.ietf.org/doc/draft-ietf-rats-concise-ta-stores/

- CoRIMs and CoMIDs are described in https://datatracker.ietf.org/doc/draft-ietf-rats-corim/

The payloads are encoded in CBOR.

It is useful to clone the `veraison/corim` repository, which contains examples. The command is:

~~~bash
git clone https://github.com/veraison/corim.git
~~~

In the subsequent steps, we will create the CBOR-based payloads for submission to Veraison using a RESTful API.
Hence, we will have to create a CoMID for the reference value and a CoTS for the trust anchor.

The command line tool, `cocli`, uses so-called templates as input to generate the CBOR files. For completeness,
the JSON files are shown below.

The first command creates a CoMID based on two input files, one for trust anchors and another one for reference values.
The `comid-psa-ta.json` file contains information about the trust anchor(s) and the `comid-psa-refval.json` file contains
the reference value(s).

In our example, we need to provision the public key of the Attester to Veraison. This allows Veraison to verify the
digital signature covering the Evidence (and the PSA attestation token in our case).

The Evidence, which we will create later in this walkthrough, contains the hashes of the bootloader (BL), the PSA Root of Trust (PRoT) and the Application Root of Trust (ARoT).
The details of these software components are explained in https://datatracker.ietf.org/doc/draft-tschofenig-rats-psa-token/
The goal is to provision the "reference" or "golden values" of these hashes to the Verifier. This allows the Verifier to match the Evidence claims sent by the Attester with the expected values.

~~~bash
cocli comid create --template comid-psa-ta.json \
                    --template comid-psa-refval.json \
                    --output-dir tmp                       
~~~

The public key, which we previously created, has to be copied into the `comid-psa-ta.json` file of the `verification-keys` field. Search for the `BEGIN PUBLIC KEY` section in the JSON file below.
The UEID field contains information about the specific instance of the device and serves as a key identifier. Hence, this instance ID has to match the Instance ID claim in the Evidence.
The Implementation ID claim uniquely identifies the implementation of the immutable PSA RoT. Veraison uses this claim to locate the details of the PSA RoT implementation. The Implementation ID is also
found in the Evidence claims.

comid-psa-ta.json:

~~~json
{
  "lang": "en-GB",
  "tag-identity": {
    "id": "366D0A0A-5988-45ED-8488-2F2A544F6242",
    "version": 0
  },
  "entities": [
    {
      "name": "ACME Ltd.",
      "regid": "https://acme.example",
      "roles": [
        "tagCreator",
        "creator",
        "maintainer"
      ]
    }
  ],
  "triples": {
    "attester-verification-keys": [
      {
        "environment": {
          "class": {
            "id": {
              "type": "psa.impl-id",
              "value": "YWNtZS1pbXBsZW1lbnRhdGlvbi1pZC0wMDAwMDAwMDE="
            },
            "vendor": "ACME",
            "model": "RoadRunner"
          },
          "instance": {
            "type": "ueid",
            "value": "Ac7rrnuJJ6MiflMDz14PH3s0u1Qq1yUKwD+83jbsLxUI"
          }
        },
        "verification-keys": [
          {
            "type": "pkix-base64-key",
	    "value": "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEOq7AxYePubv1b3bhcszgwycyDKDBvIRL400LA4xoJWB5dU4DPiTbzFTjKXJCzO3oKitDv8RB260WMNhItGJQbg==\n-----END PUBLIC KEY-----"

          }
        ]
      }
    ]
  }
}
~~~

The reference values of the software contain the measurements with measurements for each component. The semantics of the reference values are described in https://datatracker.ietf.org/doc/html/draft-fdb-rats-psa-endorsements-03 (see Section 3.3), which is a measurement-map of a CoMID reference-triple-record. Because of the relationship to the software component contained in the Evidence
consult also Section 4.4.1 of https://datatracker.ietf.org/doc/draft-tschofenig-rats-psa-token/.

Four fields are worth mentioning, namely 
 
 - Label: The Label, which corresponds to the Measurement Type in the Evidence, is short string representing the role of this software component, such as "BL" or "PRoT".
 - Version: The Version is the issued software version in the form of a text string. 
 - Signer-ID: The Signer ID is the hash of a signing authority public key for the software component. This can be used by a Verifier to ensure the components
   were signed by an expected trusted source.
 - digests: The Digests field, which corresponds to the Measurement Value in the Evidence, represents a hash of the invariant software component in memory at startup time.

comid-psa-refval.json:

~~~json
{
  "lang": "en-GB",
  "tag-identity": {
    "id": "43BBE37F-2E61-4B33-AED3-53CFF1428B16",
    "version": 0
  },
  "entities": [
    {
      "name": "ACME Ltd.",
      "regid": "https://acme.example",
      "roles": [
        "tagCreator",
        "creator",
        "maintainer"
      ]
    }
  ],
  "triples": {
    "reference-values": [
      {
        "environment": {
          "class": {
            "id": {
              "type": "psa.impl-id",
              "value": "YWNtZS1pbXBsZW1lbnRhdGlvbi1pZC0wMDAwMDAwMDE="
            },
            "vendor": "ACME",
            "model": "RoadRunner"
          }
        },
        "measurements": [
          {
            "key": {
              "type": "psa.refval-id",
              "value": {
                "label": "BL",
                "version": "2.1.0",
                "signer-id": "rLsRx+TaIXIFUjzkzhokWuGiOa48a/2eeHH35di66Gs="
              }
            },
            "value": {
              "digests": [
                "sha-256;h0KPxSKAPTEGXnvOPPA/5HUJZjHl4Hu9eg/eYMTPJcc="
              ]
            }
          },
          {
            "key": {
              "type": "psa.refval-id",
              "value": {
                "label": "PRoT",
                "version": "1.3.5",
                "signer-id": "rLsRx+TaIXIFUjzkzhokWuGiOa48a/2eeHH35di66Gs="
              }
            },
            "value": {
              "digests": [
                "sha-256;AmOCmYm2/ZVPcrqvL8ZLwuLwHWktTecphuqAj26ZgT8="
              ]
            }
          },
          {
            "key": {
              "type": "psa.refval-id",
              "value": {
                "label": "ARoT",
                "version": "0.1.4",
                "signer-id": "rLsRx+TaIXIFUjzkzhokWuGiOa48a/2eeHH35di66Gs="
              }
            },
            "value": {
              "digests": [
                "sha-256;o6XnFfDMV0pzw/m+u2vCTzL/1bZ7OHJEwskJ2neaFHg="
              ]
            }
          }
        ]
      }
    ]
  }
}
~~~

Next, we combine the two previously generated files (`comid-psa-refval.cbor` and `comid-psa-ta.cbor`) into a CoRIM using the following command:

~~~bash
cocli corim create --template corim-psa.json \
                    --comid tmp/comid-psa-refval.cbor \
                    --comid tmp/comid-psa-ta.cbor \
                    --output tmp/psa-corim.cbor
~~~

The content of the `corim-psa.json` file is shown below. It indicates the attestation Evidence profile being used and other meta-data.

~~~json
{
  "corim-id": "5c57e8f4-46cd-421b-91c9-08cf93e13cfc",
  "profiles": [
    "http://arm.com/psa/iot/1"
  ],
  "validity": {
    "not-before": "2021-12-31T00:00:00Z",
    "not-after": "2025-12-31T00:00:00Z"
  }
}
~~~

The CoRIM file can be submitted to Veraison using the following command:

~~~bash
cocli corim submit \
     --corim-file=tmp/psa-corim.cbor \
     --api-server=http://provisioning-service:8888/endorsement-provisioning/v1/submit \
     --media-type='application/corim-unsigned+cbor; profile="http://arm.com/psa/iot/1"'
~~~

By dumping the Veraison database with the "veraison stores" command we can see the correctly entered entries:

~~~json
TRUST ANCHORS:
--------------
{
  "scheme": "PSA_IOT",
  "type": "trust anchor",
  "subType": "",
  "attributes": {
    "PSA_IOT.hw-model": "RoadRunner",
    "PSA_IOT.hw-vendor": "ACME",
    "PSA_IOT.iak-pub": "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEOq7AxYePubv1b3bhcszgwycyDKDBvIRL400LA4xoJWB5dU4DPiTbzFTjKXJCzO3oKitDv8RB260WMNhItGJQbg==\n-----END PUBLIC KEY-----",
    "PSA_IOT.impl-id": "YWNtZS1pbXBsZW1lbnRhdGlvbi1pZC0wMDAwMDAwMDE=",
    "PSA_IOT.inst-id": "Ac7rrnuJJ6MiflMDz14PH3s0u1Qq1yUKwD+83jbsLxUI"
  }
}

ENDORSEMENTS:
-------------
{
  "scheme": "PSA_IOT",
  "type": "reference value",
  "subType": "PSA_IOT.sw-component",
  "attributes": {
    "PSA_IOT.hw-model": "RoadRunner",
    "PSA_IOT.hw-vendor": "ACME",
    "PSA_IOT.impl-id": "YWNtZS1pbXBsZW1lbnRhdGlvbi1pZC0wMDAwMDAwMDE=",
    "PSA_IOT.measurement-desc": "sha-256",
    "PSA_IOT.measurement-type": "BL",
    "PSA_IOT.measurement-value": "h0KPxSKAPTEGXnvOPPA/5HUJZjHl4Hu9eg/eYMTPJcc=",
    "PSA_IOT.signer-id": "rLsRx+TaIXIFUjzkzhokWuGiOa48a/2eeHH35di66Gs=",
    "PSA_IOT.version": "2.1.0"
  }
}
{
  "scheme": "PSA_IOT",
  "type": "reference value",
  "subType": "PSA_IOT.sw-component",
  "attributes": {
    "PSA_IOT.hw-model": "RoadRunner",
    "PSA_IOT.hw-vendor": "ACME",
    "PSA_IOT.impl-id": "YWNtZS1pbXBsZW1lbnRhdGlvbi1pZC0wMDAwMDAwMDE=",
    "PSA_IOT.measurement-desc": "sha-256",
    "PSA_IOT.measurement-type": "PRoT",
    "PSA_IOT.measurement-value": "AmOCmYm2/ZVPcrqvL8ZLwuLwHWktTecphuqAj26ZgT8=",
    "PSA_IOT.signer-id": "rLsRx+TaIXIFUjzkzhokWuGiOa48a/2eeHH35di66Gs=",
    "PSA_IOT.version": "1.3.5"
  }
}
{
  "scheme": "PSA_IOT",
  "type": "reference value",
  "subType": "PSA_IOT.sw-component",
  "attributes": {
    "PSA_IOT.hw-model": "RoadRunner",
    "PSA_IOT.hw-vendor": "ACME",
    "PSA_IOT.impl-id": "YWNtZS1pbXBsZW1lbnRhdGlvbi1pZC0wMDAwMDAwMDE=",
    "PSA_IOT.measurement-desc": "sha-256",
    "PSA_IOT.measurement-type": "ARoT",
    "PSA_IOT.measurement-value": "o6XnFfDMV0pzw/m+u2vCTzL/1bZ7OHJEwskJ2neaFHg=",
    "PSA_IOT.signer-id": "rLsRx+TaIXIFUjzkzhokWuGiOa48a/2eeHH35di66Gs=",
    "PSA_IOT.version": "0.1.4"
  }
}
~~~


## Manually Creating Attestation Evidence

We use the `evcli` tool to create attestation Evidence. Note that only two attestation formats are currently supported, namely the Arm PSA Token and Arm CCA. The repository can be found here: https://github.com/veraison/evcli/tree/main. In a more realistic setup, we would be using either software that emulates an Attester or, even better, a device that supports this functionality (like an Arm v8-M development board).

To install the code, run

~~~bash
go install github.com/veraison/evcli/v2@latest
~~~

Also, clone the repository to re-use the examples:

~~~bash
git clone https://github.com/veraison/evcli.git
~~~

Note that we will re-use the previously created JWT in this example.

The `evcli` repository contains documentation for the use of the PSA attestation token format, which can be found at https://github.com/veraison/evcli/blob/main/README-PSA.md

Two inputs are necessary to create the PSA attestation token, namely 

* A set of claims, and
* A private key to sign the token.

We are using the following claims, in JSON format, and encoding them into a file `psa-evidence.json`. Note that the combination of the `psa-instance-id` and the `psa-implementation-id` is used to identify the key. The `signer-id` contains the hash of the public key used to sign the software/firmware. These concepts are described in https://datatracker.ietf.org/doc/draft-tschofenig-rats-psa-token/

Note that the content of the evidence needs to correspond to the endorsements. Omitting claims or software components will cause verification failures.

~~~json
{
  "eat-profile": "http://arm.com/psa/2.0.0",
  "psa-client-id": 1,
  "psa-security-lifecycle": 12288,
  "psa-implementation-id": "YWNtZS1pbXBsZW1lbnRhdGlvbi1pZC0wMDAwMDAwMDE=",
  "psa-boot-seed": "3q2+796tvu/erb7v3q2+796tvu/erb7v3q2+796tvu8=",
  "psa-hardware-version": "1234567890123",
  "psa-software-components": [
    {
      "measurement-type": "BL",
      "measurement-value": "h0KPxSKAPTEGXnvOPPA/5HUJZjHl4Hu9eg/eYMTPJcc=",
      "signer-id": "rLsRx+TaIXIFUjzkzhokWuGiOa48a/2eeHH35di66Gs=",
      "version": "2.1.0"
    },
    {
      "measurement-type": "PRoT",
      "measurement-value": "AmOCmYm2/ZVPcrqvL8ZLwuLwHWktTecphuqAj26ZgT8=",
      "signer-id": "rLsRx+TaIXIFUjzkzhokWuGiOa48a/2eeHH35di66Gs=",
      "version": "1.3.5"
    },
    {
      "measurement-type": "ARoT",
      "measurement-value": "o6XnFfDMV0pzw/m+u2vCTzL/1bZ7OHJEwskJ2neaFHg=",
      "signer-id": "rLsRx+TaIXIFUjzkzhokWuGiOa48a/2eeHH35di66Gs=",
      "version": "0.1.4"
    }
  ],
  "psa-instance-id": "Ac7rrnuJJ6MiflMDz14PH3s0u1Qq1yUKwD+83jbsLxUI",
  "psa-verification-service-indicator": "https://psa-verifier.org",
  "psa-nonce": "QUp8F0FBs9DpodKK8xUg8NQimf6sQAfe2J1ormzZLxk="
}
~~~

To create a PSA attestation token from the supplied claims and an attestation key in JSON Web Key (JWK) format we use the following command. Follow the instructions in the previous sub-section to create the attestation key.

~~~bash
evcli psa create \
    --claims=psa-evidence.json \
    --key=jwk.json \
    --token=my-token.cbor
~~~

The specification of the PSA attestation token can be found at https://datatracker.ietf.org/doc/html/draft-tschofenig-rats-psa-token, which contains an explanation of the various claims and their semantics.

The command above will produce the PSA attestation token in CBOR format and protect it using COSE_Sign1. The result is stored in `my-token.cbor`.

To verify the Evidence locally, the token and the public key need to be provided to the following command:

~~~bash
evcli psa check --token=my-token.cbor --key=pub.json
~~~

If successful, it will return the list of claims:

~~~
>> "my-token.cbor" verified
>> embedded claims:
{"eat-profile":http://arm.com/psa/2.0.0,"psa-client-id":1,"psa-security-lifecycle":12288,"psa-implementation-id":"UFFSU1RVVldQUVJTVFVWV1BRUlNUVVZXUFFSU1RVVlc=","psa-boot-seed":"3q2+796tvu/erb7v3q2+796tvu/erb7v3q2+796tvu8=","psa-software-components":[{"measurement-type":"BL","measurement-value":"AAECBAABAgQAAQIEAAECBAABAgQAAQIEAAECBAABAgQ=","signer-id":"UZIA/1GSAP9RkgD/UZIA/1GSAP9RkgD/UZIA/1GSAP8="},{"measurement-type":"PRoT","measurement-value":"BQYHCAUGBwgFBgcIBQYHCAUGBwgFBgcIBQYHCAUGBwg=","signer-id":"UZIA/1GSAP9RkgD/UZIA/1GSAP9RkgD/UZIA/1GSAP8="}],"psa-nonce":"QUp8F0FBs9DpodKK8xUg8NQimf6sQAfe2J1ormzZLxk=","psa-instance-id":"AaChoqOgoaKjoKGio6ChoqOgoaKjoKGio6ChoqOgoaKj","psa-verification-service-indicator":https://psa-verifier.org}
~~~

The `psa check` subcommand verifies the digital signature over the supplied PSA attestation token and checks whether its claim set is well-formed.

To test it against the Verifier, the `psa verify-as` subcommand is used.

It has two modes, namely one where the tool acts as an Attester and another mode where it acts as a Relying Party. The Relying Party mode uses the previously generated PSA token as input while the Attester mode creates the PSA attestation token on-the-fly. 

Below, we use the Relying Party mode:

~~~bash
evcli psa verify-as relying-party \
    --api-server=http://verification-service:8080/challenge-response/v1/newSession \
    --token=my-token.cbor
~~~

The response will be an Attestation Result encoded as a JWT, which is signed with a JSON Web Signature (JWS).

For example, the following JWT is an example response returned by the Verifier. It is a string consisting of three values separated by '.'. The first part is the header containing the signing algorithm and other information. The second part is the signed payload, and the last part is the digital signature itself.

~~~
eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJlYXIudmVyaWZpZXItaWQiOnsiYnVpbGQiOiJOL0EiLCJkZXZlbG9wZXIiOiJWZXJhaXNvbiBQcm9qZWN0In0sImVhdF9ub25jZSI6IkZlaUJMMFlzMHl2WGlCYkFGTXMxT0hEWFh0dzA4UkdxX1NFU0pkU2FYUHNLazJBOF9BcnVMNDVaOFFxdWtUOG8iLCJlYXRfcHJvZmlsZSI6InRhZzpnaXRodWIuY29tLDIwMjM6dmVyYWlzb24vZWFyIiwiaWF0IjoxNzA2MDIzOTc5LCJzdWJtb2RzIjp7IlBTQV9JT1QiOnsiZWFyLmFwcHJhaXNhbC1wb2xpY3ktaWQiOiJwb2xpY3k6UFNBX0lPVCIsImVhci5zdGF0dXMiOiJhZmZpcm1pbmciLCJlYXIudHJ1c3R3b3J0aGluZXNzLXZlY3RvciI6eyJjb25maWd1cmF0aW9uIjowLCJleGVjdXRhYmxlcyI6MiwiZmlsZS1zeXN0ZW0iOjAsImhhcmR3YXJlIjoyLCJpbnN0YW5jZS1pZGVudGl0eSI6MiwicnVudGltZS1vcGFxdWUiOjIsInNvdXJjZWQtZGF0YSI6MCwic3RvcmFnZS1vcGFxdWUiOjJ9LCJlYXIudmVyYWlzb24uYW5ub3RhdGVkLWV2aWRlbmNlIjp7ImVhdC1wcm9maWxlIjoiaHR0cDovL2FybS5jb20vcHNhLzIuMC4wIiwicHNhLWJvb3Qtc2VlZCI6IjNxMis3OTZ0dnUvZXJiN3YzcTIrNzk2dHZ1L2VyYjd2M3EyKzc5NnR2dTg9IiwicHNhLWNsaWVudC1pZCI6MSwicHNhLWltcGxlbWVudGF0aW9uLWlkIjoiWVdOdFpTMXBiWEJzWlcxbGJuUmhkR2x2YmkxcFpDMHdNREF3TURBd01ERT0iLCJwc2EtaW5zdGFuY2UtaWQiOiJBYzdycm51Sko2TWlmbE1EejE0UEgzczB1MVFxMXlVS3dEKzgzamJzTHhVSSIsInBzYS1ub25jZSI6IkZlaUJMMFlzMHl2WGlCYkFGTXMxT0hEWFh0dzA4UkdxL1NFU0pkU2FYUHNLazJBOC9BcnVMNDVaOFFxdWtUOG8iLCJwc2Etc2VjdXJpdHktbGlmZWN5Y2xlIjoxMjI4OCwicHNhLXNvZnR3YXJlLWNvbXBvbmVudHMiOlt7Im1lYXN1cmVtZW50LXR5cGUiOiJCTCIsIm1lYXN1cmVtZW50LXZhbHVlIjoiaDBLUHhTS0FQVEVHWG52T1BQQS81SFVKWmpIbDRIdTllZy9lWU1UUEpjYz0iLCJzaWduZXItaWQiOiJyTHNSeCtUYUlYSUZVanpremhva1d1R2lPYTQ4YS8yZWVISDM1ZGk2NkdzPSIsInZlcnNpb24iOiIyLjEuMCJ9LHsibWVhc3VyZW1lbnQtdHlwZSI6IlBSb1QiLCJtZWFzdXJlbWVudC12YWx1ZSI6IkFtT0NtWW0yL1pWUGNycXZMOFpMd3VMd0hXa3RUZWNwaHVxQWoyNlpnVDg9Iiwic2lnbmVyLWlkIjoickxzUngrVGFJWElGVWp6a3pob2tXdUdpT2E0OGEvMmVlSEgzNWRpNjZHcz0iLCJ2ZXJzaW9uIjoiMS4zLjUifSx7Im1lYXN1cmVtZW50LXR5cGUiOiJBUm9UIiwibWVhc3VyZW1lbnQtdmFsdWUiOiJvNlhuRmZETVYwcHp3L20rdTJ2Q1R6TC8xYlo3T0hKRXdza0oybmVhRkhnPSIsInNpZ25lci1pZCI6InJMc1J4K1RhSVhJRlVqemt6aG9rV3VHaU9hNDhhLzJlZUhIMzVkaTY2R3M9IiwidmVyc2lvbiI6IjAuMS40In1dLCJwc2EtdmVyaWZpY2F0aW9uLXNlcnZpY2UtaW5kaWNhdG9yIjoiaHR0cHM6Ly9wc2EtdmVyaWZpZXIub3JnIn19fX0.r85Kv2iRZvQ2mIn70YKKfYF4apv7lhXdoiqao0Z6UlltXifDig9mPDLMvI4JKXKhlumzRZN3kCR54pcJBuCasw
~~~

The attestation result can be processed by a dedicated command line tool called 'arc'. The benefit of 'arc' is the proper decoding of the result. The documentation of the 'arc' tool can be found at https://github.com/veraison/ear/tree/main/arc

First, install the tool with the following command:

~~~bash
go install github.com/veraison/ear/arc@latest
~~~

To obtain the public key for verifying the attestation result fetch it from .well-known using the following command:

~~~curl
wget http://localhost:8080/.well-known/veraison/verification
~~~

The result may be something like this:

~~~json
{
  "ear-verification-key": {
    "alg": "ES256",
    "crv": "P-256",
    "kty": "EC",
    "x": "usWxHK2PmfnHKwXPS54m0kTcGJ90UiglWiGahtagnv8",
    "y": "IBOL-C3BttVivg-lSreASjpkttcsz-1rb7btKLv8EX4"
  },
  "media-types": [
    "application/eat-collection; profile=\"http://arm.com/CCA-SSD/1.0.0\"",
    "application/eat-cwt; profile=\"http://arm.com/psa/2.0.0\"",
    "application/pem-certificate-chain",
    "application/vnd.enacttrust.tpm-evidence",
    "application/vnd.parallaxsecond.key-attestation.cca",
    "application/vnd.parallaxsecond.key-attestation.tpm",
    "application/psa-attestation-token"
  ],
  "version": "commit-b50b67d",
  "service-state": "READY",
  "api-endpoints": {
    "newChallengeResponseSession": "/challenge-response/v1/newSession"
  }
}
~~~

Store the public key from the structure above in a separate file and verify the attestation result using `arc` using the following command. We assume that the attestation result is stored in `ar.txt`.

~~~bash
arc verify --pkey=public_key.json --verbose --alg=ES256 ar.txt
~~~

The result is then shown as follows:

~~~
>> "ar.txt" signature successfully verified using "public_key.json"
[claims-set]
{
    "ear.verifier-id": {
        "build": "N/A",
        "developer": "Veraison Project"
    },
    "eat_nonce": "FeiBL0Ys0yvXiBbAFMs1OHDXXtw08RGq_SESJdSaXPsKk2A8_AruL45Z8QqukT8o",
    "eat_profile": "tag:github.com,2023:veraison/ear",
    "iat": 1706023979,
    "submods": {
        "PSA_IOT": {
            "ear.appraisal-policy-id": "policy:PSA_IOT",
            "ear.status": "affirming",
            "ear.trustworthiness-vector": {
                "configuration": 0,
                "executables": 2,
                "file-system": 0,
                "hardware": 2,
                "instance-identity": 2,
                "runtime-opaque": 2,
                "sourced-data": 0,
                "storage-opaque": 2
            },
            "ear.veraison.annotated-evidence": {
                "eat-profile": "http://arm.com/psa/2.0.0",
                "psa-boot-seed": "3q2+796tvu/erb7v3q2+796tvu/erb7v3q2+796tvu8=",
                "psa-client-id": 1,
                "psa-implementation-id": "YWNtZS1pbXBsZW1lbnRhdGlvbi1pZC0wMDAwMDAwMDE=",
                "psa-instance-id": "Ac7rrnuJJ6MiflMDz14PH3s0u1Qq1yUKwD+83jbsLxUI",
                "psa-nonce": "FeiBL0Ys0yvXiBbAFMs1OHDXXtw08RGq/SESJdSaXPsKk2A8/AruL45Z8QqukT8o",
                "psa-security-lifecycle": 12288,
                "psa-software-components": [
                    {
                        "measurement-type": "BL",
                        "measurement-value": "h0KPxSKAPTEGXnvOPPA/5HUJZjHl4Hu9eg/eYMTPJcc=",
                        "signer-id": "rLsRx+TaIXIFUjzkzhokWuGiOa48a/2eeHH35di66Gs=",
                        "version": "2.1.0"
                    },
                    {
                        "measurement-type": "PRoT",
                        "measurement-value": "AmOCmYm2/ZVPcrqvL8ZLwuLwHWktTecphuqAj26ZgT8=",
                        "signer-id": "rLsRx+TaIXIFUjzkzhokWuGiOa48a/2eeHH35di66Gs=",
                        "version": "1.3.5"
                    },
                    {
                        "measurement-type": "ARoT",
                        "measurement-value": "o6XnFfDMV0pzw/m+u2vCTzL/1bZ7OHJEwskJ2neaFHg=",
                        "signer-id": "rLsRx+TaIXIFUjzkzhokWuGiOa48a/2eeHH35di66Gs=",
                        "version": "0.1.4"
                    }
                ],
                "psa-verification-service-indicator": "https://psa-verifier.org"
            }
        }
    }
}
[trustworthiness vectors]
submod(PSA_IOT):
Instance Identity [affirming]: The Attesting Environment is recognized, and the associated instance of the Attester is not known to be compromised.
Configuration [none]: The Evidence received is insufficient to make a conclusion.
Executables [affirming]: Only a recognized genuine set of approved executables, scripts, files, and/or objects have been loaded during and after the boot process.
File System [none]: The Evidence received is insufficient to make a conclusion.
Hardware [affirming]: An Attester has passed its hardware and/or firmware verifications needed to demonstrate that these are genuine/supported.
Runtime Opaque [affirming]: the Attester's executing Target Environment and Attesting Environments are encrypted and within Trusted Execution Environment(s) opaque to the operating system, virtual machine manager, and peer applications.
Storage Opaque [affirming]: the Attester encrypts all secrets in persistent storage via using keys which are never visible outside an HSM or the Trusted Execution Environment hardware.
Sourced Data [none]: The Evidence received is insufficient to make a conclusion.
~~~

Alternatively, it is also possible to display the attestation result using an online tool, for example, https://jwt.io. There are also many command line tools available to parse JWTs.

Once parsed, the header shows the digital signature algorithm that was used to protect the claims of the JWT

~~~json
{
  "alg": "ES256",
  "typ": "JWT"
}
~~~

The header is followed by this payload:

~~~json
{
    "ear.verifier-id": {
        "build": "N/A",
        "developer": "Veraison Project"
    },
    "eat_nonce": "FeiBL0Ys0yvXiBbAFMs1OHDXXtw08RGq_SESJdSaXPsKk2A8_AruL45Z8QqukT8o",
    "eat_profile": "tag:github.com,2023:veraison/ear",
    "iat": 1706023979,
    "submods": {
        "PSA_IOT": {
            "ear.appraisal-policy-id": "policy:PSA_IOT",
            "ear.status": "affirming",
            "ear.trustworthiness-vector": {
                "configuration": 0,
                "executables": 2,
                "file-system": 0,
                "hardware": 2,
                "instance-identity": 2,
                "runtime-opaque": 2,
                "sourced-data": 0,
                "storage-opaque": 2
            },
            "ear.veraison.annotated-evidence": {
                "eat-profile": "http://arm.com/psa/2.0.0",
                "psa-boot-seed": "3q2+796tvu/erb7v3q2+796tvu/erb7v3q2+796tvu8=",
                "psa-client-id": 1,
                "psa-implementation-id": "YWNtZS1pbXBsZW1lbnRhdGlvbi1pZC0wMDAwMDAwMDE=",
                "psa-instance-id": "Ac7rrnuJJ6MiflMDz14PH3s0u1Qq1yUKwD+83jbsLxUI",
                "psa-nonce": "FeiBL0Ys0yvXiBbAFMs1OHDXXtw08RGq/SESJdSaXPsKk2A8/AruL45Z8QqukT8o",
                "psa-security-lifecycle": 12288,
                "psa-software-components": [
                    {
                        "measurement-type": "BL",
                        "measurement-value": "h0KPxSKAPTEGXnvOPPA/5HUJZjHl4Hu9eg/eYMTPJcc=",
                        "signer-id": "rLsRx+TaIXIFUjzkzhokWuGiOa48a/2eeHH35di66Gs=",
                        "version": "2.1.0"
                    },
                    {
                        "measurement-type": "PRoT",
                        "measurement-value": "AmOCmYm2/ZVPcrqvL8ZLwuLwHWktTecphuqAj26ZgT8=",
                        "signer-id": "rLsRx+TaIXIFUjzkzhokWuGiOa48a/2eeHH35di66Gs=",
                        "version": "1.3.5"
                    },
                    {
                        "measurement-type": "ARoT",
                        "measurement-value": "o6XnFfDMV0pzw/m+u2vCTzL/1bZ7OHJEwskJ2neaFHg=",
                        "signer-id": "rLsRx+TaIXIFUjzkzhokWuGiOa48a/2eeHH35di66Gs=",
                        "version": "0.1.4"
                    }
                ],
                "psa-verification-service-indicator": "https://psa-verifier.org"
            }
        }
    }
}
~~~

The claims contained in this Attestation Result are described in https://datatracker.ietf.org/doc/draft-fv-rats-ear/. The trustworthiness vector shows the processing of the evaluation result. The overall appraisal status for the attester is found in the `ear.status` field. The values for these claims are re-used from another specification, namely from AR4SI (see https://datatracker.ietf.org/doc/draft-ietf-rats-ar4si/).

To use the Attester mode, use the following command assuming the private key is available in JWK format and has been copied into the same directory where the two input files are located.

~~~bash
evcli psa verify-as attester \
    --api-server=http://verification-service:8080/challenge-response/v1/newSession \
    --claims=psa-evidence-without-nonce.json \
    --key=jwk.json
~~~

The content of `psa-evidence-without-nonce.json` corresponds to the content of the previously used file `psa-evidence.json` but with the nonce claim omitted.

If successful, this protocol interaction will produce an attestation result as a JWT. 
