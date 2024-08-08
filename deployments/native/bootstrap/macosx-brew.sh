#!/bin/sh

set -eux
set -o pipefail

brew install step coreutils gettext openssl sqlite3 protobuf jq
brew link --force gettext 

go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
go install github.com/mitchellh/protoc-gen-go-json@latest
