# Copyright 2024 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
.DEFAULT_TARGET: quick-deploy

SHELL = /bin/bash

THIS_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

VERAISON_ROOT ?= $(HOME)/veraison-deployment
export VERAISON_ROOT

.PHONY: bootstrap
bootstrap:
	$(THIS_DIR)/deployment.sh bootstrap

.PHONY: check
check:
	$(THIS_DIR)/deployment.sh check-requirements

.PHONY: quick-deploy
quick-deploy:
	$(THIS_DIR)/deployment.sh -e quick-init-all

.PHONY: really-clean
really-clean:
	$(VERAISON_ROOT)/bin/veraison stop-systemd 2>/dev/null || /bin/true
	$(VERAISON_ROOT)/bin/veraison disable-systemd 2>/dev/null || /bin/true
	$(VERAISON_ROOT)/bin/veraison kill-tmux 2>/dev/null || /bin/true
	rm -rf $(VERAISON_ROOT)
