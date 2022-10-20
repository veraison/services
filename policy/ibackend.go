// Copyright 2022 Contributors to the Veraison project.
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
		policy string,
		result map[string]interface{},
		evidence map[string]interface{},
		endorsements []string,
	) (map[string]interface{}, error)
	Close()
}
