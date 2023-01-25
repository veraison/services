# Copyright 2023 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
#

all:

schemes.gen.go: $(shell find ../scheme/ -type f -name \*.go)
	python3 ../scripts/gen-schemes ../scheme/ > schemes.gen.go

include ../mk/common.mk
include ../mk/pkg.mk
include ../mk/lint.mk
include ../mk/test.mk
