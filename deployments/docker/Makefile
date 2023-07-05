# Copyright 2023 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
##############################################################################
# Config from the environment:
#
# Specifies the host name for the container started with "make debug". This can
# be used to substitute executables run in the container for one of the Veraison
# services on the network. E.g. DEBUG_HOST=vts-service
DEBUG_HOST ?= builder
#
# Additional flags passed to Docker when building various images. E.g. you can
# bypass local layer cache by specifying DOCKER_BUILD_FLAGS=--no-cache
DOCKER_BUILD_FLAGS ?=
##############################################################################

.DEFAULT_TARGET: all

SHELL = /bin/bash

THIS_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
include src/load-config.mk

SRC_DIR := $(THIS_DIR)src/
BUILDER_CONTEXT := $(CONTEXT_DIR)/builder

vts_FLAGS := -v $(STORES_VOLUME):/opt/veraison/stores
management_FLAGS := -v $(STORES_VOLUME):/opt/veraison/stores -p $(MANAGEMENT_PORT):$(MANAGEMENT_PORT)
provisioning_FLAGS := -p $(PROVISIONING_PORT):$(PROVISIONING_PORT)
verification_FLAGS := -p $(VERIFICATION_PORT):$(VERIFICATION_PORT)

ifneq ($(DEBUG_PORT),)
DEBUG_PORT_FLAG := -p $(DEBUG_PORT):$(DEBUG_PORT)
endif

.PHONY: all
all: builder build deploy deployment prune

.PHONY: builder
builder: .built/builder

.built/builder: $(SRC_DIR)builder.docker $(SRC_DIR)builder-dispatcher $(SRC_DIR)builder-bashrc \
	        $(THIS_DIR)../../go.mod
	mkdir -p $(BUILDER_CONTEXT)
	cp $(SRC_DIR)builder-dispatcher $(BUILDER_CONTEXT)/
	cp $(SRC_DIR)builder-bashrc $(BUILDER_CONTEXT)/
	cp $(THIS_DIR)../../go.{mod,sum} $(BUILDER_CONTEXT)/
	docker build $(DOCKER_BUILD_FLAGS) \
		-t veraison/builder $(BUILDER_CONTEXT) -f $(SRC_DIR)builder.docker \
		--build-arg BUILDER_UID=$(shell id -u) --build-arg BUILDER_GID=$(shell id -g)
	@mkdir -p .built && touch .built/builder

.PHONY: build
build: .built/build

.built/build: .built/builder
	docker container run --rm \
		-v $(THIS_DIR)../..:/veraison/build veraison/builder build
	@mkdir -p .built && touch .built/build

.PHONY: deploy
deploy: .built/deploy

.built/deploy: .built/build
	@# If DEPLOY_DEST contains /, assume a directory and make sure it exists.
	@if [[ "$(shell echo "$(DEPLOY_DEST)" | sed 's/[^\/]//g')" ]]; then \
		mkdir -p $(DEPLOY_DEST); \
	fi
	docker container run --rm \
		-v $(THIS_DIR)../..:/veraison/build -v $(DEPLOY_DEST):/veraison/deploy \
		veraison/builder deploy
	@mkdir -p .built && touch .built/deploy

.PHONY: debug
debug: .built/builder .built/network
	docker container run --rm \
		--network $(VERAISON_NETWORK) \
		-v $(STORES_VOLUME):/veraison/stores -v $(LOGS_VOLUME):/veraison/logs \
		-v $(THIS_DIR)../..:/veraison/build -v $(DEPLOY_DEST):/veraison/deploy \
		$(DEBUG_FLAGS) $(DEBUG_PORT_FLAG) --hostname $(DEBUG_HOST)\
		-i -t --user $(DEBUG_USER) --entrypoint /bin/bash \
		veraison/builder

.PHONY: services
services:
	@# Older versions of make (including on Ubuntu LTS) do not seem to always resolve the
	@# image targets (possibly because of the need to recursively resolve %,
	@# but I haven't looked too much into it). Recursively calling $(MAKE) here
	@# resolves the issue.
	$(MAKE) .built/vts-container .built/provisioning-container .built/verification-container \
		.built/management-container

.PHONY: vts
vts: deploy .built/vts-container

.PHONY: vts-image
vts-image: deploy .built/vts-image

.PHONY: provisioning
provisioning: deploy .built/provisioning-container

.PHONY: provisioning-image
provisioning-image: deploy .built/provisioning-image

.PHONY: verification
verification: deploy .built/verification-container

.PHONY: verification-image
verification-image: deploy .built/verification-image

.PHONY: management
management: deploy .built/management-container

.PHONY: management-image
management-image: deploy .built/management-image

.PHONY: network
network: .built/network

