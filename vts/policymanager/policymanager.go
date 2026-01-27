// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policymanager

import (
	"context"
	"errors"

	"github.com/spf13/viper"
	"github.com/veraison/corim/comid"
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

func (o *PolicyManager) Evaluate(
	ctx context.Context,
	appraisalContext *appraisal.Context,
	endorsements []*comid.ValueTriple,
) error {
	policyKey := o.getPolicyKey(appraisalContext)

	pol, err := o.getPolicy(policyKey)
	if err != nil {
		if errors.Is(err, policy.ErrNoPolicy) {
			o.logger.Debugw("no policy", "policy-id", policyKey)
			return nil // No policy? No problem!
		}

		return err
	}

	sessionContext := map[string]any{
		"nonce": appraisalContext.Result.Nonce,
	}

	for submodName, submodAppraisal := range appraisalContext.Result.Submods {
		evaluated, err := o.Agent.Evaluate(
			ctx,
			sessionContext,
			appraisalContext,
			pol,
			submodName,
			submodAppraisal,
			endorsements,
		)
		if err != nil {
			return err
		}
		appraisalContext.Result.Submods[submodName] = evaluated
	}

	if err := appraisalContext.UpdatePolicyID(pol.UUID.String()); err != nil {
		return err
	}

	return nil
}

func (o *PolicyManager) getPolicyKey(a *appraisal.Context) policy.PolicyKey {
	return policy.PolicyKey{
		TenantId: a.Evidence.TenantID,
		Scheme:   a.Scheme,
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
