# Copyright 2024 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
.DEFAULT_TARGET: deb

SHELL = /bin/bash

THIS_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

PACKAGE_DEST ?= /tmp

.PHONY: deb
deb:
	$(THIS_DIR)/deployment.sh create-deb $(PACKAGE_DEST)

.PHONY: really-clean
really-clean:
	rm -rf $(PACKAGE_DEST)/veraison-deb-package
