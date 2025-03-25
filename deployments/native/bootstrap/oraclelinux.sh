#!/bin/sh
# Copyright 2024 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

arch=$(arch)
req_go_version="1.23.0"
install_go="no"
go_pkg=""
step_pkg=""

case $arch in
	x86_64)
		go_pkg=go1.23.4.linux-amd64.tar.gz
		step_pkg=step-cli_amd64.rpm
		;;
	aarch64)
		go_pkg=go1.23.4.linux-arm64.tar.gz
		step_pkg=step-cli_arm64.rpm
		;;
	*)
		echo -e "Unsupported architecture for Oracle Linux"
		;;
esac

go=$(command -v go)
if [ "$go" == "" ]; then
	install_go="yes"
fi

if [ "$install_go" == "no" ]; then
	cur_go_version=`go version | { read _ _ cur_go_version _; echo ${cur_go_version#go}; }`
	if [ "$(printf '%s\n' "$goversion" "req_go_version" | sort -V | head -n1)" != "req_go_version" ]; then
		install_go="yes"		
	fi
fi

if [ "$install_go" == "yes" ]; then
	wget https://go.dev/dl/$go_pkg -O /tmp/$go_pkg
	sudo tar -C /usr/local -xzf /tmp/$go_pkg
	sudo ln -s /usr/local/go/bin/go /usr/local/bin/go
	sudo ln -s /usr/local/go/bin/gofmt /usr/local/bin/gofmt
fi

sudo dnf install -y --enablerepo=ol9_codeready_builder git protobuf protobuf-devel gettext sqlite openssl jq
sudo dnf install -y https://dl.smallstep.com/cli/docs-cli-install/latest/$step_pkg

go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
go install github.com/mitchellh/protoc-gen-go-json@latest

