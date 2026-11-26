# Copyright 2021-2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

SHELL = /bin/bash
TOPDIR = $(CURDIR)

SCHEME_LOADER ?= plugins
ifeq ($(SCHEME_LOADER),builtin)
	BUILD_TAGS += builtin
endif

NVAT_OK := $(shell $(TOPDIR)/scripts/check-nvat.sh; echo $$?)
ifeq ($(NVAT_OK),0)
	BUILD_TAGS += nvat
endif

include mk/common.mk
include mk/services.mk
include mk/plugins.mk
include mk/test.mk
include mk/lint.mk
include mk/cover.mk

.PHONY: build
build: services plugins

.PHONY: clean
clean: clean-services clean-plugins

.DEFAULT_GOAL := help

define __MAKEFILE_HELP
Available targets:

	services:      build services executables, puting them under cmd/bin/
	plugins:       build plugin executables, puting them under plugins/bin/
	scheme-<NAME>  build the specified plugn's executable; the name is taken
	               from the executlable's name without the .plugin exension
	               e.g. scheme-arm-cca
	store-<NAME>  build the specified plugn's executable; the name is taken
	               from the executlable's name without the .plugin exension
	               e.g. store-corim-store
	build:         combination of services and plugins above
	test:          run unit tests
	integ-test:    run integration tests
	lint:          run the Go linter against the source
	coverage:      run a check to make sure that unit test coverage is
	               above a threshold defined in .testcoverage.yml
	report:        open html coverage report in default browser
	clean:         clean up build artefacts
	really-clean:  clean up deployment and integration-test related artefacts
	docker-deploy: create and start the docker deployment (docker must be
	               installed, and the user must be in the docker group)
	bootstrap:     install required dependencies (only works on Arch, Ubuntu,
	               and MacOSX using homebrew)
	native-deploy: create and start the native deployment
	deb:           create .deb package for installation on Debian or Ubuntu
	rpm:           create .rpm package for installation on Fedora-like distros
endef
export __MAKEFILE_HELP

.PHONY: help
help:
	@echo "$$__MAKEFILE_HELP"

define __DOCKER_DEPLOY_MESSAGE

=============================================================================
Veraison has been deployed on the local system via Docker. If you're using
bash you can access to the frontend via the following command:

	source deployments/docker/env.bash

(there is an equivalent env.zsh for zsh). You can then view frontend help via

	veraison -h

In addition to the veraison frontend, env.bash will also set up aliases for
cocli, evcli, and polcli utilities.

=============================================================================
endef
export __DOCKER_DEPLOY_MESSAGE

.PHONY: docker-deploy
docker-deploy:
	make -C deployments/docker all
	deployments/docker/veraison start
	@echo "$$__DOCKER_DEPLOY_MESSAGE"

.PHONY: really-clean
really-clean: clean clean-mocks
	make -C integration-tests really-clean
	make -C deployments/debian really-clean
	make -C deployments/docker really-clean
	make -C deployments/native really-clean

.PHONY: integ-test
integ-test:
	make -C integration-tests test

.PHONY: bootstrap
bootstrap:
	make -C deployments/native bootstrap

VERAISON_ROOT ?= $(HOME)/veraison-deployment
export VERAISON_ROOT

define __NATIVE_DEPLOY_MESSAGE

=============================================================================
Veraison has been deployed natively on the local system. If you're using
bash, you can access to the frontend via the following command:

	source ${VERAISON_ROOT}/env/env.bash

(there is an equivalent env.zsh for zsh). You can then view frontend help via

	veraison -h

In addition to the veraison frontend, env.bash will also set up aliases for
cocli, evcli, and polcli utilities.

=============================================================================
endef
export __NATIVE_DEPLOY_MESSAGE

.PHONY: native-deploy
native-deploy:
	make -C deployments/native quick-deploy
	@if [[ "$(shell type -p systemctl)" != "" && \
		"$(shell type -p powershell.exe)" == "" ]]; then \
		$(VERAISON_ROOT)/bin/veraison start-services; \
	fi
	@echo "$$__NATIVE_DEPLOY_MESSAGE"

.PHONY: deb
deb:
	make -C deployments/debian deb

.PHONY: rpm
rpm:
	make -C deployments/rpm rpm
