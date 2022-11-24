# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
#
.DEFAULT_GOAL := all

.PHONY: lint
lint:

.PHONY: redshirt
redshirt: redshirt/redshirt.go
	make -C redshirt

.PHONY: trooper
trooper: trooper/trooper.go
	make -C trooper

.PHONY: powercell
powercell: powercell/powercell.go
	make -C powercell

.PHONY: gascartridge
gascartridge: gascartridge/gascartridge.go
	make -C gascartridge

.PHONY: all
all: redshirt trooper

.PHONY: clean
clean:
	rm -rf bin/

include ../../mk/common.mk
include ../../mk/test.mk
