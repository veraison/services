// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package trustedservices

import "github.com/veraison/services/proto"

type ITrustedServices interface {
	Init() error
	Close() error
	Run() error

	proto.VTSServer
}
