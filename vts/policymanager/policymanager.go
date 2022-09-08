// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policymanager

import (
	"context"
	"errors"
	"fmt"

	"github.com/veraison/services/config"
	"github.com/veraison/services/kvstore"
	"github.com/veraison/services/policy"
	"github.com/veraison/services/proto"
)

var ErrNoPolicy = errors.New("no policy found")

type PolicyManager struct {
	Store kvstore.IKVStore
	Agent *policy.PolicyAgent
}

func New(cfg config.Store, store kvstore.IKVStore) (*PolicyManager, error) {
	agent, err := policy.CreateAgent(cfg)
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

	policy, err := o.getPolicy(evidence)
	if err != nil {
		if err == ErrNoPolicy {
			return nil // No policy? No problem!
		}

		return err
	}

	updatedResult, err := o.Agent.Evaluate(ctx, policy, ac.Result, evidence, endorsements)
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

	vals, err := o.Store.Get(policyID)
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) {
			return nil, fmt.Errorf("%w: %q", ErrNoPolicy, policyID)
		}
		return nil, err
	}

	// TODO(setrofim): for now, assuming that there should be exactly one
	// matching policy. Once we have a more sophisticated policy management
	// framework worked out, we might allow multiple policies here.
	if len(vals) == 0 {
		return nil, fmt.Errorf("%w: %q", ErrNoPolicy, policyID)
	} else if len(vals) > 1 {
		return nil, fmt.Errorf("found %d policy entries for id %q; must be at most 1",
			len(vals), policyID)
	}

	return &policy.Policy{ID: policyID, Rules: vals[0]}, nil
}
