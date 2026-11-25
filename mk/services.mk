# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

TOPDIR ?= $(CURDIR)
SERVICES_BIN = $(TOPDIR)/cmd/bin/

GO_TAGS_FLAG = $(if $(BUILD_TAGS_CSV),-tags=$(BUILD_TAGS_CSV))

BUILD_FLAGS = $(GO_TAGS_FLAG) \
	      -ldflags "-X 'github.com/veraison/services/pkg/config.Version=$(VERSION_FROM_GIT)'" \
	      -o $(SERVICES_BIN)

PROTO_FILES := $(shell find $(TOPDIR)/pkg -name '*.proto')
PB_GO_FILES := $(PROTO_FILES:.proto=.pb.go)

export PATH := $(shell go env GOPATH)/bin:$(PATH)

%.pb.go: %.proto
	cd $(<D) && protoc --proto_path=. \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(<F)

.PHONY: protogen
protogen: $(PB_GO_FILES)

.PHONY: services
services: protogen
	CGO_ENABLED=1 go build $(BUILD_FLAGS) $(TOPDIR)/cmd/...

.PHONY: clean-services
clean-services:
	rm -rf $(SERVICES_BIN)
