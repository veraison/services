// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"context"

	"github.com/spf13/viper"
	"github.com/veraison/ear"
	"github.com/veraison/services/proto"
)

type IAgent interface {
	Init(v *viper.Viper) error
	GetBackendName() string
	Evaluate(ctx context.Context,
		policy *Policy,
		result *ear.AttestationResult,
		evidence *proto.EvidenceContext,
		endorsements []string) (*ear.AttestationResult, error)
	Close()
}
