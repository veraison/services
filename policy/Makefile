# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := test

GOPKG := github.com/veraison/services/policy

INTERFACES := ibackend.go

MOCKPKG := mocks

lint-hook-pre: _mocks
	make -C ../proto/ protogen

include ../mk/common.mk
include ../mk/pkg.mk
include ../mk/lint.mk
include ../mk/test.mk
