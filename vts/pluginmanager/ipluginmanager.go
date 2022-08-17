// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package pluginmanager

import (
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme"
)

type ISchemePluginManager interface {
	Init() error
	Close() error

	LookupByMediaType(mediaType string) (scheme.IScheme, error)
	LookupByAttestationFormat(format proto.AttestationFormat) (scheme.IScheme, error)
	SupportedVerificationMediaTypes() ([]string, error)
}
