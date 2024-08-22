// Copyright 2022-2023 Contributors to the Veraison project.
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
	mt, err := api.NormalizeMediaType(mt)
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
		if v == mt {
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
