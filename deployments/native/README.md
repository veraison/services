This directory contains scripts, config templates, and other sources for a
native deployment of Veraison on the system.

## Dependencies

To build Veraison services, you will need a Go toolchain, at least v1.22.
If Go is already installed, you can check the version:
```sh
go version
```
If not, the instructions for downloading and installing Go on your platform can
be found [here](https://go.dev/dl/).

You will also need the protobuf compiler, `protoc`, with plugins for Go. The
plugins may be installed via `go install`:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
go install github.com/mitchellh/protoc-gen-go-json@latest
```

You will need GNU `make` (at least version 3.81) to drive the build process.

The deployment script is written for `bash` shell, and relies on `envsubst`
utility (which comes as part of `gettext` package on most systems), as well as
some common Linux utilities that are often already present on a system but may
sometimes need to be installed: `which`, `find`, `grep`, `sed`.

> **MacOSX users:** To enable `envsubst` and force it to link properly:
> ```sh
> brew install gettext
> brew link --force gettext
> ```

For a full deployment, the following tools will also be needed:

- `sqlite3` to initialize and manipulate the stores
- `jq` to display the stores' contents
- `openssl` (at least version 3.0) to generate the TLS certificates (not needed
  if you plan to use the included example certs).
- `step` to generate the signing key (not needed if you plan to use the included
  example key.
- `tmux` can optionally be used to run the deployment (not required).

The first three should be available from package managers of most
distributions. For installing `step`, please see [its
documentation](https://smallstep.com/docs/step-cli/installation/).

### Bootstrap scripts

To simplify dependency installation, we have bootstrap scripts available for
Linux Arch and Ubuntu, as well as MacOSX (using [homebrew](https://brew.sh))
(see `bootstrap/` subdirectory). Running

```bash
make bootstrap
```

will try to automatically select the appropriate script to run.

### Checking dependencies

You can check whether the required dependencies are present by running

```bash
make check
```

This will report an error if a mandatory dependency is missing, or warnings if
one or more optional dependencies are missing.

## Quick deployment

To get Veraison running quickly you can run

```bash
make quick-deploy
```

in the `deployments/native` directory, or `make native-deploy` in the repo's
top-level directory.

This will create a deployment under `~/veraison-deployment` using the included
example key and certificates.

The deployment is controlled via a CLI frontend script,
`~/veraison-deployment/bin/veraison`. You can create an alias for it so that it
may be invoked simply as `veraison` by sourcing the deployments env file:

```bash
source ~/veraison-deployment/env/env.bash
```

(the command above assumes you're using `bash` shell; there is also an
equivalent file for `zsh`).

## Running Veraison

You have three options for running the services:

```bash
# option 1
veraison start-term

```

Will spawn a virtual terminal for each service.

```bash
# option 2
export SESSION_NAME=veraison
veraison start-tmux $SESSION_NAME
```

will create a tmux session with the specified name (if no name is specified,
the default is "veraison"), and will start services inside panes within that
session.  You can then attach to the session with

```bash
tmux attach -t $SESSION_NAME
```

Note that this requires that `tmux` is installed on the system.

Finally

```bash
# option 3
veraison start
```

will install and start systemd/launchd services for the current user.

(Note: if you've deployed by running `make native-deploy` at top level, that will
automatically try to start services via systemd/launchd on systems that use it.)

### testing the deployment

You can check that Veraison is running properly by running through the
[end-to-end flow](../../end-to-end/README.md).

## Step-by-step deployment

There are multiple steps to creating a functioning Veraison deployment. `make
quick-deploy` described above executes these steps in sequence with default
options for minimal hassle.

This section describes how to create a deployment by executing each step
individually using the `deployment.sh` script, and the options available for
each step.

### Deployment destination

The destination for the deployment is specified by `VERAISON_ROOT` environment
variable. If the variable does not exist, it defaults to
`~/veraison-deployment` (as seen above).

```bash
export VERAISON_ROOT=~/alternate-deployment
```

If the specified location does not exist, it will be created by the script.

### Step 1: build Veraison

Veraison services can be built with

```bash
./deployment.sh build
```

This is equivalent to running `make COMBINED_PLUGINS=1` at the top level. Note
that `COMBINED_PLUGINS` must be set, as the deployment will not use split
`-handler` plugins.

### Step 2: create and populate deployment directory

```bash
./deployment.sh deploy
```

This create the deployment directory structure under `VERAISON_ROOT` (including
the `VERAISON_ROOT` itself if it doesn't already exist) and will copy the
service executables, (combined) plugins, the CLI frontend, and configuration
into the directory structure.

Alternatively, if you specify `-s` option:

```bash
./deployment.sh -s deploy
```

then the service executables and plugins will be symbolically linked, rather
than copied. This is useful for development, so that you don't have keep
re-copying the executables after rebuilding them.

### Step 3: set up TLS certificates

Veraison services communicate with clients and with each other over TLS. Each
service will use its own certificate and associated key that it will pick up
from `${VERAISON_ROOT}/certs/` based on its name (e.g. VTS service will use
`${VERAISON_ROOT}/certs/vts.crt`.

These certificates can be generated for a deployment by providing a root
certificate that will be used to sign them. The root certificate maybe obtained
from a known Certificate Authority (such as [Let's
Encrypt](https://letsencrypt.org/getting-started/)), or created locally (in
which case, it will be self-signed).

Certificate generation relies on `openssl` which must be installed on the
system.

#### Create root certificate

A root certificate may be created by running

```bash
./deployment.sh create-root-cert "/C=US/O=Acme Inc."
```

This will create `${VERAISON_ROOT}/certs/rootCA.crt` (and the associated
`${VERAISON_ROOT}/certs/rootCA.key`) with the specified subject. The subject
may be omitted from the invocation, in which case `/O=Veraison` will be used.

#### Create service certificates

Once you have a root certificate (either generated in the sub-section above, or
obtained via some other means), you can use it to generate certificates for the
individual services.

In addition to the root certificate and key, you will also need to specify a
comma-separated list of names that will be used to populate the Common Name
field and Subject Alternative Name extension in the generated certificates.

```bash
./deployment.sh init-certificates $(cat /etc/hostname),localhost \
    ${VERAISON_ROOT}/certs/rootCA.crt ${VERAISON_ROOT}/certs/rootCA.key
```

If you only plan to use the services locally, it is sufficient to only provide
"localhost" as the first argument.

The names in the list may contain the placeholder `@@` that will be replaced
with the name of the service.

For example, if you specify the first argument as

    @@.my-domain.com,@@-service.alt-domain.com,localhost

then the certificate generated for the VTS service  will contain the Subject

    /CN=vts.my-domain.com

and the SAN extension with

    DNS:vts.my-domain.com, DNS:vts-service.alt-domain.com, DNS:localhost

> [!NOTE]
> if you're using your own root certificate rather than generating one as in
> sub-section above, you do not need to rename or copy it into
> `${VERAISON_ROOT}/certs/` -- this will be done as part of `init-certificates`
> command (the root cert key will not be copied as it won't be needed after
> service certificates are generated).

#### Using example certificates

It is also possible to just use included example certificates rather then
generating new ones for the deployment (these will, obviously, only work for
local use, as they won't be configured with your host's name).

```bash
./deployment.sh -e init-certificates
```

When doing this, you do not need to create/provide a root certificate.

### Step 4: set up signing key

The signing key should be provided as `${VERAISON_ROOT}/signing/skey.jwk`. It
will be used by the verification service to sign attestation results. The key
must be in [JWK format](https://www.rfc-editor.org/rfc/rfc7517).

You can generate a new key for the deployment with

```bash
./deployment.sh init-signing-key
```

Signing key generation relies on `step` which must be installed on the system
(see [dependencies](#dependencies) section above).

#### Using example signing key

As with certificates, it is possible use the example key instead of generating
a new one:

```bash
./deployment.sh -e init-signing-key
```

### Step 5: initialize stores

Veraison uses sqlite3 databases to store endorsements, trust anchors, and
policies. When started, services will expect these database to be initialized
with appropriate tables. This can be done by running

```bash
./deployment.sh init-stores
```

### Step 6: set up clients

Veraison services can be interacted with via CLI applications, `cocli` for
provisioning, `evcli` for verification, and `pocli` for policy management.

These clients can be installed, along with appropriate configuration with

```bash
./deployment.sh init-clients
```

#### Using existing clients with the deployment

If you already `cocli` or other clients installed, you can use them, rather
than the ones bundled with the deployment. You can use `--config` option
to point to the deployment's config file for the client, e.g.

```bash
cocli --config ${VERAISON_ROOT}/config/cocli/config.yaml OTHER_ARGS...
```

To avoid having to specify the config file each time, you can copy the
client-specific sub-directory into your `XDG_CONFIG_HOME` (usually
`~/.config/`)

```bash
cp -r ${VERAISON_ROOT}/config/cocli ~/.config/
cocli OTHER_ARGS...
```

(The examples above use `cocli` but the same goes for `evcli` and `pocli` as
well.)

## Deployment directory structure

Once deployment has been created, `${VERAISON_ROOT}` has the following
structure

```
.
├── bin
├── certs
├── config
├── env
├── logs
├── plugins
├── signing
├── stores
├── systemd
└── launchd
```

#### `bin`

This directory contains all executables associated with deployment, including
the CLI frontend, services, and clients.

(note: if the deployment was created with `-s` option, the service executables
will in fact be symlinks to their source locations.)

#### `config`

This directory has a number of sub-directories, one for the services, and one
for each client. Each subdirectory contains a `config.yaml` with associated
configuration.

#### `env`

This directly contains env files that may be sourced by various shells in order
to setup up the shell for use with the deployment (e.g. allowing the frontend
to be invoked without specifying its full path).

Currently, only `bash` and `zsh` are supported.

#### `logs`

This directory contains service logs.

#### `plugins`

This directly contains attestation scheme plugins.

(note: if the deployment was crated with `-s` option, the plugins will in fact
be symlinks to their source locations.)

#### `signing`

This directly contains `skey.jwk`, the key that will be used by the
verification service to sign attestation results.

#### `stores`

This directory contains sqlite3 database for the endorsements, trust anchors,
and policies stores.

#### `systemd`

This directory contains systemd unit files for the Veraison services. It is
split into two sub-directories: `system` and `user`. The latter contains units
meant to be installed into the user-specific service manager (i.e. using
`systemctl --user`).

#### `launchd`

This directory contains launchd user agent files for the Veraison services.
