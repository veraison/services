# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := all

PROTOSRCS := result.proto
PROTOSRCS += attestation_format.proto
PROTOSRCS += param.proto
PROTOSRCS += token.proto
PROTOSRCS += evidence.proto
PROTOSRCS += appraisal_context.proto
PROTOSRCS += vts.proto

lint-hook-pre: protogen

lint lint-extra test:

clean: ; $(RM) protogen

include ../mk/common.mk
include ../mk/proto.mk
include ../mk/lint.mk
