// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package management

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/veraison/services/builtin"
	"github.com/veraison/services/config"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/policy"
)

type PolicyManager struct {
	Agent            policy.IAgent
	Store            *policy.Store
	SupportedSchemes []string
}

func CreatePolicyManagerFromConfig(v *viper.Viper, name string) (*PolicyManager, error) {
	subs, err := config.GetSubs(v, "*po-agent", "po-store", "*plugin")
	if err != nil {
		return nil, err
	}

	agent, err := policy.CreateAgent(subs["po-agent"], log.Named(name+"-agent"))
	if err != nil {
		return nil, err
	}

	store, err := policy.NewStore(subs["po-store"], log.Named(name+"-store"))
	if err != nil {
		return nil, err
	}

	var pluginManager plugin.IManager[handler.ISchemeHandler]

	if config.SchemeLoader == "plugins" { // nolint:gocritic
		pluginManager, err = plugin.CreateGoPluginManager(
			subs["plugin"], log.Named("plugin"),
			"scheme-handler", handler.SchemeHandlerRPC)
		if err != nil {
			log.Fatalf("plugin manager initialization failed: %v", err)
		}
	} else if config.SchemeLoader == "builtin" {
		pluginManager, err = builtin.CreateBuiltinManager[handler.ISchemeHandler](
			subs["plugin"], log.Named("builtin"), "scheme-handler")
		if err != nil {
			log.Fatalf("scheme manager initialization failed: %v", err)
		}
	} else {
		log.Panicw("invalid SchemeLoader value", "SchemeLoader", config.SchemeLoader)
	}
	defer pluginManager.Close()

	supportedSchemes := pluginManager.GetRegisteredAttestationSchemes()

	return NewPolicyManager(agent, store, supportedSchemes), nil
}

func NewPolicyManager(agent policy.IAgent, store *policy.Store, schemes []string) *PolicyManager {
	return &PolicyManager{Agent: agent, Store: store, SupportedSchemes: schemes}
}

func (o *PolicyManager) IsSchemeSupported(scheme string) bool {
	for _, supported := range o.SupportedSchemes {
		if supported == scheme {
			return true
		}
	}

	return false
}

func (o *PolicyManager) Validate(ctx context.Context, policyRules string) error {
	return o.Agent.Validate(ctx, policyRules)
}

func (o *PolicyManager) Update(
	ctx context.Context,
	tenantID string,
	scheme string,
	name string,
	rules string,
) (*policy.Policy, error) {
	key, err := o.resolvePolicyKey(tenantID, scheme)
	if err != nil {
		return nil, err
	}

	return o.Store.Update(key, name, o.Agent.GetBackendName(), rules)
}

func (o *PolicyManager) GetActive(
	ctx context.Context,
	tenantID string,
	scheme string,
) (*policy.Policy, error) {
	key, err := o.resolvePolicyKey(tenantID, scheme)
	if err != nil {
		return nil, err
	}

	return o.Store.GetActive(key)
}

func (o *PolicyManager) GetPolicy(
	ctx context.Context,
	tenantID string,
	scheme string,
	policyID uuid.UUID,
) (*policy.Policy, error) {
	key, err := o.resolvePolicyKey(tenantID, scheme)
	if err != nil {
		return nil, err
	}

	return o.Store.GetPolicy(key, policyID)
}

func (o *PolicyManager) GetPolicies(
	ctx context.Context,
	tenantID string,
	scheme string,
	name string,
) ([]*policy.Policy, error) {
	key, err := o.resolvePolicyKey(tenantID, scheme)
	if err != nil {
		return nil, err
	}

	policies, err := o.Store.Get(key)
	if err != nil {
		return nil, err
	}

	if name == "" {
		return policies, nil
	}

	ret := make([]*policy.Policy, 0)

	for _, pol := range policies {
		if pol.Name == name {
			ret = append(ret, pol)
		}
	}

	return ret, nil
}

func (o *PolicyManager) Activate(
	ctx context.Context,
	tenantID string,
	scheme string,
	policyID uuid.UUID,
) error {
	key, err := o.resolvePolicyKey(tenantID, scheme)
	if err != nil {
		return err
	}

	return o.Store.Activate(key, policyID)
}

func (o *PolicyManager) DeactivateAll(
	ctx context.Context,
	tenantID string,
	scheme string,
) error {
	key, err := o.resolvePolicyKey(tenantID, scheme)
	if err != nil {
		return err
	}

	return o.Store.DeactivateAll(key)
}

func (o *PolicyManager) resolvePolicyKey(
	tenantID string,
	scheme string,
) (policy.PolicyKey, error) {
	schemeFound := false
	for _, supportedScheme := range o.SupportedSchemes {
		if supportedScheme == scheme {
			schemeFound = true
			break
		}
	}

	if !schemeFound {
		return policy.PolicyKey{}, fmt.Errorf("Unsupported attestation scheme: %q", scheme)
	}

	return policy.PolicyKey{
		TenantId: tenantID,
		Scheme:   scheme,
		Name:     o.Agent.GetBackendName(),
	}, nil
}
