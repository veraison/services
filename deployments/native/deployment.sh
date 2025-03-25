#!/bin/bash
# Copyright 2024 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
# shellcheck disable=SC2155,SC2086

_THIS_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
_ERROR='\e[0;31mERROR\e[0m'
_WARN='\e[0;33mWARNING\e[0m'
_INSTALL=$(type -p install)
_SERVICES=(vts provisioning verification management)

set -e

set -a
source ${_THIS_DIR}/deployment.cfg
VERAISON_CERTS=${VERAISON_CERTS:-${_THIS_DIR}/example/certs}
set +a

ROOT_DIR=${_THIS_DIR}/../..
BOOTSTRAP_DIR=${_THIS_DIR}/bootstrap
EXAMPLE_DIR=${_THIS_DIR}/example
SRC_BIN_DIR=${_THIS_DIR}/bin
SRC_CONFIG_DIR=${_THIS_DIR}/config
SRC_ENV_DIR=${_THIS_DIR}/env
SRC_SIGNING_DIR=${EXAMPLE_DIR}/signing
SRC_SYSTEMD_DIR=${_THIS_DIR}/systemd
SRC_LAUNCHD_DIR=${_THIS_DIR}/launchd

DEPLOYMENT_BIN_DIR=${DEPLOYMENT_DEST}/bin
DEPLOYMENT_PLUGINS_DIR=${DEPLOYMENT_DEST}/plugins
DEPLOYMENT_CERTS_DIR=${DEPLOYMENT_DEST}/certs
DEPLOYMENT_LOGS_DIR=${DEPLOYMENT_DEST}/logs
DEPLOYMENT_STORES_DIR=${DEPLOYMENT_DEST}/stores
DEPLOYMENT_CONFIG_DIR=${DEPLOYMENT_DEST}/config
DEPLOYMENT_ENV_DIR=${DEPLOYMENT_DEST}/env
DEPLOYMENT_SIGNING_DIR=${DEPLOYMENT_DEST}/signing
DEPLOYMENT_SYSTEMD_DIR=${DEPLOYMENT_DEST}/systemd
DEPLOYMENT_LAUNCHD_DIR=${DEPLOYMENT_DEST}/launchd

# VERAISON_ROOT_OVERRIDE will be used by veraison frontend script when
# generating certificates/sigining key as part of the deployment creation.
export VERAISON_ROOT_OVERRIDE=${DEPLOYMENT_DEST}

function check_requirements() {
	set +e

	local ok=true

	# note: there is no explicit test for bash, as we're inside a bash script, so
	# if bash was not installed, we would not even get here.
	# There is also no explicit test for coreutils executables (printf, head, etc..) as
	# those will be available on pretty much any functional UNIX system.

	if [[ "$(type -p grep)" == "" ]]; then
		echo -e "$_ERROR: grep must be installed."
		exit 1
	fi

	local _sort=$(type -p gsort)
	if [[ $_sort == "" ]]; then
		_sort="sort"
	fi

	if ! $_sort --help | grep -- -V &>/dev/null; then
		echo -e "$_ERROR: GNU sort must be installed (on MacOSX install coreutils)"
		exit 1
	fi

	if [[ "$(type -p go)" == "" ]]; then
		echo -e "$_ERROR: Go toolchain (at least version 1.23) must be installed."
		exit 1
	fi

	if ! printf '%s\n' 1.23 "$(go version | grep -o -E '[0-9.]+' | head -n1)" | \
		$_sort -C -V; then

		echo -e "$_ERROR: Go version must be at least 1.23."
		exit 1
	fi

	if [[ "$(type -p envsubst)" == "" ]]; then
		echo -e "$_ERROR: envsubst must be installed."
		exit 1
	fi

	if [[ "$(type -p sed)" == "" ]]; then
		echo -e "$_ERROR: sed must be installed."
		exit 1
	fi

	if [[ "$(type -p find)" == "" ]]; then
		echo -e "$_ERROR: find must be installed."
		exit 1
	fi

	if [[ "$(type -p sqlite3)" == "" ]]; then
		echo -e "$_WARN: sqlite3 (needed by some commands) is not installed"
		ok=false
	fi

	if [[ "$(type -p openssl)" == "" ]]; then
		echo -e "$_WARN: openssl (needed by some commands) is not installed"
		ok=false
	fi

	if [[ "$(type -p make)" == "" ]]; then
		echo -e "$_WARN: make (needed by some commands) is not installed"
		ok=false
	fi

	if [[ "$(type -p step)" == "" ]]; then
		echo -e "$_WARN: step (needed by some commands) is not installed"
		ok=false
	fi

	if [[ $ok == true ]]; then
		echo "ok"
	fi

	set -e
}

