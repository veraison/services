# Copyright 2025 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := test

GOPKG := github.com/veraison/services/scheme/amd-kds-coserv
SRCS := $(wildcard *.go)

SUBDIR += plugin

include ../../mk/common.mk
include ../../mk/lint.mk
include ../../mk/pkg.mk
include ../../mk/subdir.mk
include ../../mk/test.mk
