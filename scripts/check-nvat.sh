#!/usr/bin/env bash
# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

if [[ "$(type -p cc)" == "" ]]; then
	echo "C compiler not installed"
	exit 1
fi

if ! (echo -e '#include <nvat.h>\n' | cc -E -x c - >/dev/null 2>&1); then
	echo "nvat.h not found"
	exit 1
fi


