This directory contains a quick-and-dirty example of performing provisioning
and verification using command line tools that come with a Veraison deployment.
All the instructions on this page refer to the Docker-based deployment. To use
the native deployment, just substitute "docker" with "native" in all the
commands given below.  For example, `make native-deploy` instead of `make
docker-deploy`, and `./end-to-end-native provision` instead of `./end-to-end-docker
provision`

## Prerequisites

You need to have `jq` installed in your path. Optionally, `tmux` can also be
installed.

## Basic flow

#### Create and start the services deployment

This can be done with a single make command:

```sh
make -C .. docker-deploy
```

This may take a while. Once it's done, you can gain access to the frontend and
utilities by sourcing the deployment environment file:

```sh
source ../deployments/docker/env.bash
```

You can check that everything is ok with

```sh
veraison status
```

This should report that `vts`, `provisioning`, and `verification` services are
all running.

#### Provision endorsements and trust anchors

> **Note**: by default, end-to-end flow uses PSA data. It can be switched to
> use CCA data by setting the `SCHEME` environment variable:
>
>       export SCHEME=cca

This populates the stores with the endorsements and trust anchors needed for
verification later.

```sh
./end-to-end-docker provision
```

Optionally, you can verify that the store have been populated:

```sh
veraison check-stores
```

#### Perform verification

As a relying party:

```sh
./end-to-end-docker verify rp
```

As an attester:

```sh
./end-to-end-docker verify attest
```

## Clean up

You can terminate the tmux session (and therefore the Veraison services that
are running inside it) with

```sh
veraison stop
```

You can clean up the deployment with

```sh
make -C ../deployments/docker really-clean
```
