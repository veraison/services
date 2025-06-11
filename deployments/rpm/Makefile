# Copyright 2025 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
.DEFAULT_TARGET: rpm

SHELL = /bin/bash

THIS_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

PACKAGE_DEST ?= /tmp

.PHONY: rpm
rpm:
	$(THIS_DIR)/deployment.sh create-rpm $(PACKAGE_DEST)

.PHONY: really-clean
really-clean:
	rm -rf $(PACKAGE_DEST)/veraison-rpm-package
