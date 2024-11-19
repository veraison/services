SHELL := /bin/bash

THIS_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
SCRIPT := $(THIS_DIR)/deployment.sh

.PHONY: bootstrap
bootstrap:
	source $(THIS_DIR)/env/env.bash
	$(SCRIPT) bootstrap

.PHONY: deploy
deploy:
	@if [[ "$(AWS_ACCOUNT_CFG)" == "" ]]; then \
		echo "ERROR: please define AWS_ACCOUNT_CFG (see README)"; \
		exit 1; \
	fi
	source $(THIS_DIR)/env/env.bash $(AWS_ACCOUNT_CFG) && $(SCRIPT) bringup

.PHONY: really-clean
really-clean:
	source $(THIS_DIR)/env/env.bash
	$(SCRIPT) -N teardown
