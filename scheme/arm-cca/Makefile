# Copyright 2022-2024 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
.DEFAULT_GOAL := test

GOPKG := github.com/veraison/services/scheme/arm-cca
SRCS := $(wildcard *.go)

SUBDIR += plugin

include ../../mk/common.mk
include ../../mk/lint.mk
include ../../mk/pkg.mk
include ../../mk/subdir.mk
include ../../mk/test.mk
