// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package management

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/veraison/services/pkg/config"
	"github.com/veraison/services/pkg/log"
	"github.com/veraison/services/pkg/plugin"
	"github.com/veraison/services/pkg/plugin/loader"
	"github.com/veraison/services/pkg/policy"
	"github.com/veraison/services/pkg/scheme"
	"go.uber.org/zap"
)

type PolicyManager struct {
	Agent            policy.IAgent
	Store            *policy.Store
	SupportedSchemes []string
	logger           *zap.SugaredLogger
}

func CreatePolicyManagerFromConfig(v *viper.Viper, name string) (*PolicyManager, error) {
	subs, err := config.GetSubs(v, "policy", "*plugin", "*scheme")
	if err != nil {
		return nil, err
	}

	policySubs, err := config.GetSubs(subs["policy"], "agent", "store")
	if err != nil {
		return nil, err
	}

	logger := log.Named("policy")

	agent, err := policy.CreateAgent(policySubs["agent"], logger.Named("agent"))
	if err != nil {
		return nil, err
	}

	store, err := policy.NewStore(policySubs["store"], logger.Named("store"))
	if err != nil {
		return nil, err
	}

	pluginLoader, err := initPluginLoader(subs["plugin"])
	if err != nil {
		return nil, err
	}
	defer pluginLoader.Close()

	schemeManager, err := initAttestationSchemes(subs["scheme"], pluginLoader)
	if err != nil {
		return nil, err
	}
	defer schemeManager.Close()

	supportedSchemes := schemeManager.GetRegisteredAttestationSchemes()

	return NewPolicyManager(agent, store, supportedSchemes, logger), nil
}

func NewPolicyManager(
	agent policy.IAgent,
	store *policy.Store,
	schemes []string,
	logger *zap.SugaredLogger,
) *PolicyManager {
	return &PolicyManager{Agent: agent, Store: store, SupportedSchemes: schemes, logger: logger}
}

func (o *PolicyManager) IsSchemeSupported(schemeName string) bool {
	for _, supported := range o.SupportedSchemes {
		if supported == schemeName {
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
	schemeName string,
	name string,
	rules string,
) (*policy.Policy, error) {
	key, err := o.resolvePolicyKey(tenantID, schemeName)
	if err != nil {
		return nil, err
	}

	return o.Store.Update(key, name, o.Agent.GetBackendName(), rules)
}

func (o *PolicyManager) GetActive(
	ctx context.Context,
	tenantID string,
	schemeName string,
) (*policy.Policy, error) {
	key, err := o.resolvePolicyKey(tenantID, schemeName)
	if err != nil {
		return nil, err
	}

	return o.Store.GetActive(key)
}

func (o *PolicyManager) GetPolicy(
	ctx context.Context,
	tenantID string,
	schemeName string,
	policyID uuid.UUID,
) (*policy.Policy, error) {
	key, err := o.resolvePolicyKey(tenantID, schemeName)
	if err != nil {
		return nil, err
	}

	return o.Store.GetPolicy(key, policyID)
}

func (o *PolicyManager) GetPolicies(
	ctx context.Context,
	tenantID string,
	schemeName string,
	name string,
) ([]*policy.Policy, error) {
	key, err := o.resolvePolicyKey(tenantID, schemeName)
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
	schemeName string,
	policyID uuid.UUID,
) error {
	key, err := o.resolvePolicyKey(tenantID, schemeName)
	if err != nil {
		return err
	}

	return o.Store.Activate(key, policyID)
}

func (o *PolicyManager) DeactivateAll(
	ctx context.Context,
	tenantID string,
	schemeName string,
) error {
	key, err := o.resolvePolicyKey(tenantID, schemeName)
	if err != nil {
		return err
	}

	return o.Store.DeactivateAll(key)
}

func (o *PolicyManager) resolvePolicyKey(
	tenantID string,
	schemeName string,
) (policy.PolicyKey, error) {
	schemeFound := false
	for _, supportedScheme := range o.SupportedSchemes {
		if supportedScheme == schemeName {
			schemeFound = true
			break
		}
	}

	if !schemeFound {
		return policy.PolicyKey{}, fmt.Errorf("unsupported attestation scheme: %q", schemeName)
	}

	return policy.PolicyKey{
		TenantId: tenantID,
		Scheme:   schemeName,
		Name:     o.Agent.GetBackendName(),
	}, nil
}

func (o *PolicyManager) Close() {
	o.logger.Debug("closing agent")
	o.Agent.Close()

	o.logger.Debug("closing store")
	if err := o.Store.Close(); err != nil {
		o.logger.Errorf("closing store: %v", err)
	}
}

func initPluginLoader(v *viper.Viper) (plugin.ILoader, error) {
	log.Info("discovering plugins")
	type pluginConfig struct {
		Locations []string `mapstructure:"dirs"`
	}
	cfg := pluginConfig{Locations: []string{"."}}

	configLoader := config.NewLoader(&cfg)
	if err := configLoader.LoadFromViper(v); err != nil {
		return nil, fmt.Errorf("could not load plugin config: %w", err)
	}

	pluginLoader := loader.NewPluginLoader(log.Named("plugin"))
	for _, location := range cfg.Locations {
		if err := pluginLoader.LoadFrom(location); err != nil {
			return nil, fmt.Errorf("could not load plugins from %s: %w", location, err)
		}
	}

	return pluginLoader, nil
}

func initAttestationSchemes(v *viper.Viper, pluginLoader plugin.ILoader) (*scheme.Manager, error) {
	log.Info("processing attestation schemes")

	log.Debug("loading scheme configuration")
	params, err := plugin.ParametersMapFromViper(v, scheme.PluginNameFromScheme)
	if err != nil {
		return nil, fmt.Errorf("could not load scheme configuration: %w", err)
	}

	log.Debug("creating scheme manager")
	manager, err := scheme.CreateManager(pluginLoader, params, log.Named("scheme"))
	if err != nil {
		return nil, fmt.Errorf("could not create scheme manager: %w", err)
	}

	return manager, nil
}
