package policy

import (
	"context"

	"github.com/setrofim/viper"
	"github.com/veraison/services/proto"
)

type IAgent interface {
	Init(v *viper.Viper) error
	GetBackendName() string
	Evaluate(ctx context.Context,
		policy *Policy,
		result *proto.AttestationResult,
		evidence *proto.EvidenceContext,
		endorsements []string) (*proto.AttestationResult, error)
	Close()
}
