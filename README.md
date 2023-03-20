# Services

This repository contains attestation services assembled using Veraison components.

## Getting Started

This section contains the instructions for creating a test deployment of
Veraison services and trying out the end-to-end attestation flow using sample
inputs.

### Requirements

This should work on most Linux systems. You need to perform the following
setup:

- Install git
- Install Docker, and make sure that the current user is in the `docker` group.
- Install jq
- (optionally) Install tmux -- this only needed if you want to use it.

On Ubuntu you can do this with:

    sudo apt install git docker.io jq tmux
    sudo usermod -a -G docker $USER
    newgrp docker

### Creating a deployment

You can build, deploy, and start Veraison services with the following sequence
of commands:

    git clone https://github.com/veraison/services.git
    cd services
    make docker-deploy

The whole process might take a few minutes. Once the above finishes, Veraison
services should be running inside Docker containers. You can use the deployment
frontend script to check their status (you can set it up by sourcing the
deployment's `env.bash`):

    source deployments/docker/env.bash
    veraison status

See the output of `veraison -h` for the full list of available commands

### End-to-end example

An example of an end-to-end provisioning and verification flow exists
within `end-to-end` directory. This can be used to quickly check the
deployment (alternatively, you can use `make integ-test` to run the integration
tests).

> **Note**: see the [README.md](end-to-end/README.md) inside end-to-end
> directory for a more detailed explanation of the flow.

Before evidence can be attested, trust anchors and reference values need to
provisioned. These are contained within
`end-to-end/inputs/psa-endorsements.cbor` and can be provisioned with

    end-to-end/end-to-end provision

If this does not return an error, the values have been successfully
provisioned. You can verify this by checking the contents of the Veraison
stores with

    veraison stores

You should see a list of JSON structures of the provision values.

You can now verify the evidence with

    end-to-end/end-to-end verify rp

This should output the [EAR](https://github.com/thomas-fossati/draft-ear)
attestation result. The "rp" means you're verifying as the Relying Party; you
can also specify "attest" to verify as an attester.

## Provisioning

Provisioning service provides a REST-based API for external trusted supply chain actors (for example, Endorsers) to provision Reference Values, Endorsed Values (known as Endorsements), and Trust Anchors into Veraison Trusted Services.

It uses Attestation Format (e.g. PSA) specific decoders  to extract the Endorsements from received payload. On the back end it communicates with [VTS](#Veraison-Trusted-Services) to store, retrieve, and manage the Endorsements and Trust Anchors. The API details are documented under [Endorsement Provisioning Interface](https://github.com/veraison/docs/tree/main/api/endorsement-provisioning). Provisioning service acts as a front end to accept a variety of Endorsement Formats. For now, PSA (Profile 1 & Profile 2), CCA and TPM based Endorsements are supported.

Refer to [scope](https://github.com/veraison/docs/blob/main/project-overview.md#scope---provisioning) for full set of services that shall be provided by the provisioning service.

## Verification

Verification service provides a REST-based API for external Attesters or Relying Parties (known as Challengers) to submit Attestation token, containing Attestation Evidence claims. On the back end, it communicates with [VTS](#Veraison-Trusted-Services) to appraise the received Evidence and receive the Attestation Verification Results, which are then passed to the challenger.

This service acts as a frontend for accepting a variety of attestation token formats. For now, PSA (Profile 1 & Profile 2), CCA and TPM-based attestation tokens are supported.

The API is based on the Challenge/Response Interaction Models as documented in [challenge-response](https://github.com/veraison/docs/tree/main/api/challenge-response)


## Veraison Trusted Services

Veraison Trusted Services (VTS) backend provides core services to the Verification and Provisioning frontends. On the Provisioning path, it receives Endorsements and Trust Anchors, from the [Provisioning](#Provisioning) frontend, invokes Attestation Format (e.g. PSA) specific logic (a "scheme") to generate Endorsements and Trust Anchors, and saves them in the corresponding stores. On the Verification path, it invokes Attestation Format specific logic (a "scheme") to parse attestation token and extract evidence claims. From these claims, it constructs a logical index/key to fetch the trust anchors required to verify the token signature, and the associated Endorsements (`golden values`) used to appraise the evidence claims. Upon evaluation, it populates the Attestation Result and communicates it to the [Verification](#Verification) frontend.


More details about the VTS can be found under [VTS](https://github.com/veraison/docs/tree/main/architecture/verifier#vts)

## KVStore

The key-values store is the Veraison Storage Layer. It is used to store both Endorsements and Trust Anchors.

KV Store details can be found under [kvstore](https://github.com/veraison/services/tree/migration/kvstore#kv-store)
