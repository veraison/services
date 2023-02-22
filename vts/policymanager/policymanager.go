// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policymanager

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"
	"github.com/veraison/services/policy"
	"github.com/veraison/services/proto"
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
	appraisal *appraisal.Appraisal,
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

	for submod, submodAppraisal := range appraisal.Result.Submods {
		evaluated, err := o.Agent.Evaluate(
			ctx, pol, submod, submodAppraisal, evidence, endorsements)
		if err != nil {
			return err
		}
		appraisal.Result.Submods[submod] = evaluated
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
