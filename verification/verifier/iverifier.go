// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package verifier

type IVerifier interface {
	// XXX this should return an error as well
	IsSupportedMediaType(mt string) bool
	// XXX this should return an error as well
	SupportedMediaTypes() []string
	ProcessEvidence(tenantID string, data []byte, mt string) ([]byte, error)
}
