# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := all

GOPKG := github.com/veraison/services/provisioning/cmd/provisioning-service
CMD := provisioning-service
SRCS := main.go

CMD_DEPS := $(wildcard ../../api/*.go)
CMD_DEPS += $(wildcard ../../.../handler/*.go)

cmd-hook-pre test-hook-pre lint-hook-pre:
	$(MAKE) -C ../../../proto protogen

include ../../../mk/common.mk
include ../../../mk/cmd.mk
include ../../../mk/test.mk
include ../../../mk/lint.mk
include ../../../mk/pkg.mk
