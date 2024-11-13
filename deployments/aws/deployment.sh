#!/bin/bash
set -ueo pipefail

_error='\e[0;31mERROR\e[0m'
_this_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
_repo_root=$(realpath "${_this_dir}/../..")

set -a
source "${_this_dir}/deployment.cfg"
set +a

_script=${_this_dir}/bin/veraison

function help() {
	set +e
	local usage
	read -r -d '' usage <<-EOF
	Usage: deployment.sh [OPTIONS...] COMMAND [ARGS...]

	This script allows deploying Veraison to AWS.

	VPS and at least two subnets must exist and be specified via deployment.cfg.

	OPTIONS:

	Please note tht opitons MUST be specified before the command and arguments.

	  -h show this message and exist
	  -f force overwriting of existing artifacts
	  -v enable verbose output

	COMMANDS:

	help
	    Show this message and exit. The same as -h option.

	bootstrap
	    Bootstrap local system for running the deployment script. Install dependencies
	    and initialize the Python virtual enviroment for the script.

	bringup
	    Create a full Veraison deployment using configuration inside deployment.cfg.

	redeploy-stack
	    Delete and re-create the cloudformation stack using existing artifacts (DEB
	    package, AMI images, etc).

	teardown
	    Delete the existing deployment stack and all associated artificats (DEB package,
	    AMI images, etc).

	EOF
	set -e

	echo "$usage"
}

function bootstrap() {
	"${_repo_root}/deployments/debian/deployment.sh" bootstrap

	case $( uname -s ) in
		Linux)
			# shellcheck disable=SC2002
			local distrib_id
			distrib_id=$(head -n 1 </etc/lsb-release 2>/dev/null | \
				     cut -f2 -d= | tr -d \")

			case $distrib_id in
			Arch) sudo pacman -Syy packer ssh openssl;;
			Ubuntu)
				sudo apt update
				sudo apt --yes install curl openssl postgresql

				curl -fsSL https://apt.releases.hashicorp.com/gpg | sudo apt-key add -
				sudo apt-add-repository "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"
				sudo apt --yes install packer
				;;
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
			brew install packer postgresql
			;;
		*)
			echo -e "$_error: Boostrapping is currently only supported for Arch, Ubuntu, and MacOSX (via homebrew)."
			exit
			;;
	esac

	python -m venv "${VERAISON_AWS_VENV}"
	# shellcheck disable=SC1091
	source "${VERAISON_AWS_VENV}/bin/activate"
	pip install -r "${_this_dir}/misc/requirements.txt"

	set +e
	local message
	read -r -d '' message <<-EOF

	Enviroment for AWS deployment has been bootstraped. To activate it:

	    source ${_this_dir}/env/env.bash

	EOF
	set -e

	echo "$message"
}

function bringup() {
	_check_installed openssl
	_check_installed packer

	veraison configure --init \
		--vpc-id "${VERAISON_AWS_VPC_ID}" \
		--subnet-id "${VERAISON_AWS_SUBNET_ID}" \
		--rds-subnet-ids "${VERAISON_AWS_RDS_SUBNET_IDS}" \
		--admin-cidr "${VERAISON_AWS_ADMIN_CIDR}" \
		--region "${VERAISON_AWS_REGION}"

	veraison create-deb
	veraison create-key-pair
	veraison create-combined-image
	veraison create-keycloak-image
	veraison create-combined-stack

	veraison update-security-groups
	veraison create-certs
	veraison setup-rds
	veraison setup-keycloak --realm-file "${_this_dir}/misc/veraison-realm.json"
	veraison setup-services
}

function redeploy_stack() {
	_check_installed openssl

	veraison delete-stack combined
	veraison delete-certs

	veraison create-combined-stack
	veraison update-security-groups
	veraison create-certs
	veraison setup-rds
	veraison setup-keycloak --realm-file "${_this_dir}/misc/veraison-realm.json"
	veraison setup-services
}

function teardown() {
	veraison delete-stack combined
	veraison delete-certs
	veraison delete-image keycloak
	veraison delete-image combined
	veraison delete-key-pair
	veraison delete-deb
}

function veraison() {
	"${_script}" "${_script_flags[@]}" "${@}"
}

function _check_installed() {
	local what=$1

	if [[ "$(type -p "$what")" == "" ]]; then
		echo -e "$_error: $what executable must be installed to use this command."
		exit 1
	fi
}

_force=false
_verbose=false
_no_error=false

while getopts "hfNv" opt; do
	case "$opt" in
		h) help; exit 0;;
		f) _force=true;;
		N) _no_error=true;;
		v) _verbose=true;;
		*) break;;
	esac
done

shift $((OPTIND-1))
[ "${1:-}" = "--" ] && shift

_script_flags=(--deployment-name "${VERAISON_AWS_DEPLOYMENT}")
if [[ $_force == true ]]; then
	_script_flags+=(--force)
fi
if [[ $_verbose == true ]]; then
	_script_flags+=(--verbose)
fi
if [[ $_no_error == true ]]; then
	_script_flags+=(--no-error)
fi

_check_installed python

_command=$1; shift
_command=$(echo "$_command" | tr -- _ -)
case $_command in
	help) help;;
        bootstrap) bootstrap;;
	bringup) bringup;;
	redeploy-stack) redeploy_stack;;
	teardown) teardown;;
	*) echo -e "$_error: unexpected command: \"$_command\"";;
esac
# vim: set noet sts=8 sw=8:
