// Copyright 2022-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"context"
	"errors"
	"fmt"

	"github.com/veraison/services/api"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/vtsclient"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ErrInputParam = errors.New("invalid input parameter")

type Provisioner struct {
	VTSClient vtsclient.IVTSClient
}

func New(vtsClient vtsclient.IVTSClient) IProvisioner {
	return &Provisioner{
		VTSClient: vtsClient,
	}
}

func (p *Provisioner) IsSupportedMediaType(mt string) (bool, error) {
	normalizedMediaType, err := api.NormalizeMediaType(mt)
	if err != nil {
		return false, fmt.Errorf("%w: validation failed for %s (%v)", ErrInputParam, mt, err)
	}

	mts, err := p.VTSClient.GetSupportedProvisioningMediaTypes(
		context.Background(),
		&emptypb.Empty{},
	)
	if err != nil {
		return false, err
	}

	for _, v := range mts.MediaTypes {
		if v == normalizedMediaType {
			return true, nil
		}
	}

	return false, nil
}

func (p *Provisioner) SupportedMediaTypes() ([]string, error) {
	mts, err := p.VTSClient.GetSupportedProvisioningMediaTypes(
		context.Background(),
		&emptypb.Empty{},
	)
	if err != nil {
		return nil, err
	}

	return mts.GetMediaTypes(), nil
}

func (p *Provisioner) SubmitEndorsements(tenantID string, data []byte, mt string) error {
	// return p.VTSClient.SubmitEndorsements(context.Background(),)
	sReq := &proto.SubmitEndorsementsRequest{MediaType: mt, Data: data}
	sRes, err := p.VTSClient.SubmitEndorsements(context.Background(), sReq)
	if err != nil {
		if errors.As(err, &vtsclient.NoConnectionError{}) {
			return errors.New("no connection")
		}
		return fmt.Errorf("submit endorsements failed: %w", err)
	}

	if !sRes.GetStatus().Result {
		return fmt.Errorf(
			"submit endorsements failed: %s",
			sRes.Status.GetErrorDetail(),
		)
	}
	return nil
}

func (p *Provisioner) GetVTSState() (*proto.ServiceState, error) {
	return p.VTSClient.GetServiceState(context.TODO(), &emptypb.Empty{})
}

func (p *Provisioner) GetEndorsements(keyPrefix string, endorsementType string) (*proto.GetEndorsementsResponse, error) {
	req := &proto.GetEndorsementsRequest{
		KeyPrefix:       keyPrefix,
		EndorsementType: endorsementType,
	}
	return p.VTSClient.GetEndorsements(context.Background(), req)
}

func (p *Provisioner) DeleteEndorsements(key string, endorsementType string) (*proto.DeleteEndorsementsResponse, error) {
	req := &proto.DeleteEndorsementsRequest{
		Key:             key,
		EndorsementType: endorsementType,
	}
	return p.VTSClient.DeleteEndorsements(context.Background(), req)
}
