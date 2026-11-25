// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package verification

import "github.com/veraison/services/pkg/vts"

//go:generate mockgen -source iverifier.go -destination mock_verifier_test.go -package verification
type IVerifier interface {
	ServerState() (*vts.State, error)
	IsSupportedMediaType(mt string) (bool, error)
	SupportedMediaTypes() ([]string, error)
	ProcessEvidence(tenantID string, nonce []byte, data []byte, mt string) ([]byte, error)
}
