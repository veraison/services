package policy

import (
	"context"

	"github.com/veraison/services/config"
	"github.com/veraison/services/proto"
)

type IAgent interface {
	GetBackend() IBackend
	Init(cfg config.Store) error
	GetBackendName() string
	Evaluate(ctx context.Context,
		policy *Policy,
		result *proto.AttestationResult,
		evidence *proto.EvidenceContext,
		endorsements []string) (*proto.AttestationResult, error)
	Close()
}