function bootstrap() {
	case $( uname -s ) in
		Linux)
			# shellcheck disable=SC2002
			local distrib_id=$(cat /etc/os-release | grep -w ID | cut -f2 -d= | tr -d \")

			case $distrib_id in
			arch) ${BOOTSTRAP_DIR}/arch.sh;;
			ubuntu) ${BOOTSTRAP_DIR}/ubuntu.sh;;
			ol) ${BOOTSTRAP_DIR}/oraclelinux.sh;;
			*)
				echo -e "$_ERROR: Boostrapping is currently only supported for Arch, Ubuntu and Oracle Linux. For other systems, please see one of the scripts in ${BOOTSTRAP_DIR}, and adapt the commmand to your system."
				exit
				;;
			esac
			;;
		Darwin)
			if ! type brew > /dev/null; then
				echo -e "$_ERROR: homebrew (https://brew.sh) must be installed."
				exit 1
			fi
			${BOOTSTRAP_DIR}/macosx-brew.sh
			;;
		*)
			echo -e "$_ERROR: Boostrapping is currently only supported for Arch, Ubuntu, and MacOSX (via homebrew)."
			echo -e "For other systems, please see one of the scripts in ${BOOTSTRAP_DIR}, and adapt the commmand to your system."
			exit
			;;
	esac

}

function build() {
	make -C ${ROOT_DIR} COMBINED_PLUGINS=1
}

function create_deployment() {
	local mode=$1

	_init_deployment_dir

	if  [[ $mode == "symlink" ]]; then
		_symlink_bins
	else
		_deploy_bins
	fi

	_deploy_frontend

	_deploy_services_config

	_deploy_env

	if [[ $_force_systemd == true ]]; then
		_deploy_systemd_units
	elif [[ $_force_launchd == true ]]; then
		_deploy_launchd_units
	else
		case $( uname -s ) in
			Linux) _deploy_systemd_units;;
			Darwin) _deploy_launchd_units;;
		esac
	fi
}

function create_root_cert() {
	local _f=""
	if [[ $_force == true ]]; then
		_f="-f"
	fi

	${DEPLOYMENT_BIN_DIR}/veraison $_f gen-root-cert "$1"
}

function init_certs() {
	local mode=$1
	local template=$2
	local root_cert_path=$3
	local root_cert_key_path=$4

	if [[ $mode == "copy" ]]; then
		_deploy_certs
	else
		_gen_certs $template $root_cert_path $root_cert_key_path
	fi
}

function init_signing_key() {
	local mode=$1

	if [[ $mode == "copy" ]]; then
		_deploy_signing_key
	else
		_gen_signing_key
	fi
}

function init_sqlite_stores() {
	${DEPLOYMENT_BIN_DIR}/veraison init-sqlite-stores ${DEPLOYMENT_STORES_DIR}
}

function init_clients() {
	_init_client evcli github.com/veraison/evcli/v2@0d3a093
	_init_client cocli github.com/veraison/cocli@4eada925
	_init_client pocli github.com/veraison/pocli@2fa24ea3
}

function quick_init_all(){
	local bins_mode=$1
	local cnk_mode=$2
	local template=$3
	local root_cert_path=$4
	local root_cert_key_path=$5

	if [[ "$template" == "" ]]; then
		template='localhost'
	fi

	build
	create_deployment $bins_mode
	init_certs  $cnk_mode $template $root_cert_path $root_cert_key_path
	init_signing_key $cnk_mode
	init_sqlite_stores
	init_clients
}

