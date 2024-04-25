# Copyright 2022-2024 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

SUBDIR := arm

all: all-hook-pre
clean:

all-hook-pre all-test-pre all-lint-pre:
	$(MAKE) -C ../../handler

include ../../mk/common.mk
include ../../mk/lint.mk
include ../../mk/test.mk
include ../../mk/subdir.mk
