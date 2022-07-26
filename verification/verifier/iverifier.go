// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package verifier

import "github.com/veraison/services/proto"

type IVerifier interface {
	GetVTSState() (*proto.ServiceState, error)
	IsSupportedMediaType(mt string) (bool, error)
	SupportedMediaTypes() ([]string, error)
	ProcessEvidence(tenantID string, data []byte, mt string) ([]byte, error)
}
