// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policymanager

import (
	"context"
	"errors"

	"github.com/spf13/viper"
	"github.com/veraison/services/policy"
	"github.com/veraison/services/vts/appraisal"
	"go.uber.org/zap"
)

type PolicyManager struct {
	Store *policy.Store
	Agent policy.IAgent

	logger *zap.SugaredLogger
}

func New(v *viper.Viper, store *policy.Store, logger *zap.SugaredLogger) (*PolicyManager, error) {
	agent, err := policy.CreateAgent(v, logger)
	if err != nil {
		return nil, err
	}

	logger.Infow("agent created", "agent", agent.GetBackendName())

	pm := &PolicyManager{Agent: agent, Store: store, logger: logger}

	return pm, nil
}

// XXX(tho) revisit coupling between appraisal and policy manager
func (o *PolicyManager) Evaluate(
	ctx context.Context,
	appraisal *appraisal.Appraisal,
) error {
	policyKey := o.getPolicyKey(appraisal.EvidenceContext.TenantId, appraisal.Scheme)

	pol, err := o.getPolicy(policyKey)
	if err != nil {
		if errors.Is(err, policy.ErrNoPolicy) {
			o.logger.Debugw("no policy", "policy-id", policyKey)
			return nil // No policy? No problem!
		}

		return err
	}

	appraisalContext := map[string]interface{}{
		"nonce": appraisal.Result.Nonce,
	}

	for submod, submodAppraisal := range appraisal.Result.Submods {
		evaluated, err := o.Agent.Evaluate(
			ctx,
			appraisalContext,
			appraisal.Scheme,
			pol,
			submod,
			submodAppraisal,
			appraisal.EvidenceContext,
			appraisal.Endorsements,
		)
		if err != nil {
			return err
		}
		appraisal.Result.Submods[submod] = evaluated
	}
	if err := appraisal.UpdatePolicyID(pol); err != nil {
		return err
	}

	return nil
}

func (o *PolicyManager) getPolicyKey(tenantID string, scheme string) policy.PolicyKey {
	return policy.PolicyKey{
		TenantId: tenantID,
		Scheme:   scheme,
		Name:     o.Agent.GetBackendName(),
	}
}

func (o *PolicyManager) getPolicy(policyKey policy.PolicyKey) (*policy.Policy, error) {
	p, err := o.Store.GetActive(policyKey)
	if err != nil {
		return nil, err
	}

	return p, nil
}
