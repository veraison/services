_this_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
_deployment_root="${_this_dir}/.."
_deployment_cfg="${_deployment_root}/deployment.cfg"

set -a
# shellcheck source=../deployment.cfg
source "${_deployment_cfg}"
set +a

# shellcheck disable=SC1091
source "${VERAISON_AWS_VENV}/bin/activate"

export PATH="${_deployment_root}/bin":${PATH}
