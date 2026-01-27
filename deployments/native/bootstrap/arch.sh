#!/bin/sh
# Copyright 2024-2025 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
sudo pacman -Syy bash findutils grep sed openssl protobuf go make gettext sqlite3 jose jq

go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
go install github.com/mitchellh/protoc-gen-go-json@latest
go install github.com/veraison/corim-store/cmd/corim-store@latest

