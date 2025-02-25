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
	  -l delete logs (only applies to teardown command)
	  -v enable verbose output

	COMMANDS:

	help
	    Show this message and exit. The same as -h option.

	bootstrap
	    Bootstrap local system for running the deployment script. Install dependencies
	    and initialize the Python virtual enviroment for the script.

	bringup
	    Create a full Veraison deployment using configuration inside deployment.cfg.

	teardown [-l]
	    Delete the existing deployment stack and all associated artificats (DEB package,
	    AMI images, etc). By default, CloudWatch logs will be kept. To also delete the
	    logs, use -l flag.

	delete-logs | clear-logs
	    Delete logs stored in CloudWatch.

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
			distrib_id=$(cat /etc/os-release | grep -w ID | \
				     cut -f2 -d= | tr -d \")

			case $distrib_id in
			arch) sudo pacman -Syy packer ssh;;
			Ubuntu)
				sudo apt update
				sudo apt --yes install curl postgresql

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
	_check_installed packer

	# shellcheck disable=SC2153
	veraison configure --init \
		--admin-cidr "${VERAISON_AWS_ADMIN_CIDR}" \
		--vpc-cidr "${VERAISON_AWS_VPC_CIDR}" \
		--region "${VERAISON_AWS_REGION}" \
		--dns-name "${VERAISON_AWS_DNS_NAME}" \
		--vts-port "${VTS_PORT}" \
		--provisioning-port "${PROVISIONING_PORT}" \
		--verification-port "${VERIFICATION_PORT}" \
		--management-port "${MANAGEMENT_PORT}" \
		--keycloak-port "${KEYCLOAK_PORT}" \
		--keycloak-version "${KEYCLOAK_VERSION}" \
		--keycloak-admin "${KEYCLOAK_ADMIN}" \
		--scaling-min-size "${SCALING_MIN_SIZE}" \
		--scaling-max-size "${SCALING_MAX_SIZE}" \
		--scaling-cpu-util-target "${SCALING_CPU_UTIL_TARGET}" \
		--scaling-request-count-target "${SCALING_REQUEST_COUNT_TARGET}" \
		--cw-log-retention-days "${CLOUDWATCH_LOG_RETENTION_DAYS}" \
		--iam-logger-role-name "${IAM_LOGGER_ROLE_NAME}" \
		--iam-instance-profile-name "${IAM_INSTANCE_PROFILE_NAME}" \
		--iam-permission-boundary-arn "${IAM_PERMISSION_BOUNDARY_ARN}"

	veraison create-deb
	veraison create-key-pair

	veraison create-vpc-stack

	veraison create-sentinel-image
	veraison create-rds-stack
	veraison update-security-groups # need to access sentinel to set up RDS
	veraison setup-rds

	veraison create-services-image
	veraison create-keycloak-image

	veraison create-services-stack
}

function teardown() {
	set +e
	veraison delete-stack services
	veraison delete-stack rds
	veraison delete-stack vpc

	veraison delete-image keycloak
	veraison delete-image services
	veraison delete-image sentinel

	veraison delete-key-pair
	veraison delete-deb

	if [[ ${_delete_logs} == true ]]; then
		delete_logs
	fi

	set -e
}

function delete_logs() {
	veraison delete-logs
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
_delete_logs=false

while getopts "hflNv" opt; do
	case "$opt" in
		h) help; exit 0;;
		f) _force=true;;
		N) _no_error=true;;
		l) _delete_logs=true;;
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
	teardown) teardown;;
	delete-logs | clear-logs) delete_logs;;
	*) echo -e "$_error: unexpected command: \"$_command\"";;
esac
# vim: set noet sts=8 sw=8:
