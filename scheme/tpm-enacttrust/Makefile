# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
.DEFAULT_GOAL := test

GOPKG := github.com/veraison/services/scheme/tpm-enacttrust
SRCS := $(wildcard *.go)

SUBDIR += plugin

# auto-generated
COPYRIGHT_FLAGS += --ignore */test_vectors.go

include ../../mk/common.mk
include ../../mk/lint.mk
include ../../mk/pkg.mk
include ../../mk/subdir.mk
include ../../mk/test.mk
