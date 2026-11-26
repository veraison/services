// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package coserv

import (
	"context"

	"github.com/veraison/services/pkg/vts"
)

type EndorsementDistributor struct {
	ctx     context.Context
	service vts.IService
}

func NewEndrosementDistributor(
	ctx context.Context,
	service vts.IService,
) IEndorsementDistributor {
	return &EndorsementDistributor{
		ctx:     ctx,
		service: service,
	}
}

func (o *EndorsementDistributor) ServerState() (*vts.State, error) {
	return o.service.GetServiceState(o.ctx)
}

func (o *EndorsementDistributor) GetEndorsements(
	tenantID string,
	query string,
	mediaType string,
) ([]byte, error) {
	ctx := vts.SetTenantID(o.ctx, tenantID)
	return o.service.GetEndorsements(ctx, mediaType, query)
}
