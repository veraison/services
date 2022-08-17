# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := all

GOPKG := github.com/veraison/services/provisioning/cmd
CMD := provisioning-service
SRCS := main.go

CMD_DEPS := $(wildcard ../api/*.go)
CMD_DEPS += $(wildcard ../decoder/*.go)
CMD_DEPS += $(wildcard ../storeclient/*.go)

cmd-hook-pre test-hook-pre lint-hook-pre:
	$(MAKE) -C ../../proto protogen
	$(MAKE) -C ../decoder protogen

include ../../mk/common.mk
include ../../mk/cmd.mk
include ../../mk/test.mk
include ../../mk/lint.mk
include ../../mk/pkg.mk
