This directory contains a quick-and-dirty example of performing provisioning
and verification using command line tools that come with a Veraison deployment.

## Prerequisites

You need to have `jq` installed in your path. Optionally, `tmux` can also be
installed.

## Basic flow

> **Note**: by default, end-to-end flow uses PSA data. It can be switched to
> use CCA data by setting the `SCHEME` environment variable:
>
>       export SCHEME=cca

#### 1. Create and start the services deployment.


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


#### 4. Provision endorsements and trust anchors

This populates the stores with the endorsements and trust anchors needed for
verification later.

```sh
./end-to-end provision
```

Optionally, you can verify that the store have been populated:

```sh
veriason check-stores
```

#### 5. Perform verification

As a relying party:

```sh
./end-to-end verify rp
```

As an attester:

```sh
./end-to-end verify attest
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

