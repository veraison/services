#!/bin/bash
set -eo pipefail

_error='\e[0;31mERROR\e[0m'
_this_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
_deb_src=${_this_dir}/debian
_repo_root=$(realpath "${_this_dir}/../..")
_version=$("${_repo_root}/scripts/get-veraison-version")


function bootstrap() {
	"${_repo_root}/deployments/native/deployment.sh" bootstrap

	case $( uname -s ) in
		Linux)
			# shellcheck disable=SC2002
			local distrib_id
			distrib_id=$(head -n 1 </etc/lsb-release 2>/dev/null | \
				     cut -f2 -d= | tr -d \")

			case $distrib_id in
			Arch) sudo pacman -Syy dpkg ;;
			Ubuntu) ;;
			*)
				echo -e "$_error: Boostrapping is currently only supported for Arch and Ubuntu."
				exit
				;;
			esac
			;;
		Darwin)
			if ! type brew > /dev/null; then
				echo -e "$_error: homebrew (https://brew.sh) must be installed."
				exit 1
			fi
			brew install dpkg
			;;
		*)
			echo -e "$_error: Boostrapping is currently only supported for Arch, Ubuntu, and MacOSX (via homebrew)."
			exit
			;;
	esac
}

function create_deb() {
	_check_installed dpkg
	_check_installed envsubst

	local work_dir=${1:-/tmp}
	local arch; arch="$(dpkg --print-architecture)"
	local pkg_dir=${work_dir}/veraison_${_version}_${arch}

	set -a
	source "${_this_dir}/deployment.cfg"
	set +a

	export VERAISON_ROOT=/opt/veraison
	export DEPLOYMENT_DEST=${pkg_dir}${VERAISON_ROOT}
	export VTS_HOST=$VERAISON_HOST
	export PROVISIONING_HOST=$VERAISON_HOST
	export VERIFICATION_HOST=$VERAISON_HOST
	export MANAGEMENT_HOST=$VERAISON_HOST

	rm -rf "${pkg_dir}"
	"${_repo_root}/deployments/native/deployment.sh" quick-init-all

	mkdir -p "${pkg_dir}/DEBIAN"
	cp "${_deb_src}"/{postinst,prerm} "${pkg_dir}/DEBIAN/"
	chmod 0775 "${pkg_dir}"/DEBIAN/{postinst,prerm}
	export _VERAISON_VERSION=${_version}
	envsubst < "${_deb_src}/control.template" > "${pkg_dir}/DEBIAN/control"

	dpkg --build "${pkg_dir}"

	echo "done."
}

function help() {
	set +e
	local usage
	read -r -d '' usage <<-EOF
	Usage: deployment.sh [OPTIONS...] COMMAND [ARGS...]

	This script allows packaging a Veraison deployment as .deb package suitable
	for installation on Debian and derivatives (such as Ubuntu). 

	OPTIONS:

	Please note tht opitons MUST be specified before the command and arguments.

	  -h show this message and exist

	COMMANDS:

	help
	    Show this message and exit. The same as -h option.

        bootstrap
            Set up the enviroment for creating the deployment, installing any
            necessary dependencies.

	create-deb [DIR]
	    Create a Debian package under DIR. If DIR is not specified, /tmp will be
            used. Upon successful completion, it will contain the .deb package and a 
            subdirectory with the sources used to created the package. This command 
            relies on the "native" deployment to creates the package sources.
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
	create-deb) create_deb "$1";;
	*) echo -e "$_error: unexpected command: \"$_command\"";;
esac
# vim: set noet sts=8 sw=8:
