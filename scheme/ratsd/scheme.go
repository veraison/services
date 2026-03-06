// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package ratsd

import (
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"go.uber.org/zap"
)

var Descriptor = handler.SchemeDescriptor{
	Name:          "RATSD",
	VersionMajor:  1,
	VersionMinor:  0,
	CorimProfiles: []string{""},
	EvidenceMediaTypes: []string{
		`application/cmw-collection+cbor; eat_profile="tag:github.com,2025:veraison/ratsd/cmw"`,
	},
}

type Implementation struct {
	logger *zap.SugaredLogger
}

func NewImplementation() *Implementation {
	return &Implementation{
		logger: log.Named(Descriptor.Name),
	}
}
