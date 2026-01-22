# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := test

GOPKG := github.com/veraison/services/scheme/veraison-client
SRCS := $(wildcard *.go)

INTERFACES := iveraisonchallengeresponseclient.go
INTERFACES += iveraisondiscoveryclient.go

MOCKPKG := mocks

lint-hook-pre: _mocks

SUBDIR += plugin

include ../../mk/common.mk
include ../../mk/lint.mk
include ../../mk/pkg.mk
include ../../mk/subdir.mk
include ../../mk/test.mk
