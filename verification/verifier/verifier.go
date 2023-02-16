// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package verifier

import (
	"context"

	"github.com/spf13/viper"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/trustedservices"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Verifier struct {
	VTSClient trustedservices.ITrustedServices
}

func New(v *viper.Viper, vtsClient trustedservices.ITrustedServices) IVerifier {
	return &Verifier{
		VTSClient: vtsClient,
	}
}

// TODO(setrofim): this should not be part of the verifier.
func (o *Verifier) GetVTSState() (*proto.ServiceState, error) {
	return o.VTSClient.GetServiceState(context.TODO(), &emptypb.Empty{})
}

func (o *Verifier) IsSupportedMediaType(mt string) (bool, error) {
	mts, err := o.VTSClient.GetSupportedVerificationMediaTypes(
		context.Background(),
		&emptypb.Empty{},
	)
	if err != nil {
		return false, err
	}

	for _, v := range mts.MediaTypes {
		if v == mt {
			return true, nil
		}
	}

	return false, nil
}

func (o *Verifier) SupportedMediaTypes() ([]string, error) {
	mts, err := o.VTSClient.GetSupportedVerificationMediaTypes(
		context.Background(),
		&emptypb.Empty{},
	)
	if err != nil {
		return nil, err
	}

	return mts.GetMediaTypes(), nil
}

func (o *Verifier) ProcessEvidence(tenantID string, data []byte, mt string) ([]byte, error) {
	token := &proto.AttestationToken{
		TenantId:  tenantID,
		Data:      data,
		MediaType: mt,
	}

	appraisalCtx, err := o.VTSClient.GetAttestation(
		context.Background(),
		token,
	)
	if err != nil {
		return nil, err
	}

	return appraisalCtx.Result, nil
}
