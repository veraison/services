// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"context"

	"github.com/spf13/viper"
)

//go:generate mockgen -source ibackend.go -destination mock_backend_test.go -package policy
type IBackend interface {
	Init(v *viper.Viper) error
	GetName() string
	Evaluate(
		ctx context.Context,
		sessionContext map[string]any,
		scheme string,
		policy string,
		result map[string]any,
		evidence map[string]any,
		endorsements []map[string]any,
	) (map[string]any, error)
	Validate(ctx context.Context, policy string) error
	Close()
}
