// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"github.com/veraison/ear"
	"github.com/veraison/services/config"
	"github.com/veraison/services/proto"
	"go.uber.org/zap"
)

var ErrBadResult = "could not create updated AttestationResult: %w from JSON %s"
var ErrNoStatus = "backend returned outcome with no status field: %v"
var ErrNoTV = "backend returned no trust-vector field, or its not a map[string]interface{}: %v"

type cfg struct {
	Backend string
}

func (o cfg) Validate() error {
	if _, ok := backends[o.Backend]; !ok {
		return fmt.Errorf("backend %q is not supported", o.Backend)
	}

	return nil
}

// CreateAgent creates a new PolicyAgent using the backend specified in the
// config with "policy.backend" directive. If this directive is absent, the
// default backend, "opa",  will be used.
func CreateAgent(v *viper.Viper, logger *zap.SugaredLogger) (IAgent, error) {
	cfg := cfg{Backend: DefaultBackend}

	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return nil, err
	}

	return &Agent{Backend: backends[cfg.Backend], logger: logger}, nil
}

type Agent struct {
	Backend IBackend

	logger *zap.SugaredLogger
}

func (o *Agent) Init(v *viper.Viper) error {
	if err := o.Backend.Init(v); err != nil {
		return err
	}

	return nil
}

// GetBackendName returns a string containing the name of the backend used by
// the agent.
func (o *Agent) GetBackendName() string {
	return o.Backend.GetName()
}

// Evaluate the provided policy w.r.t. to the specified evidence and
// endorsements, and return an updated AttestationResult. The policy may
// overwrite the result status or any of the values in the result trust vector.
func (o *Agent) Evaluate(
	ctx context.Context,
	sessionContext map[string]interface{},
	scheme string,
	policy *Policy,
	submod string,
	appraisal *ear.Appraisal,
	evidence *proto.EvidenceContext,
	endorsements []string,
) (*ear.Appraisal, error) {

	resultMap := appraisal.AsMap()
	appraisalUpdated := false

	updatedByPolicy, err := o.Backend.Evaluate(
		ctx,
		sessionContext,
		scheme,
		policy.Rules,
		resultMap,
		evidence.Evidence.AsMap(),
		endorsements,
	)
	if err != nil {
		return nil, fmt.Errorf("could not evaluate policy: %w", err)
	}

	o.logger.Debugw("policy evaluated", "policy-id", policy.StoreKey, "updated", updatedByPolicy)

	updatedStatus, ok := updatedByPolicy["ear.status"]
	if !ok {
		return nil, fmt.Errorf(ErrNoStatus, updatedByPolicy)
	}

	if updatedStatus != "" {
		appraisalUpdated = true
		resultMap["ear.status"] = updatedByPolicy["ear.status"]
	}

	updatedTV, ok := updatedByPolicy["ear.trustworthiness-vector"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(ErrNoTV, updatedByPolicy)
	}

	for k, v := range updatedTV {
		if v != "" && v != ear.NoClaim {
			appraisalUpdated = true
			resultMap["ear.trustworthiness-vector"].(map[string]interface{})[k] = v
		}
	}

	updatedAddedClaims, ok := updatedByPolicy["ear.veraison.policy-claims"].(*map[string]interface{})
	if ok {
		appraisalUpdated = true
		resultMap["ear.veraison.policy-claims"] = *updatedAddedClaims
	}

	if appraisalUpdated {
		evaluatedAppraisal, err := ear.ToAppraisal(resultMap)
		if err != nil {
			return nil, fmt.Errorf("bad appraisal data from policy: %w", err)
		}
		evaluatedAppraisal.AppraisalPolicyID = appraisal.AppraisalPolicyID

		return evaluatedAppraisal, nil
	} else {
		// policy did not update anything, so return the original appraisal
		return appraisal, nil
	}
}

// Validate performs basic validation of the provided policy rules, returning
// an error if it fails. the nature of the validation performed is
// backend-specific, however it would typically amount to a syntax check.
// Successful validation does not guarantee that the policy will execute
// correctly againt actual inputs.
func (o *Agent) Validate(ctx context.Context, policyRules string) error {
	return o.Backend.Validate(ctx, policyRules)
}

func (o *Agent) GetBackend() IBackend {
	return o.Backend
}

func (o *Agent) Close() {
	o.Backend.Close()
}
