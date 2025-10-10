// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package endorsementdistributor

import (
	"context"
	"errors"
	"fmt"

	"github.com/veraison/services/log"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/vtsclient"
	"google.golang.org/protobuf/types/known/emptypb"
)

type EndorsementDistributor struct {
	VTSClient vtsclient.IVTSClient
}

func New(vtsClient vtsclient.IVTSClient) IEndorsementDistributor {
	return &EndorsementDistributor{
		VTSClient: vtsClient,
	}
}

func (ed *EndorsementDistributor) GetEndorsements(tenantID string, query string, mediaType string) ([]byte, error) {
	req := &proto.EndorsementQueryIn{Query: query, MediaType: mediaType}

	res, err := ed.VTSClient.GetEndorsements(context.Background(), req)
	if err != nil {
		if errors.As(err, &vtsclient.NoConnectionError{}) {
			return nil, errors.New("no connection")
		}
		return nil, fmt.Errorf("get endorsements failed: %w", err)
	}

	if !res.GetStatus().Result {
		return nil, fmt.Errorf(
			"get endorsements failed: %s",
			res.Status.GetErrorDetail(),
		)
	}

	return res.ResultSet, nil
}

func (ed *EndorsementDistributor) SupportedMediaTypes() ([]string, error) {
	res, err := ed.VTSClient.GetSupportedCoservMediaTypes(
		context.Background(),
		&emptypb.Empty{},
	)
	if err != nil {
		if errors.As(err, &vtsclient.NoConnectionError{}) {
			return nil, errors.New("no connection")
		}
		return nil, fmt.Errorf("get supported endorsement profiles failed: %w", err)
	}

	log.Debugw("GetSupportedCoservMediaTypes", "media types", res.MediaTypes)

	return res.MediaTypes, nil
}

func (ed *EndorsementDistributor) GetPublicKey() (*proto.PublicKey, error) {
	res, err := ed.VTSClient.GetCoservSigningPublicKey(
		context.Background(),
		&emptypb.Empty{},
	)
	if err != nil {
		if errors.As(err, &vtsclient.NoConnectionError{}) {
			return nil, errors.New("no connection")
		}
		return nil, fmt.Errorf("get CoSERV signing public key failed: %w", err)
	}

	log.Debugw("GetPublicKey", "key", res)

	return res, nil
}
