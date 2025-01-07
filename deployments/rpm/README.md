This directory contains scripts and other resources for creating .rpm
packages for installation on Fedora-like distros (such as RHEL and
Oracle Linux). The build process involves creating a native deployment
and then packaging it up using `rpmbuild`. Veraison services run as
`VERAISON_USER` as specified in `deployment.cfg`, which defaults to
`veraison`. If this user isn't available, RPM creates it. 

## Dependencies

In addition to [dependencies for the native 
deployment](../native/README.md#dependencies), `rpm-build` must be installed. To
install all dependencies to build an rpm, run

```sh
make bootstrap
```

## Building the package

The location where the package will be built is specified with `PACKAGE_DEST`
environment variable. It will default to `/tmp` if not set. To build the
package simply do

```sh
make rpm
```
This will create the following RPM package
`${PACKAGE_DEST}/veraison_VERSION_ARCH/rpmbuild/RPMS/ARCH/veraison-VERSION.FLA.ARCH.rpm`
where `VERSION` is the Veraison version as reported by the
[`get-veraison-version`](../scripts/get-veraison-version) script,
`ARCH` is the architecture of your system as reported by `arch`, and
`FLA` is the distro flavor such as el8 and el9.

## Install the package

The following command installs the RPM package

```sudo dnf install ${PACKAGE_DEST}/veraison_VERSION_ARCH/rpmbuild/RPMS/ARCH/veraison-VERSION.FLA.ARCH.rpm```
