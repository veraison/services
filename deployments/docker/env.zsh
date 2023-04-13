__VERAISON_DIR=$( cd -- "$( dirname -- "${(%):-%N}" )" &> /dev/null && pwd )

set -a
source $__VERAISON_DIR/deployment.cfg
set +a

alias veraison="$__VERAISON_DIR/veraison"
alias cocli="$__VERAISON_DIR/veraison -- cocli"
alias evcli="$__VERAISON_DIR/veraison -- evcli"
alias polcli="$__VERAISON_DIR/veraison -- polcli"
