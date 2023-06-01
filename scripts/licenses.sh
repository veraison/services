#!/bin/bash
# Copyright 2022-2023 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

set -e

type go-licenses &> /dev/null || go get github.com/google/go-licenses

MODULES+=(". github.com/veraison/services/...")

function header() {
  echo "Go package,License file URL,License type"
}

for module in "${MODULES[@]}"
do
  dir=$(echo "${module}" | cut -d' ' -f1)
  mod=$(echo "${module}" | cut -d' ' -f2)

  >&2 echo ">> retrieving licenses [ ${mod} ]"
  header
  ( cd "${dir}" && go-licenses csv ${mod} )
done
