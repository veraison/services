package policy

import (
	"context"

	"github.com/veraison/services/config"
)

type IBackend interface {
	Init(cfg config.Store) error
	GetName() string
	Evaluate(
		ctx context.Context,
		policy string,
		result map[string]interface{},
		evidence map[string]interface{},
		endorsements []string,
	) (map[string]interface{}, error)
	Close()
}
