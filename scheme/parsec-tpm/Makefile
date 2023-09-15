# Copyright 2023 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := test

GOPKG := github.com/veraison/services/scheme/parsec-tpm
SRCS := $(wildcard *.go)

SUBDIR += plugin

# auto-generated
COPYRIGHT_FLAGS += --ignore */corim_test_vectors.go

include ../../mk/common.mk
include ../../mk/lint.mk
include ../../mk/pkg.mk
include ../../mk/subdir.mk
include ../../mk/test.mk
