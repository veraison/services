This directory contains a quick-and-dirty solution for performing deployment
and end-to-end testing of Veraison services. It is intended as a stop-gap to
assist with testing before a proper deployment and integration testing solution
is implemented.

## Prerequisites

In addition to usual Veraison development dependencies, the following utilities
are assumed to be installed and in PATH: `sqlite3`, `jq`, `tmux`. All of these
should be readily installable from your operating system's package
repositories.

You will also need `cocli` and `evcli` utilities that are available in other
repositories under the Veraison project. If you do not already have them
installed, you can do so with:

```sh
go install github.com/veraison/corim/cocli@demo-psa-1.0.0
go install github.com/veraison/evcli@demo-psa-1.0.0
```

## Configuration

You can optionally set the following environment variables to alter the
script's behavior:

`DEPLOY_DIR`: The location into which Veraison services will be deployed. This
directory will be created if it doesn't already exist (its parent directory
must be writable to by current user). If this is not specified, it defaults to
`/tmp/veraison/`.

`TMUX_SESSION`: The name of the tmux session that will be created for running
the services. This defaults to `veraison`.

## Basic flow

#### 1. Build Veraison

If you already built Veraison by running `make` in the top-level directory of
the repo, then you can skip this step.

```sh
./end-to-end build
```

#### 2. Deploy Veraison

This creates the deployment directory structure and copies built artefacts into
it. This also initializes the sqlite stores.

```sh
./end-to-end deploy
```

#### 3. Run Veraison services

This starts the provisioning and verification API services, and the vts
backend.

```sh
./end-to-end run
```

You should now be able to attach to the tmux session with the services using `tmux
attach -t veraison` (assuming you haven't changed `TMUX_SESSION`). This will
give you to terminal output from the running services.

#### 4. Provision endorsements and trust anchors

This populates the stores with the endorsements and trust anchors needed for
verification later.

```sh
./end-to-end provision
```

Optionally, you can verify that the store have been populated:

```sh
./end-to-end check-stores
```

#### 5. Perform verification

As a relying party:

```sh
./end-to-end virify rp
```

As an attester:

```sh
./end-to-end virify attest
```

## Clean up

You can terminate the tmux session (and therefore the Veraison services that
are running inside it) with

```sh
./end-to-end stop
```

You can clean up the deployment directory with

```sh
./end-to-end clean
```

(note: if you're using the default directory under `/tmp/`, it should be
automatically cleaned up on next reboot.)


## Redeployment for quick development iteration

You can rebuild Veraison, terminate the tmux session and combine steps (3) and (4) in the [Basic flow](#basic-flow) above with a single command:

```sh
./end-to-end redeploy
```

This is the only command you need to run after making changes to Veraison
before you can re-test.
