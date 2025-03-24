# Services

This repository contains attestation services assembled using Veraison components.

## Pre-built packages

Packages are automatically generated for each monthly tag, you can find them
attached to the [corresponding job
here](https://github.com/veraison/services/actions/workflows/time-package.yml).

## Getting Started

This section contains the instructions for creating a test deployment of
Veraison services and trying out the end-to-end attestation flow using sample
inputs. You have two options for the deployment: either via Docker, or
natively.

### Docker deployment

#### Requirements

This should work on most Linux systems. You need to perform the following
setup:

- Install git
- Install Docker (and ensure the current user is in the `docker` group) and the Docker Buildx plugin (required for building Docker images).
- Install jq
- (optionally) Install tmux -- this only needed if you want to use it.

On Ubuntu you can do this with:

    sudo apt install git docker.io jq tmux docker-buildx
    sudo usermod -a -G docker $USER
    newgrp docker

#### Creating the deployment

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


The `veraison` command allows starting and stopping `veraison` services and
viewing and manipulating Veraison logs and stores. See the output of `veraison
-h` for the full list of available commands.

> [!WARNING]
> The docker deployment is not suitable for production. It is only intended to
> be used in development environments. It is not hardened and cannot handle
> high traffic volumes.

### Native deployment

Clone and cd into the repo:

    git clone https://github.com/veraison/services.git
    cd services

Please see the [native deployment
README](deployments/native/README.md#dependencies) for the full description of
dependencies. We have bootstrap scripts for Ubuntu, Arch, and MacOSX that
ensure that the required dependencies are installed.

    make bootstrap

This will invoke the system package manager via sudo, so you may be prompted
for your password.

If you are on a different OS, please look at one of the scripts under
deployments/native/bootstrap/ and install the equivalent packages for your
system (on MacOS, they should be available via brew).

Veraison will be deployed into the directory specified by `VERAISON_ROOT`
environment variable (or into `${HOME}/veraison-deployment` if it is not set).

    export VERAISON_ROOT=${HOME}/veraison-deployment

You can build, deploy, and start Veraison services with the following sequence
of commands:

    make native-deploy

The whole process might take a few minutes.

You can interact with the deployment via the frontend,
`${VERAISON_ROOT}/bin/varaison`. You can alias the script to just `veraison` by
source an env file from the deployment:

    source ${VERAISON_ROOT}/env/env.bash

(there is an equivalent `env.zsh` for `zsh`).

If you are on a Linux distribution with systemd or on MacOSX, Veraison services
should be running as user systemd units/launchd user agents. Otherwise, you can
run the services inside virtual terminals with

    veraison start-term

(If you have tmux installed, and would prefer to use a tmux session rather than
multiple terminals, you can use `veraison start-tmux` instead.)

> [!IMPORTANT]
> **Windows Subsystem for Linux (WSL) users:** WSL does not support  user
> `systemd` services or spawning virtual terminals, so Veraison will not be
> running after `make native-deploy`, and `veraison start-term` command will
> not work (it will say "no suitable terminal found"). So the only option for
> running veraison under WSL is using `veraison start-tmux` (or by manually
> launching service executables).

You can interact with the deployment via the frontend script. Please see the
script help for details:

    veraison -h

Please see [deployments/native/README.md](deployments/native/README.md) for
more detailed explanation and step-by-step manual deployment instructions.

### End-to-end example

An example of an end-to-end provisioning and verification flow exists
within `end-to-end` directory. This can be used to quickly check the
deployment (alternatively, you can use `make integ-test` to run the integration
tests).

> [!NOTE]
> see [end-to-end/README.md](end-to-end/README.md) for a more detailed
> explanation of the flow.

> [!NOTE]
> There are versions of the script for native and docker deployments. Listings
> below assume docker deployment; if you're using the native deployment, change
> script suffix to "-native"

Before evidence can be attested, trust anchors and reference values need to
provisioned. These are contained within
`end-to-end/inputs/psa-endorsements.cbor` and can be provisioned with

    end-to-end/end-to-end-docker provision

If this does not return an error, the values have been successfully
provisioned. You can verify this by checking the contents of the Veraison
stores with

    veraison stores

You should see a list of JSON structures of the provision values.

You can now verify the evidence with

    end-to-end/end-to-end-docker verify rp

This should output the [EAR](https://github.com/thomas-fossati/draft-ear)
attestation result. The "rp" means you're verifying as the Relying Party; you
can also specify "attest" to verify as an attester.

## Provisioning

Provisioning service provides a REST-based API for external trusted supply
chain actors (for example, Endorsers) to provision Reference Values, Endorsed
Values (known as Endorsements), and Trust Anchors into Veraison Trusted
Services.

This service acts as a frontend for accepting a
[CoRIM](https://github.com/veraison/corim) payload containing Endorsements and
Trust Anchors. On the back end it communicates with
[VTS](#Veraison-Trusted-Services) which receives the payload and uses
Attestation Scheme (e.g. PSA) specific decoders to extract, store, retrieve,
and manage the Endorsements and Trust Anchors. The API details are documented
under [Endorsement Provisioning
Interface](https://github.com/veraison/docs/tree/main/api/endorsement-provisioning).
This service accept a variety of Endorsement Formats. For now, PSA (Profile 1 &
Profile 2), CCA, TPM and Parsec (CCA and TPM) based Endorsements are supported.

Refer to
[scope](https://github.com/veraison/docs/blob/main/project-overview.md#scope---provisioning)
for full set of services that shall be provided by the provisioning service.

## Verification

Verification service provides a REST-based API for external Attesters or
Relying Parties (known as Challengers) to submit Attestation token, containing
Attestation Evidence claims. On the back end, it communicates with
[VTS](#Veraison-Trusted-Services) to appraise the received Evidence and receive
the Attestation Verification Results, which are then passed to the challenger.

This service acts as a frontend for accepting a variety of attestation token
formats. For now, PSA (Profile 1 & Profile 2), CCA and TPM-based attestation
tokens are supported.

The API is based on the Challenge/Response Interaction Models as documented in
[challenge-response](https://github.com/veraison/docs/tree/main/api/challenge-response)


## Veraison Trusted Services

Veraison Trusted Services (VTS) backend provides core services to the
Verification and Provisioning frontends. On the Provisioning path, it receives
a CoRIM payload from the [Provisioning](#Provisioning) frontend, invokes
Attestation Scheme (e.g. PSA) specific logic to decode the payload and generate
Endorsements and Trust Anchors, and saves them in the corresponding stores. On
the Verification path, it invokes Attestation Scheme specific logic to parse
attestation token and extract evidence claims. From these claims, it constructs
a logical index/key to fetch the trust anchors required to verify the token
signature, and the associated Endorsements (`golden values`) used to appraise
the evidence claims. Upon evaluation, it populates the Attestation Result and
communicates it to the [Verification](#Verification) frontend.


More details about the VTS can be found under
[VTS](https://github.com/veraison/docs/tree/main/architecture/verifier#vts)

## KVStore

The key-values store is the Veraison Storage Layer. It is used to store both
Endorsements and Trust Anchors.

KV Store details can be found under
[kvstore](https://github.com/veraison/services/tree/main/kvstore/README.md#kv-store)

