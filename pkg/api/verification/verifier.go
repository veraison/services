// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package verification

import (
	"context"
	"errors"

	"github.com/veraison/services/pkg/fault"
	"github.com/veraison/services/pkg/scheme"
	"github.com/veraison/services/pkg/util"
	"github.com/veraison/services/pkg/vts"
)

var ErrInputParam = errors.New("invalid input parameter")

type Verifier struct {
	ctx     context.Context
	service vts.IService
}

func NewVerifier(ctx context.Context, service vts.IService) IVerifier {
	return &Verifier{
		ctx:     ctx,
		service: service,
	}
}

func (o *Verifier) ServerState() (*vts.State, error) {
	return o.service.GetServiceState(o.ctx)
}

func (o *Verifier) IsSupportedMediaType(mt string) (bool, error) {
	normalizedMediaType, err := util.NormalizeMediaType(mt)
	if err != nil {
		return false, fault.BadInput("validation failed for %s (%w)", mt, err)
	}

	serviceState, err := o.service.GetServiceState(o.ctx)
	if err != nil {
		return false, err
	}

	for _, supportedMediaType := range serviceState.VerificationMediaTypes {
		if supportedMediaType == normalizedMediaType {
			return true, nil
		}
	}

	return false, nil
}

func (o *Verifier) SupportedMediaTypes() ([]string, error) {
	serviceState, err := o.service.GetServiceState(o.ctx)
	if err != nil {
		return nil, err
	}

	return serviceState.VerificationMediaTypes, nil
}

func (o *Verifier) ProcessEvidence(
	tenantID string,
	nonce []byte,
	data []byte,
	mt string,
) ([]byte, error) {
	ctx := vts.SetTenantID(o.ctx, tenantID)
	evidence := &scheme.Evidence{
		Data:      data,
		MediaType: mt,
		Nonce:     nonce,
	}

	return o.service.GetAttestation(ctx, evidence)
}

func (o *Verifier) PublicKey() (string, error) {
	state, err := o.service.GetServiceState(o.ctx)
	if err != nil {
		return "", err
	}

	return state.EarSigningPublicKey, nil
}
