// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package trustedservices

import "github.com/veraison/services/proto"

// MediaTypeMap maintains the association between the media types supported by
// the active plugins and the attestation scheme they implement.
// The contents of this map are computed dynamically by querying each loaded
// plugin via their GetSupportedMediaTypes interface.
type MediaTypeMap map[string]proto.AttestationFormat

type ITrustedServices interface {
	Init() error
	Close() error
	Run() error

	proto.VTSServer
}