function setup_keycloak() {
	local path=$(realpath $1)
	local names_string=$2
	local ca_cert=$3
	local ca_cert_key=$4
	local port=${5:-11111}

	local cert_path="${path%/}/conf/server"
	local i=0
	local dns=()

	IFS=',' read -ra names <<< "$names_string"
	for n in "${names[@]}"; do
		dns[i]="DNS:${n}"
		i=$((i+1))
	done

	local san=$(IFS=','; echo "${dns[*]}"; IFS=$' \t\n')

	openssl ecparam -name prime256v1 -genkey -noout -out "${cert_path}.key.pem"
	openssl req -new -key "${cert_path}.key.pem" -out "${cert_path}.csr" \
		-subj "/CN=${names_string%%,*}"
	openssl x509 -req -in "${cert_path}.csr" -CA ${ca_cert} -CAkey ${ca_cert_key} \
		-CAcreateserial -out "${cert_path}.crt.pem" -days 3650 -sha256 \
		-extfile <(echo "subjectAltName = ${san}")

	rm "${cert_path}.csr"

	keytool -genkeypair -storepass password -storetype PKCS12 -keyalg RSA -keysize 2048 \
		-dname "CN=server" -alias server -ext "SAN:c=DNS:localhost,IP:127.0.0.1" \
		-keystore "${path}/conf/server.keystore"

	mkdir -p  "${path}/data/import"
	cp "${EXAMPLE_DIR}/keycloak/veraison-realm.json" "${path}/data/import"

	echo "https-port = ${port}" >> "${path}/conf/keycloak.conf"
	echo "hostname = ${names_string%%,*}" >> "${path}/conf/keycloak.conf"
}

function help() {
	set +e
	read -r -d '' usage <<-EOF
	Usage: deployment.sh [OPTIONS] COMMAND [ARGS]

	This script creates and initializes a deployment of Veraison services on this
	machine using the native tools (must be installed -- see check-requirements
	command below).

	The root directory for the deployment is taken from the VERAISON_ROOT
	environment variable. This directory will be created if it does not already
	exist. The default for this and other variable that control the process and
	the configuration of the resulting deployment is specified inside
	deployment.cfg in the same directory as this script.

	OPTIONS:

	Please note that options MUST be specified before the command and arguments.

	  -h show this message and exit
	  -e copy the example certs and signing key into the deployment, instead of
	     generating new ones
	  -f force overwriting of existing files
	  -s create symlinks rather than copying binaries; useful during development,
	     so that binaries do not need to be re-deployed after being re-compiled

	Some options only apply to certain commands; if an option is applicable to a
	command, it will be listed for that command alongside its name and arguments.

	COMMANDS:

	help
	    Show this message and exit. The same as -h option.

	boostrap
	    Run a bootstrap script that installs all required dependencies. Currently,
	    only supported for Arch and Ubuntu.

	build
	    Build Veraison services with combined plugins.

	check-requirements | check-reqs
	    Check that the local system has all the requirements for this script
	    installed, reporting an error if it doesn't.

	[-s] deploy
	    Create an uninitialized deployment under VERAISON_ROOT, setting up the
	    directory structure, copying the services and plugins executables, and
	    creating configuration based on deployment.cfg. If the -s option is used,
	    executables will be symlinked rather than copied.

	    This command assumes that Veraison services have already been built using
	    combined plugins configuration. This command will not copy/symlink split
	    plugins (those that end with "-handler.plugin").

	    This command does not install client executables or their configuration
	    (see init-client command below).

	[-f] create-root-cert [SUBJECT]
	    Creates a self-signed certificate called rootCA.crt (and the associted key,
	    rootCA.key) with the specified SUBJECT under
	    ${VERAISON_ROOT}/certs/.
	    If the subject is not specified, then it defaults to "/O=Veraison".

	    This command will not overwrite the existing rootCA.crt unless -f option
	    is specified.

	[-e] [-f] init-certificates | init-certs [NAME_TEMPLATE ROOT_CERT ROOT_CERT_KEY]
	    Initialize x509 certificates for the deployment, either by generating new
	    ones, or copying the examples (if the -e option is specified).

	    When generating new certificates, the user must specify the NAME_TEMPLATE
	    that will be used to populate the Common Name and Subject Alternative
	    Names, as well as paths to a root certificate and corresponding key that
	    will be used to sign the generated certificates.

	    NAME_TEMPLATE is a string of comma-separated (no spaces) names, where each
	    name may contain "@@" as the placeholder that will be replaced with the
	    name of the service when generating a certificate for that service. The
	    first name of the list will used to populate the Common Name.

	    For example, the NAME_TEMPLATE

	        @@-service,localhost

	    will result in verification service certificate with the Common Name
	    "verification-service", and the Subject Alternative Name extension of
	    "DNS:verification-service, DNS:localhost".

	    This command will not overwrite any existing certificates unless -f option
	    is specified.

	init-clients
	    Install executables and configuration for Veraison services command line
	    clients: cocli, evcli, an pocli.

	[-e] [-f] init-signing-key
	    Initialize the signing key that will be used by verification service to
	    sign attestation results. If -e option is used, then the example signing
	    key will be copied into they deployment. Otherwise, a new EC P-256 key
	    will be generated.

	init-stores
	    Initialize the sqlite3 stores for the endorsements, trust anchors, and
	    policies. If the stores have already been initialized, this is a no-op,
	    the existing data in the stores will not be touched.

	[-e] [-s] quick-init-all [NAME_TEMPLATE ROOT_CERT ROOT_CERT_KEY]
	    Create a fully-initialized deployment with a single command. This broadly
	    equivalent to executing the following sequence:

	        ./deployment.sh build
	        ./deployment.sh [-s] deploy
	        ./deployment.sh init-clients
	        ./deployment.sh init-stores
	        ./deployment.sh [-e] init-certs [NAME_TEMPLATE ROOT_CERT ROOT_CERT_KEY]
	        ./deployment.sh [-e] init-signing-key

	setup-keycloak PATH NAMES ROOT_CERT ROOT_CERT_KEY [PORT]
	    Setup a fresh standalone Keycloak server located at PATH to work with the
	    Veraison deployment. This includes generating server certificates
	    using the specified NAMES (must be a comma-separated list of names
	    to be included in the Subject Alternate Names in the cert), and
	    signed with CA cert and corresponding key specified by ROOT_CERT and
	    ROOT_CERT_KEY. The server will be configured to run on port PORT
	    (defaults to 11111 if not specified).

	    Please see the README for more details.

	Note: for commands that contain dashes (e.g. init-stores), underscores may also
	      be used (e.g. init_stores).
	EOF

	echo "$usage"
	set -e
}

