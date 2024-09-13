This directory contains scripts and other resources for creating .deb packages
for installation on Debian or Ubuntu systems. This involves first creating a
native deployment, and then packaging it up using `dpkg`.

## Dependencies

In addition to [dependencies for the native 
deployment](../native/README.md#dependencies), `dpkg` must be installed. If you
are on a Debian or Ubuntu system, `dpkg` will already be present as it the
package manager for your system. If you are on Arch, you can install it via

```sh
# on Arch
pacman -S dpkg
```

If you are on another system, you will need to find how to install `dpkg` on
your own (first check that it is not the package manager for the system, then
search the system's standard packages; if all else fails -- duckduckgo/brave is
your friend).

## Building the package

The location where the package will be built is specified with `PACKAGE_DEST`
environment variable. It will default to `/tmp` if not set. To build the
package simply do

```sh
make deb
```

This will create
`${PACKAGE_DEST}/veraison_deb_package/veraison_VERSION_ARCH.deb`, where `VERSION`
is the Veraison version as reported by the
[`get-veraison-version`](../scripts/get-veraison-version) script, and `ARCH` is
the architecture of your system as reported by `dpkg --print-architecture`.

Alongside the package, there will be a subdirectory with the same name but
without the .deb suffix that contains the "sources" used to build the package.
