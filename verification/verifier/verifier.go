// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package verifier

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"
	"github.com/veraison/services/api"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/vtsclient"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ErrInputParam = errors.New("invalid input parameter")

type Verifier struct {
	VTSClient vtsclient.IVTSClient
}

func New(v *viper.Viper, vtsClient vtsclient.IVTSClient) IVerifier {
	return &Verifier{
		VTSClient: vtsClient,
	}
}

func (o *Verifier) GetVTSState() (*proto.ServiceState, error) {
	return o.VTSClient.GetServiceState(context.TODO(), &emptypb.Empty{})
}

func (o *Verifier) IsSupportedMediaType(mt string) (bool, error) {
	dropParams := false
	mt, err := api.NormalizeMediaType(mt, dropParams)
	if err != nil {
		return false, fmt.Errorf("%w: validation failed for %s (%v)", ErrInputParam, mt, err)
	}

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

func (o *Verifier) IsSupportedCompositeEvidenceMediaType(mt string) (bool, error) {
	dropParams := true
	mt, err := api.NormalizeMediaType(mt, dropParams)
	if err != nil {
		return false, fmt.Errorf("%w: validation failed for %s (%v)", ErrInputParam, mt, err)
	}

	mts, err := o.VTSClient.GetSupportedCompositeEvidenceMediaTypes(
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

func (o *Verifier) ProcessEvidence(
	tenantID string,
	nonce []byte,
	data []byte,
	mt string,
) ([]byte, error) {
	token := &proto.AttestationToken{
		TenantId:  tenantID,
		Data:      data,
		MediaType: mt,
		Nonce:     nonce,
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

func (o *Verifier) GetPublicKey() (*proto.PublicKey, error) {
	return o.VTSClient.GetEARSigningPublicKey(context.Background(), &emptypb.Empty{})
}

func (o *Verifier) ProcessCompositeEvidence(
	tenantID string,
	nonce []byte,
	data []byte,
	mt string,
) ([]byte, error) {
	token := &proto.AttestationToken{
		TenantId:  tenantID,
		Data:      data,
		MediaType: mt,
		Nonce:     nonce,
	}

	appraisalCtx, err := o.VTSClient.GetCompositeAttestation(
		context.Background(),
		token,
	)
	if err != nil {
		return nil, err
	}

	return appraisalCtx.Result, nil
}

func (o *Verifier) SupportedCompositeEvidenceMediaTypes() ([]string, error) {
	// TODO(tho) remove hardcoded list and uncomment gRPC call once VTS supports it
	return []string{}, nil

	// return []string{
	// 	"application/cmw+cbor",
	// 	"application/cmw+cose",
	// 	"application/cmw+json",
	// 	"application/cmw+jws",
	// }, nil

	// mts, err := o.VTSClient.GetSupportedCompositeEvidenceMediaTypes(
	// 	context.Background(),
	// 	&emptypb.Empty{},
	// )
	// if err != nil {
	// 	return nil, err
	// }

	// return mts.GetMediaTypes(), nil
}
