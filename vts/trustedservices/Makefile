# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

GOPKG := github.com/veraison/services/vts/trustedservices

INTERFACES += ../../plugin/imanager.go
INTERFACES += ../../plugin/ipluggable.go
INTERFACES += ../earsigner/iearsigner.go
INTERFACES += ../../handler/ievidencehandler.go
INTERFACES += ../../handler/istorehandler.go
INTERFACES += ../../kvstore/ikvstore.go
INTERFACES += ../policymanager/ipolicymanager.go

MOCKPKG := mocks

all-hook-pre test-hook-pre lint-hook-pre: _mocks
	$(MAKE) -C ../../proto protogen

include ../../mk/common.mk
include ../../mk/pkg.mk
include ../../mk/lint.mk
include ../../mk/test.mk
