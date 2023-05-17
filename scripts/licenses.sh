#!/bin/bash
# Copyright 2022-2023 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

set -e

type go-licenses &> /dev/null || go get github.com/google/go-licenses

MODULES+=(". github.com/veraison/services/...")

for module in "${MODULES[@]}"
do
  dir=$(echo "${module}" | cut -d' ' -f1)
  mod=$(echo "${module}" | cut -d' ' -f2)

  echo ">> retrieving licenses [ ${mod} ]"
  ( cd "${dir}" && go-licenses csv ${mod} )
done
