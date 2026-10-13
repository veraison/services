#!/bin/bash
set -euo pipefail

distro="ubuntu$(. /etc/os-release && printf '%s' "$VERSION_ID" | tr -d .)"
case "$(dpkg --print-architecture)" in
    amd64) repo_arch=x86_64 ;;
    arm64) repo_arch=sbsa ;;
    *) echo "unsupported architecture: $(dpkg --print-architecture)" >&2; exit 1 ;;
esac
keyring="${RUNNER_TEMP:-/tmp}/cuda-keyring.deb"
curl --fail --location --silent --show-error \
  "https://developer.download.nvidia.com/compute/cuda/repos/${distro}/${repo_arch}/cuda-keyring_1.1-1_all.deb" \
  --output "$keyring"
sudo dpkg --install "$keyring"
sudo apt-get update
sudo apt-get install --yes libnvat-dev
