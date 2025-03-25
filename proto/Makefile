# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := all

PROTOSRCS += token.proto
PROTOSRCS += evidence.proto
PROTOSRCS += appraisal_context.proto
PROTOSRCS += state.proto
PROTOSRCS += vts.proto
PROTOSRCS += status.proto
PROTOSRCS += endorsement_query.proto

lint-hook-pre: protogen
	protolint lint $(PROTOSRCS)

test:

clean: ; $(RM) protogen

include ../mk/common.mk
include ../mk/proto.mk
include ../mk/lint.mk
