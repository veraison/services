// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"context"

	"github.com/spf13/viper"
	"github.com/veraison/corim/comid"
	"github.com/veraison/ear"
)

//go:generate mockgen -source iagent.go -destination mock_agent_test.go -package policy
type IAgent interface {
	Init(v *viper.Viper) error
	GetBackendName() string
	Evaluate(ctx context.Context,
		sessionContext map[string]any,
		scheme string,
		claims map[string]any,
		policy *Policy,
		submod string,
		appraisal *ear.Appraisal,
		endorsements []*comid.ValueTriple,
	) (*ear.Appraisal, error)
	Validate(ctx context.Context, policyRules string) error
	Close()
}