function _init_deployment_dir() {
	mkdir -p ${DEPLOYMENT_BIN_DIR}
	mkdir -p ${DEPLOYMENT_PLUGINS_DIR}
	mkdir -p ${DEPLOYMENT_CERTS_DIR}
	mkdir -p ${DEPLOYMENT_LOGS_DIR}
	mkdir -p ${DEPLOYMENT_CONFIG_DIR}
	mkdir -p ${DEPLOYMENT_ENV_DIR}
	mkdir -p ${DEPLOYMENT_SIGNING_DIR}
	mkdir -p ${DEPLOYMENT_STORES_DIR}
	case $( uname -s ) in
		Linux) mkdir -p ${DEPLOYMENT_SYSTEMD_DIR};;
		Darwin) mkdir -p ${DEPLOYMENT_LAUNCHD_DIR};;
	esac
}

function _deploy_systemd_units() {
	mkdir -p ${DEPLOYMENT_SYSTEMD_DIR}/user
	mkdir -p ${DEPLOYMENT_SYSTEMD_DIR}/system

	for service in "${_SERVICES[@]}"; do
		cat ${SRC_SYSTEMD_DIR}/user/veraison-${service}.service.template | envsubst > \
			${DEPLOYMENT_SYSTEMD_DIR}/user/veraison-${service}.service

		cat ${SRC_SYSTEMD_DIR}/system/veraison-${service}.service.template | envsubst > \
			${DEPLOYMENT_SYSTEMD_DIR}/system/veraison-${service}.service
	done
}

function _deploy_launchd_units() {
	mkdir -p ${DEPLOYMENT_LAUNCHD_DIR}

	for service in "${_SERVICES[@]}"; do
		cat ${SRC_LAUNCHD_DIR}/com.veraison-project.${service}.plist.template | envsubst > \
			${DEPLOYMENT_LAUNCHD_DIR}/com.veraison-project.${service}.plist
	done
}


function _deploy_certs() {
	for service in "${_SERVICES[@]}"; do
		cp ${VERAISON_CERTS}/${service}.{crt,key} ${DEPLOYMENT_CERTS_DIR}
	done

	cp ${VERAISON_CERTS}/rootCA.crt ${DEPLOYMENT_CERTS_DIR}
}

function _gen_certs() {
	local template=$1
	local root_cert_path=$2
	local root_key_path=$3

	local _f=""
	if [[ $_force == true ]]; then
		_f="-f"
	fi

	${DEPLOYMENT_BIN_DIR}/veraison $_f gen-service-certs $template \
		$root_cert_path $root_key_path
}

function _deploy_signing_key() {
	cp ${SRC_SIGNING_DIR}/skey.jwk ${DEPLOYMENT_SIGNING_DIR}/skey.jwk
}

function _gen_signing_key() {
	local _f=""
	if [[ $_force == true ]]; then
		_f="-f"
	fi

	${DEPLOYMENT_BIN_DIR}/veraison $_f gen-signing-key
}

