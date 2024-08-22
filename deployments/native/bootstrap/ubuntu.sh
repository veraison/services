#!/bin/sh
# Copyright 2024 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
sudo apt update
sudo apt install --yes git protobuf-compiler golang-1.22 make gettext sqlite3 openssl jq

sudo ln -s /usr/lib/go-1.22/bin/go /usr/local/bin/go
sudo ln -s /usr/lib/go-1.22/bin/gofmt /usr/local/bin/gofmt

wget https://dl.smallstep.com/cli/docs-cli-install/latest/step-cli_amd64.deb -O /tmp/step-cli_amd64.deb
sudo dpkg -i /tmp/step-cli_amd64.deb

go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
go install github.com/mitchellh/protoc-gen-go-json@latest

