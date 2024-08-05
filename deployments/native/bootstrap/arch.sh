#!/bin/sh
sudo pacman -Syy bash findutils grep sed openssl protobuf go make gettext sqlite3 step-cli jq
sudo ln -s /usr/bin/step-cli /usr/local/bin/step

go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
go install github.com/mitchellh/protoc-gen-go-json@latest

