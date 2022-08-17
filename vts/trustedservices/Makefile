# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

GOPKG := github.com/veraison/services/vts/trustedservices

all-hook-pre test-hook-pre lint-hook-pre:
	$(MAKE) -C ../../proto protogen

include ../../mk/common.mk
include ../../mk/pkg.mk
include ../../mk/lint.mk
include ../../mk/test.mk
