// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"context"

	"github.com/spf13/viper"
)

type IBackend interface {
	Init(v *viper.Viper) error
	GetName() string
	Evaluate(
		ctx context.Context,
		sessionContext map[string]interface{},
		scheme string,
		policy string,
		result map[string]interface{},
		evidence map[string]interface{},
		endorsements []string,
	) (map[string]interface{}, error)
	Validate(ctx context.Context, policy string) error
	Close()
}
