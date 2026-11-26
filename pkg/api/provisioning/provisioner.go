// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package provisioning

import (
	"context"

	"github.com/veraison/services/pkg/fault"
	"github.com/veraison/services/pkg/util"
	"github.com/veraison/services/pkg/vts"
)

type Provisioner struct {
	ctx     context.Context
	service vts.IService
}

func NewProvisioner(ctx context.Context, service vts.IService) IProvisioner {
	return &Provisioner{
		ctx:     ctx,
		service: service,
	}
}

func (o *Provisioner) ServerVersion() (string, error) {
	vtsState, err := o.service.GetServiceState(o.ctx)
	if err != nil {
		return "", err
	}
	version := vtsState.Version
	return version, nil
}

func (o *Provisioner) IsSupportedMediaType(mt string) (bool, error) {
	normalizedMediaType, err := util.NormalizeMediaType(mt)
	if err != nil {
		return false, fault.BadInput("validation failed for %s (%w)", mt, err)
	}

	serviceState, err := o.service.GetServiceState(o.ctx)
	if err != nil {
		return false, err
	}

	for _, supportedMediaType := range serviceState.ProvisioningMediaTypes {
		if supportedMediaType == normalizedMediaType {
			return true, nil
		}
	}

	return false, nil
}

func (o *Provisioner) SupportedMediaTypes() ([]string, error) {
	serviceState, err := o.service.GetServiceState(o.ctx)
	if err != nil {
		return nil, err
	}

	return serviceState.ProvisioningMediaTypes, nil
}

func (o *Provisioner) SubmitEndorsements(tenantID string, data []byte, mt string) error {
	ctx := vts.SetTenantID(o.ctx, tenantID)
	return o.service.SubmitEndorsements(ctx, mt, data)
}

func (o *Provisioner) SetEndorsementsState(tenantID string, data []byte, setActive bool) error {
	ctx := vts.SetTenantID(o.ctx, tenantID)
	return o.service.SetEndorsementsState(ctx, data, setActive)
}
