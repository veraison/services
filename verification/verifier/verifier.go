package verifier

import (
	"context"
	"fmt"

	"github.com/veraison/services/config"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/vtsclient"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Verifier struct {
	Config    config.Store
	VTSClient vtsclient.IVTSClient
}

func New(cfg config.Store, vtsClient vtsclient.IVTSClient) IVerifier {
	return &Verifier{
		Config:    cfg,
		VTSClient: vtsClient,
	}
}

func (o *Verifier) IsSupportedMediaType(mt string) bool {
	mts, err := o.VTSClient.GetSupportedVerificationMediaTypes(
		context.Background(),
		&emptypb.Empty{},
	)
	if err != nil {
		return false
	}

	for _, v := range mts.MediaTypes {
		if v == mt {
			return true
		}
	}

	return false
}

func (o *Verifier) SupportedMediaTypes() []string {
	mts, err := o.VTSClient.GetSupportedVerificationMediaTypes(
		context.Background(),
		&emptypb.Empty{},
	)
	if err != nil {
		return nil
	}

	return mts.GetMediaTypes()
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
	res, err := appraisalCtx.Result.MarshalJSON()
	if err != nil {
		return nil, err
	}
	fmt.Printf("Apprisal Context %s,", string(res))
	return appraisalCtx.Result.MarshalJSON()

}
