#!/bin/bash
set -eo pipefail

_error='\e[0;31mERROR\e[0m'
_this_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

_repo_root=$(realpath "${_this_dir}/../..")

_package_name="veraison-services"

function create_rpm() {
	## Uncomment the following to set build version
	#export VERAISON_BUILD_VERSION="version"

	local version=$("${_repo_root}/scripts/get-veraison-version")
	export _VERAISON_VERSION=${version}

	export GOOS=linux

	## Create an archive of the source needed for rpmbuild; cleanup
	## build files from previous builds
	if [ -d "/tmp/${_package_name}-${_VERAISON_VERSION}" ]; then
		mv /tmp/${_package_name}-${_VERAISON_VERSION} /tmp/${_package_name}-${_VERAISON_VERSION}-old
	fi

	if [ -f "/tmp/${_package_name}-${_VERAISON_VERSION}.tar.gz" ]; then
		mv /tmp/${_package_name}-${_VERAISON_VERSION}.tar.gz /tmp/${_package_name}-${_VERAISON_VERSION}.tar.gz.old
	fi

	mkdir -p /tmp/${_package_name}-${_VERAISON_VERSION}
	cp -r ${_repo_root}/. /tmp/${_package_name}-${_VERAISON_VERSION}/
	pushd /tmp
	tar -cvzf ${_package_name}-${_VERAISON_VERSION}.tar.gz --exclude=perf --exclude=.git --exclude=.github ${_package_name}-${_VERAISON_VERSION}/
	popd

	## Create RPM build tree, if not present. Cleanup up source files
	## from previous builds
	rpmdev-setuptree

	if [ -f "${HOME}/rpmbuild/SOURCES/${_package_name}-${_VERAISON_VERSION}.tar.gz" ]; then
		rm -f ${HOME}/rpmbuild/SOURCES/${_package_name}-${_VERAISON_VERSION}.tar.gz
	fi

	if [ -d "${HOME}/rpmbuild/BUILD/${_package_name}-${_VERAISON_VERSION}" ]; then
		rm -rf ${HOME}/rpmbuild/BUILD/${_package_name}-${_VERAISON_VERSION}
	fi

	## Kickoff RPM build
	mv /tmp/${_package_name}-${_VERAISON_VERSION}.tar.gz ${HOME}/rpmbuild/SOURCES
	cp ${_this_dir}/${_package_name}.spec.template ${HOME}/rpmbuild/SPECS/${_package_name}.spec
	sed -i -e "s/_VERSION_/${_VERAISON_VERSION}/g" ${HOME}/rpmbuild/SPECS/${_package_name}.spec
	rpmbuild -ba ${HOME}/rpmbuild/SPECS/${_package_name}.spec

	## Cleanup temporary build files
	rm -rf /tmp/${_package_name}-${_VERAISON_VERSION}
	rm -f /tmp/${_package_name}-${_VERAISON_VERSION}.tar.gz

	echo "done."
}

function help() {
	set +e
	local usage
	read -r -d '' usage <<-EOF
	Usage: deployment.sh [OPTIONS...] COMMAND [ARGS...]

	This script allows packaging a Veraison deployment as .rpm package suitable
	for installation on Fedora-like Linux distros (such as RHEL and Oracle Linux). 

	OPTIONS:

	Please note that opitons MUST be specified before the command and arguments.

	  -h show this message and exist

	COMMANDS:

	help
	    Show this message and exit. The same as -h option.

	create-rpm [DIR]
	    Create a RPM package under DIR. If DIR is not specified, /tmp will be
            used. Upon successful completion, it will contain the .rpm package and a 
            subdirectory with the sources used to create the package. This command 
            relies on the "native" deployment to create the package sources.
	EOF
	set -e

	echo "$usage"
}

while getopts "h" opt; do
	case "$opt" in
		h) help; exit 0;;
		*) break;;
	esac
done

_command=$1; shift
_command=$(echo "$_command" | tr -- _ -)
case $_command in 
	help) help;;
	create-rpm) create_rpm "$1";;
	*) echo -e "$_error: unexpected command: \"$_command\"";;
esac
# vim: set noet sts=8 sw=8:
