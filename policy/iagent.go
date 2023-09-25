// Copyright 2022-2023 Contributors to the Veraison project.
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
		appraisalContext map[string]interface{},
		scheme string,
		policy *Policy,
		submod string,
		appraisal *ear.Appraisal,
		evidence *proto.EvidenceContext,
		endorsements []string,
	) (*ear.Appraisal, error)
	Validate(ctx context.Context, policyRules string) error
	Close()
}
