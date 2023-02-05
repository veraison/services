# Copyright 2023 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
#
# variables:
# * COVERAGE_THRESHOLD  - a float expressing the precentage minimum threshold for
#                         acceptible test coverage.
# * IGNORE_COVERAGE - a list of go packages for which coverage threshold will
#                     not be checked.
#
#
# targets:
# * coverage - generate coverage reports and ensure they're above COVER_THRESHOLD
#              for everything in current dirtory.
#

COVERAGE_THRESHOLD ?= 80.0

IGNORE_FLAGS = $(foreach P,$(IGNORE_COVERAGE),--ignore $(P))

.PHONY: coverage
coverage:
	make TEST_ARGS="-short -cover" all test | grep "coverage:.*of statements"  | sort -u | \
		python scripts/check-coverage --threshold $(COVERAGE_THRESHOLD) $(IGNORE_FLAGS)