function _deploy_services_config() {
	mkdir -p ${DEPLOYMENT_CONFIG_DIR}/services/
	cat ${SRC_CONFIG_DIR}/services.yaml.template | envsubst > \
		${DEPLOYMENT_CONFIG_DIR}/services/config.yaml
}

function _deploy_env() {
	for f in env.bash env.zsh; do
		cat ${SRC_ENV_DIR}/$f | envsubst > ${DEPLOYMENT_ENV_DIR}/$f
	done
}

function _symlink_bins() {
	exes=(
		"${ROOT_DIR}/provisioning/cmd/provisioning-service/provisioning-service"
		"${ROOT_DIR}/verification/cmd/verification-service/verification-service"
		"${ROOT_DIR}/management/cmd/management-service/management-service"
		"${ROOT_DIR}/vts/cmd/vts-service/vts-service"
	)

	local _f=""
	if [[ $_force == true ]]; then
		_f="-f"
	fi

	for path in "${exes[@]}"; do
		chmod +x "$path"
		ln -s $_f "$path" "${DEPLOYMENT_BIN_DIR}/$(basename $path)"
	done

	find "${ROOT_DIR}/scheme/bin/" -name '*.plugin' -print0 | grep -z -v handler |
	    while IFS= read -r -d '' path; do
		chmod +x "$path"
		ln -s $_f "$path" "${DEPLOYMENT_PLUGINS_DIR}/$(basename $path)"
	    done
}

function _deploy_bins() {
	$_INSTALL -m 0755 ${ROOT_DIR}/provisioning/cmd/provisioning-service/provisioning-service \
		${ROOT_DIR}/verification/cmd/verification-service/verification-service \
		${ROOT_DIR}/management/cmd/management-service/management-service \
		${ROOT_DIR}/vts/cmd/vts-service/vts-service \
		${DEPLOYMENT_BIN_DIR}

	find "${ROOT_DIR}/scheme/bin/" -name '*.plugin' -print0 | grep -z -v handler |
	    while IFS= read -r -d '' path; do
		    $_INSTALL -m 0755 "$path" "${DEPLOYMENT_PLUGINS_DIR}/$(basename $path)"
	    done
}

function _deploy_frontend {
	local veraison_root=${VERAISON_ROOT//\//\\\/}

	cat ${SRC_BIN_DIR}/veraison | \
		sed -e "s/_veraison_root=.*/_veraison_root=${veraison_root}/" > \
		${DEPLOYMENT_BIN_DIR}/veraison

	chmod 0755 ${DEPLOYMENT_BIN_DIR}/veraison
}

function _init_client() {
	local client=$1
	local install_spec=$2

	GOBIN=${DEPLOYMENT_BIN_DIR} go install ${install_spec}

	mkdir -p ${DEPLOYMENT_CONFIG_DIR}/${client}/
	echo "creating ${DEPLOYMENT_CONFIG_DIR}/${client}/config.yaml"
	cat ${SRC_CONFIG_DIR}/${client}.yaml.template | envsubst > \
		${DEPLOYMENT_CONFIG_DIR}/${client}/config.yaml
}

OPTIND=1

_force=false
_binaries="install"
_certs_and_keys="generate"
_force_systemd=false
_force_launchd=false

while getopts "hefsLS" opt; do
	case "$opt" in
		h) help; exit 0;;
		e) _certs_and_keys="copy";;
		f) _force=true;;
		L) _force_launchd=true;;
		s) _binaries="symlink";;
		S) _force_systemd=true;;
		*) break;;
	esac
done

if [[ ($_force_systemd == true) && ($_force_launchd == true) ]]; then
	echo "ERROR: cannot specify -S and -L  at the same time"
	exit 1
fi

shift $((OPTIND-1))
[ "${1:-}" = "--" ] && shift

command=$1; shift
command=$(echo $command | tr -- _ -)
case $command in
    help) help;;
    bootstrap) bootstrap;;
    build) build;;
    check-requirements | check-reqs)  check_requirements;;
    create-root-cert) create_root_cert "$1";;
    deploy) create_deployment "$_binaries";;
    init-certificates | init-certs) init_certs "$_certs_and_keys" "$1" "$2" "$3";;
    init-clients) init_clients;;
    init-signing-key) init_signing_key "$_certs_and_keys";;
    init-stores) init_sqlite_stores;;
    quick-init-all) quick_init_all "$_binaries" "$_certs_and_keys" "$1" "$2" "$3";;
    setup-keycloak) setup_keycloak "$1" "$2" "$3" "$4" "$5";;
    *) echo -e "$_ERROR: unexpected command: \"$command\"";;
esac
