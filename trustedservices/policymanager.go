// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package trustedservices

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"
	"github.com/veraison/services/policy"
	"github.com/veraison/services/proto"
	"go.uber.org/zap"
)

type PolicyManager struct {
	Store *policy.Store
	Agent policy.IAgent

	logger *zap.SugaredLogger
}

func NewPolicyManager(
	v *viper.Viper,
	store *policy.Store,
	logger *zap.SugaredLogger,
) (*PolicyManager, error) {
	agent, err := policy.CreateAgent(v, logger)
	if err != nil {
		return nil, err
	}

	logger.Infow("agent created", "agent", agent.GetBackendName())

	pm := &PolicyManager{Agent: agent, Store: store, logger: logger}

	return pm, nil
}

func (o *PolicyManager) Evaluate(
	ctx context.Context,
	appraisal *Appraisal,
	endorsements []string,
) error {
	evidence := appraisal.EvidenceContext
	policyID := o.getPolicyID(evidence)

	pol, err := o.getPolicy(policyID)
	if err != nil {
		if errors.Is(err, policy.ErrNoPolicy) {
			o.logger.Debugw("no policy", "policy-id", policyID)
			return nil // No policy? No problem!
		}

		return err
	}

	appraisal.Result, err = o.Agent.Evaluate(ctx, pol, appraisal.Result, evidence, endorsements)
	if err != nil {
		return err
	}

	return nil
}

func (o *PolicyManager) getPolicyID(ec *proto.EvidenceContext) string {
	return fmt.Sprintf("%s://%s",
		o.Agent.GetBackendName(),
		ec.TenantId,
	)

}

func (o *PolicyManager) getPolicy(policyID string) (*policy.Policy, error) {
	p, err := o.Store.GetLatest(policyID)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
