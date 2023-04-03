# Copyright 2023 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
.DEFAULT_TARGET: test

# Sets the user for the debug container. Set this to "root" to run  as root
# (not recommended for actually running tests, but allows you to make one-off
# modifications to the container's environment (e.g. installing additional
# tools) before changing to tavern (via "su tavern") to start running tests.
TEST_USER ?= tavern

SHELL = /bin/bash

THIS_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
include ../deployments/docker/src/load-config.mk

DEPLOYMENT_SRC_DIR := $(THIS_DIR)../deployments/docker/
SCHEME_DIR := $(THIS_DIR)../scheme/
DOCKER_DIR := $(THIS_DIR)docker/
TESTER_CONTEXT := $(CONTEXT_DIR)/tester

$(DEPLOYMENT_SRC_DIR).built/%:
	$(MAKE) -C $(DEPLOYMENT_SRC_DIR) $*


CONTAINER_FLAGS := --env-file  $(DEPLOYMENT_SRC_DIR)deployment.cfg --network veraison-net \
		   -v $(THIS_DIR):/integration-tests

DEPLOYMENT_DEPS := $(DEPLOYMENT_SRC_DIR).built/network \
		   $(DEPLOYMENT_SRC_DIR).built/vts-container \
		   $(DEPLOYMENT_SRC_DIR).built/provisioning-container \
		   $(DEPLOYMENT_SRC_DIR).built/verification-container

CLEANFILES := .pytest_cache utils/__pycache__ __generated__

.PHONY: image
image: .built/image

.built/image: docker/* $(DEPLOYMENT_SRC_DIR).built/deploy \
	      $(SCHEME_DIR)tpm-enacttrust/test/cmd/gen-token/gen-token
	mkdir -p $(TESTER_CONTEXT)
	cp $(SCHEME_DIR)tpm-enacttrust/test/cmd/gen-token/gen-token $(TESTER_CONTEXT)/gen-enacttrust-token
	cp $(DEPLOY_DEST)/utils/{cocli,evcli} $(TESTER_CONTEXT)/
	cp $(THIS_DIR)docker/* $(TESTER_CONTEXT)/
	docker build -t veraison/test $(TESTER_CONTEXT) \
		 --build-arg TESTER_UID=$(shell id -u) --build-arg TESTER_GID=$(shell id -g)
	@mkdir -p .built && touch .built/image

.PHONY: start-services
start-services: $(DEPLOYMENT_SRC_DIR).built/deployment
	@if [[ $(shell $(DEPLOYMENT_SRC_DIR)veraison short-status) != "running" ]]; then \
		$(DEPLOYMENT_SRC_DIR)veraison start; \
	fi

.PHONY: debug
debug: .built/image start-services
	docker container run --rm --interactive  --tty \
		$(CONTAINER_FLAGS) \
		--user $(TEST_USER)\
		--entrypoint /bin/bash \
		veraison/test

.PHONY: test
test: .built/image start-services
	docker container run --rm --tty $(CONTAINER_FLAGS) veraison/test

.PHONY: clean
clean:
	$(RM) -r $(CLEANFILES)

.PHONY: docker-clean
docker-clean:
	@if [[ $(shell $(DEPLOYMENT_SRC_DIR)veraison short-status) != "stopped" ]]; then \
		$(DEPLOYMENT_SRC_DIR)veraison stop; \
	fi
	docker image rm -f veraison/test
	$(RM) -r .built

.PHONY: really-clean
really-clean: clean docker-clean

$(SCHEME_DIR)tpm-enacttrust/test/cmd/gen-token/gen-token: \
	  $(shell find $(SCHEME_DIR)tpm-enacttrust -name \*.go)
	make -C $(SCHEME_DIR)tpm-enacttrust/test/cmd/gen-token

define __MAKEFILE_HELP
Available targets:

	help:            display this message
	test:            run integration tests
	clean:           remove artefects generated while running integration
	                 tests
	debug:           start a shell inside a test runner container.
	start-services:  start the Veraison services to be tested (this done
	                 automatically when executing the test target)
	image:           create the test runner docker image
	docker-clean:    remove the test runner docker image (also stops veraison
		         services)
	really-clean:    a combination of clean and docker-clean
endef
export __MAKEFILE_HELP

.PHONY: help
help:
	@echo "$$__MAKEFILE_HELP"
