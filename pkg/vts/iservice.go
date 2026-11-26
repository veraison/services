// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package vts

import (
	"context"

	"github.com/veraison/services/pkg/scheme"
)

type IService interface {
	GetServiceState(ctx context.Context) (*State, error)
	GetAttestation(ctx context.Context, evidence *scheme.Evidence) ([]byte, error)
	GetEndorsements(ctx context.Context, mediaType string, query string) ([]byte, error)
	SubmitEndorsements(ctx context.Context, mediaType string, data []byte) error
	SetEndorsementsState(ctx context.Context, data []byte, active bool) error
}
