# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

export TOPDIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

SUBDIR += builtin
SUBDIR += config
SUBDIR += handler
SUBDIR += kvstore
SUBDIR += log
SUBDIR += management
SUBDIR += plugin
SUBDIR += policy
SUBDIR += proto
SUBDIR += provisioning
SUBDIR += scheme
SUBDIR += verification
SUBDIR += vts
SUBDIR += vtsclient

COVERAGE_THRESHOLD := 60.0
# plugin coverage is low because it is mostly tested via plugin/test, a
# separate package (this is necessary due to to the nature of the code being
# tested. plugin/test coverage is low because it's purely test code).
IGNORE_COVERAGE += github.com/veraison/services/plugin
IGNORE_COVERAGE += github.com/veraison/services/plugin/test
# There is protobuf-generated stuff here, which skews coverage.
IGNORE_COVERAGE += github.com/veraison/services/handler

include mk/cover.mk

define __MAKEFILE_HELP
Available targets:

	test:          run unit tests
	integ-test:    run integration tests
	lint:          run the Go linter against the source
	coverage:      run a check to make sure that unit test coverage is
	               above a pre-determined threshold ($(COVERAGE_THRESHOLD)%)
	clean:         clean up build artefacts
	really-clean:  clean up deployment and integration-test related artefacts
	docker-deploy: create and start the docker deployment (docker must be
	               installed, and the user must be in the docker group)
endef
export __MAKEFILE_HELP

.PHONY: help
help:
	@echo "$$__MAKEFILE_HELP"

ifeq ($(filter help,$(MAKECMDGOALS)),help)
__NO_RECURSE = true
endif

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

ifeq ($(filter docker-deploy,$(MAKECMDGOALS)),docker-deploy)
__NO_RECURSE = true
endif

.PHONY: really-clean
really-clean:
	make -C integration-tests really-clean
	make -C deployments/docker really-clean

ifeq ($(filter really-clean,$(MAKECMDGOALS)),really-clean)
__NO_RECURSE = true
endif

.PHONY: integ-test
integ-test:
	make -C integration-tests test

ifeq ($(filter integ-test,$(MAKECMDGOALS)),integ-test)
__NO_RECURSE = true
endif

ifndef __NO_RECURSE
include mk/subdir.mk
endif
