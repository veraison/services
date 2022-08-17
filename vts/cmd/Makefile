# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := all

GOPKG := github.com/veraison/services/vts/cmd
CMD := vts-service
SRCS := main.go

CMD_DEPS += $(wildcard ../trustedservices/*.go)
CMD_DEPS += $(wildcard ../pluginmanager/*.go)

include ../../mk/common.mk
include ../../mk/cmd.mk
include ../../mk/test.mk
include ../../mk/lint.mk
include ../../mk/pkg.mk
