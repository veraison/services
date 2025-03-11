#!/bin/bash
set -eo pipefail

_error='\e[0;31mERROR\e[0m'
_this_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

_repo_root=$(realpath "${_this_dir}/../..")
_version=$("${_repo_root}/scripts/get-veraison-version")


function bootstrap() {
	sudo dnf install -y rpm-build
	"${_repo_root}/deployments/native/deployment.sh" bootstrap
}

function create_rpm() {
	_check_installed rpmbuild

	local work_dir=${1:-/tmp}
	local arch; arch="$(arch)"
	local pkg_dir=${work_dir}/veraison_${_version}_${arch}

	set -a
	source "${_this_dir}/deployment.cfg"
	set +a

	export VERAISON_ROOT=${VERAISON_ROOT}
	export DEPLOYMENT_DEST=${pkg_dir}${VERAISON_ROOT}
	export VTS_HOST=$VERAISON_HOST
	export PROVISIONING_HOST=$VERAISON_HOST
	export VERIFICATION_HOST=$VERAISON_HOST
	export MANAGEMENT_HOST=$VERAISON_HOST

	export _VERAISON_VERSION=${_version}

	export GOOS=linux

	rm -rf "${pkg_dir}"
	"${_repo_root}/deployments/native/deployment.sh" -S quick-init-all

	mkdir -p ${pkg_dir}/rpmbuild/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS}
	tar -C ${DEPLOYMENT_DEST} -cvzf veraison-${_VERAISON_VERSION}.tar.gz .
	mv veraison-${_VERAISON_VERSION}.tar.gz ${pkg_dir}/rpmbuild/BUILD/
	cp veraison.spec.template ${pkg_dir}/rpmbuild/BUILD/veraison.spec

	sed -i -e "s/_VERSION_/${_VERAISON_VERSION}/g" ${pkg_dir}/rpmbuild/BUILD/veraison.spec
	sed -i -e "s/_VERAISON_USER_/${VERAISON_USER}/g" ${pkg_dir}/rpmbuild/BUILD/veraison.spec
	sed -i -e "s/_VERAISON_GROUP_/${VERAISON_GROUP}/g" ${pkg_dir}/rpmbuild/BUILD/veraison.spec

	rpmbuild --define "_topdir ${pkg_dir}/rpmbuild" -bb ${pkg_dir}/rpmbuild/BUILD/veraison.spec

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

        bootstrap
            Set up the enviroment for creating the deployment, installing any
            necessary dependencies.

	create-rpm [DIR]
	    Create a RPM package under DIR. If DIR is not specified, /tmp will be
            used. Upon successful completion, it will contain the .rpm package and a 
            subdirectory with the sources used to create the package. This command 
            relies on the "native" deployment to create the package sources.
	EOF
	set -e

	echo "$usage"
}

function _check_installed() {
	local what=$1

	if [[ "$(type -p "$what")" == "" ]]; then
		echo -e "$_error: $what executable must be installed to use this command."
		exit 1
	fi
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
        bootstrap) bootstrap;;
	create-rpm) create_rpm "$1";;
	*) echo -e "$_error: unexpected command: \"$_command\"";;
esac
# vim: set noet sts=8 sw=8:
