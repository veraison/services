// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policymanager

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"
	"github.com/veraison/services/policy"
	"github.com/veraison/services/proto"
)

type PolicyManager struct {
	Store *policy.Store
	Agent policy.IAgent
}

func New(v *viper.Viper, store *policy.Store) (*PolicyManager, error) {
	agent, err := policy.CreateAgent(v)
	if err != nil {
		return nil, err
	}

	pm := &PolicyManager{Agent: agent, Store: store}

	return pm, nil
}

func (o *PolicyManager) Evaluate(
	ctx context.Context,
	ac *proto.AppraisalContext,
	endorsements []string,
) error {
	evidence := ac.Evidence

	pol, err := o.getPolicy(evidence)
	if err != nil {
		if errors.Is(err, policy.ErrNoPolicy) {
			return nil // No policy? No problem!
		}

		return err
	}

	updatedResult, err := o.Agent.Evaluate(ctx, pol, ac.Result, evidence, endorsements)
	if err != nil {
		return err
	}

	ac.Result = updatedResult

	return nil
}

func (o *PolicyManager) getPolicy(ev *proto.EvidenceContext) (*policy.Policy, error) {
	policyID := fmt.Sprintf("%s://%s/%s",
		o.Agent.GetBackendName(),
		ev.TenantId,
		ev.Format.String(),
	)

	p, err := o.Store.GetLatest(policyID)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
