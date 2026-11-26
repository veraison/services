// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"github.com/veraison/corim/comid"
	"github.com/veraison/ear"
	"github.com/veraison/services/pkg/config"
	"go.uber.org/zap"
)

type Manager struct {
	Store *Store
	Agent IAgent

	logger *zap.SugaredLogger
}

func CreateManager(v *viper.Viper, logger *zap.SugaredLogger) (*Manager, error) {
	logger.Debug("loading policy configuration")
	for _, key := range v.AllKeys() {
		top := strings.SplitN(key, ".", 2)[0]
		if top != "store" && top != "agent" {
			return nil, fmt.Errorf("unexpected entry in policy settings: %q", top)
		}
	}

	subs, err := config.GetSubs(v, "store", "agent")
	if err != nil {
		return nil, err
	}

	logger.Debug("initializing policy store")
	store, err := NewStore(subs["store"], logger.Named("store"))
	if err != nil {
		return nil, fmt.Errorf("could not initialize policy store: %w", err)
	}

	logger.Debug("creating policy manager")
	manager, err := NewManager(subs["agent"], store, logger)
	if err != nil {
		return nil, fmt.Errorf("could not create policy manager: %w", err)
	}

	return manager, nil
}

func NewManager(v *viper.Viper, store *Store, logger *zap.SugaredLogger) (*Manager, error) {
	agent, err := CreateAgent(v, logger.Named("agent"))
	if err != nil {
		return nil, err
	}

	logger.Infow("agent created", "agent", agent.GetBackendName())

	pm := &Manager{Agent: agent, Store: store, logger: logger}

	return pm, nil
}

func (o *Manager) Evaluate(
	ctx context.Context,
	tenantID string,
	scheme string,
	claims map[string]any,
	result *ear.AttestationResult,
	endorsements []*comid.ValueTriple,
) error {
	policyKey := o.getPolicyKey(tenantID, scheme)

	pol, err := o.getPolicy(policyKey)
	if err != nil {
		if errors.Is(err, ErrNoPolicy) || errors.Is(err, ErrNoActivePolicy) {
			o.logger.Debugw("no policy", "policy-id", policyKey)
			return nil // No policy? No problem!
		}

		return err
	}

	sessionContext := map[string]any{
		"nonce": result.Nonce,
	}

	for submodName, submodAppraisal := range result.Submods {
		evaluated, err := o.Agent.Evaluate(
			ctx,
			sessionContext,
			scheme,
			claims,
			pol,
			submodName,
			submodAppraisal,
			endorsements,
		)
		if err != nil {
			return err
		}
		result.Submods[submodName] = evaluated
	}

	return updatePolicyID(result, pol.UUID.String())
}

func (o *Manager) Close() {
	o.Agent.Close()
	if err := o.Store.Close(); err != nil {
		o.logger.Errorf("closing store: %s", err.Error())
	}
}

func (o *Manager) getPolicyKey(tenantID, scheme string) PolicyKey {
	return PolicyKey{
		TenantId: tenantID,
		Scheme:   scheme,
		Name:     o.Agent.GetBackendName(),
	}
}

func (o *Manager) getPolicy(policyKey PolicyKey) (*Policy, error) {
	p, err := o.Store.GetActive(policyKey)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func updatePolicyID(result *ear.AttestationResult, polID string) error {
	for _, submod := range result.Submods {
		updatedID := *submod.AppraisalPolicyID + "/" + polID
		submod.AppraisalPolicyID = &updatedID
	}

	return nil
}