.built/%-image: $(SRC_DIR)%.docker $(DEPLOY_DEST)/*
	@# Before rebuilding the image, remove any containers that were based
	@# on the previous version.
	@# The "nil" ensures the command doesn't fail due to an empty argument list.
	docker rm -f $(shell docker ps -a -q --filter "ancestor=veraison/$*") nil 2>/dev/null
	@
	DOCKER_BUILDKIT=1 docker build $(DOCKER_BUILD_FLAGS) \
		-t veraison/$* $(DEPLOY_DEST) -f $(SRC_DIR)$*.docker
	@mkdir -p .built && touch .built/$*-image

.built/%-container: .built/%-image .built/network
	docker container create --tty --name $*-service --hostname $*-service \
		$($*_FLAGS) \
		--network $(VERAISON_NETWORK) \
		-v $(LOGS_VOLUME):/opt/veraison/logs veraison/$*
	@mkdir -p .built && touch .built/$*-container

.built/network:
	docker network create --driver bridge $(VERAISON_NETWORK)
	@mkdir -p .built && touch .built/network

.PHONY: manager
manager: .built/manager

.built/manager: $(SRC_DIR)manager.docker $(SRC_DIR)manager-dispatcher
	mkdir -p $(BUILDER_CONTEXT)
	cp $(SRC_DIR)manager-dispatcher $(DEPLOY_DEST)/
	docker build $(DOCKER_BUILD_FLAGS) \
		-t veraison/manager $(DEPLOY_DEST) -f $(SRC_DIR)manager.docker \
		 --build-arg MANAGER_UID=$(shell id -u) --build-arg MANAGER_GID=$(shell id -g)
	@mkdir -p .built && touch .built/manager

.PHONY: deployment
deployment: .built/deployment

.built/deployment: network services manager
	@mkdir -p .built && touch .built/deployment

.PHONY: prune
prune:
	docker image prune -f

.PHONY: clean
clean:
	if [[ "x$(shell docker image ls | grep veraison/builder)" != "x" ]]; then \
	docker container run --rm \
		-v $(THIS_DIR)../..:/veraison/build veraison/builder clean; \
		rm -f .built/build; \
	fi

.PHONY: docker-clean
docker-clean:
	@# If DEPLOY_DEST does not contain /, assume a volume and clean it up.
	@if [[ "$(shell echo "$(DEPLOY_DEST)" | sed 's/[^\/]//g')" == "" ]]; then \
		docker volume rm -f $(DEPLOY_DEST); \
	fi
	@# -f ensures exit code 0, even if image doesn't exist
	docker container rm -f vts-service provisioning-service verification-service management-service
	docker volume rm -f veraison-logs veraison-stores
	@# ubuntu uses an older version of docker without -f option for network; hence the || : cludge
	docker network rm $(VERAISON_NETWORK) || :
	docker image rm -f veraison/builder veraison/vts veraison/provisioning \
		veraison/verification veraison/manager
	rm -rf  .built

.PHONY: host-clean
host-clean:
	@# If DEPLOY_DEST contains /, assume a directory and clean it up.
	@if [[ "$(shell echo "$(DEPLOY_DEST)" | sed 's/[^\/]//g')" ]]; then \
		rm -rf $(DEPLOY_DEST); \
	fi
	rm -rf $(CONTEXT_DIR)

.PHONY: really-clean
really-clean: clean docker-clean host-clean

define __MAKEFILE_HELP
Available targets:

	help:         display this message.
	all:          create everything. See below for what "everything" means.
        debug:        start a shell inside a builder container connected to the
	              deployment network.
	clean:        clean Veraison Go build artefacts.
	docker-clean: delete the deployment's docker artefacts
	host-clean:   delete the temporary locations on the hosts that were
	              created in the process of building the deployment.
	really-clean: a combination of clean, docker-clean, and host-clean.

The following targets are built when building the "all" target, so they
typically don't need to be invoked directly:

	builder:            create the builder image that will be used to build
	                    veraison (also used for debug shells).
	build:              build Veraison using a builder container (does
	                    not require Go toolchain on the host). A context
	                    for the builder image will be created under
	                    $(BUILDER_CONTEXT).
	deploy:             create the deployment structure inside
	                    $(DEPLOY_DEST). This is used as the context for
	                    building service images.
        network:            create the docker network for the deployment
	manager:            create the manager image. This is used by the
	                    deployment frontend.
	services:           create the service containers (see below).
	vts:                create the vts-service container.
	vts-image:          create the veraison/vts image used to create the
	                    above.
	provisioning:       create the provisioning-service container.
	provisioning-image: create the veraison/provisioning image used to
	                    create the above.
	verification:       create the verification-service container.
	verification-image: create the veraison/verification image used to
	                    create the above.
	prune:              prune intermediate docker images.
endef
export __MAKEFILE_HELP

.PHONY: help
help:
	@echo "$$__MAKEFILE_HELP"
