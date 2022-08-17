# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := all

GOPKG := github.com/veraison/services/verification/cmd
CMD := verification-service
SRCS := main.go

CMD_DEPS += $(wildcard ../api/*.go)
CMD_DEPS += $(wildcard ../sessionmanager/*.go)
CMD_DEPS += $(wildcard ../verifier/*.go)

include ../../mk/common.mk
include ../../mk/cmd.mk
include ../../mk/test.mk
include ../../mk/lint.mk
include ../../mk/pkg.mk
