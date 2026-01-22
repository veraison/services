// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package veraisonclient

import "github.com/veraison/apiclient/verification"

// IVeraisonChallengeResponseClient is an interface for dealing with Veraison's
// apiclient/verification ChallengeResponseConfig objects
type IVeraisonChallengeResponseClient interface {
	Run() ([]byte, error)
	SetNonce(nonce []byte) error
	SetSessionURI(u string) error
	SetEvidenceBuilder(eb verification.EvidenceBuilder) error
	SetDeleteSession(v bool)
	SetNonceSz(nonceSz uint) error
	SetIsInsecure(v bool)
	SetCerts(paths []string)
}
