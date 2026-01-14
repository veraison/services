# Copyright 2021-2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

SHELL := /bin/bash

# Pass this to sub-make
export GO111MODULE := on
export TOPDIR := $(realpath $(dir $(lastword $(MAKEFILE_LIST)))/..)

# Used to set the ServerVersion reported by  services
VERSION_FROM_GIT := $(shell $(TOPDIR)/scripts/get-veraison-version)

install:
