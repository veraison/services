# This file sets up the enviroment for use of Veraison AWS CLI frontend.
# It should be sourced (not executed!). Optionally, a path to an extra config
# file can be provided as an argument. Settings in this file will be used in
# addition to deployment.cfg, and it should use the same syntax. See ../misc/arm.cfg
# for an example.
# E.g.
#
#   $ source deployments/aws/env/env.bash deployments/aws/misc/arm.cfg
#
# (note: arm.cfg contains settings specific to Arm's Veraison development AWS account.
# unless you're a developer from Arm and have access to that account, you will need
# to provide your own version of that file.)
#
# If you want to source this file from your own script, and you're not planning on
# specifying the extra cfg, you will need to add "--" as an argument (otherwise the
# arguments from your script will be propagated).
# E.g.
#
#   # inside your script
#   source deployments/aws/env/env.bash --
#
_this_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
_deployment_root="${_this_dir}/.."
_deployment_cfg="${_deployment_root}/deployment.cfg"

set -a
# shellcheck source=../deployment.cfg
source "${_deployment_cfg}"

if [[ "${1}" != "" && "${1}" != "--" ]]; then
	# shellcheck disable=SC1090
	source "${1}"
fi
set +a

if [[ -f "${VERAISON_AWS_VENV}/bin/activate" ]]; then
	# shellcheck disable=SC1091
	source "${VERAISON_AWS_VENV}/bin/activate"
fi

export PATH="${_deployment_root}/bin":${PATH}
