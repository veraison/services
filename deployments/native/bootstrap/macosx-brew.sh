#!/bin/sh
# Copyright 2024-2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

set -eux
set -o pipefail

brew install jose coreutils gettext openssl sqlite3 protobuf jq
brew link --force gettext 

go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
go install github.com/mitchellh/protoc-gen-go-json@latest
go install github.com/veraison/corim-store/cmd/corim-store@latest
