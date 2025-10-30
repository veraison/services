#!/bin/sh
# Copyright 2024-2025 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

arch=$(dpkg --print-architecture)
go_ver=1.24

sudo apt update
sudo apt install --yes git protobuf-compiler golang-${go_ver} make gettext sqlite3 openssl jq jose

sudo ln -s /usr/lib/go-${go_ver}/bin/go /usr/local/bin/go
sudo ln -s /usr/lib/go-${go_ver}/bin/gofmt /usr/local/bin/gofmt

go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
go install github.com/mitchellh/protoc-gen-go-json@latest

