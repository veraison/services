// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/viper"
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
	policy *Policy,
	result *proto.AttestationResult,
	evidence *proto.EvidenceContext,
	endorsements []string,
) (*proto.AttestationResult, error) {
	resultBytes, err := result.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("could not marshal provided result: %w", err)
	}

	var resultMap map[string]interface{}
	if err = json.Unmarshal(resultBytes, &resultMap); err != nil {
		return nil, fmt.Errorf("could not unmarshal provided result: %w", err)
	}

	updatedByPolicy, err := o.Backend.Evaluate(
		ctx,
		policy.Rules,
		resultMap,
		evidence.Evidence.AsMap(),
		endorsements,
	)
	if err != nil {
		return nil, fmt.Errorf("could not evaluate policy: %w", err)
	}

	o.logger.Debugw("policy evaluated", "policy-id", policy.ID, "updated", updatedByPolicy)

	updatedStatus, ok := updatedByPolicy["status"]
	if !ok {
		return nil, fmt.Errorf(ErrNoStatus, updatedByPolicy)
	}

	if updatedStatus != "" {
		resultMap["status"] = updatedByPolicy["status"]
	}

	updatedTV, ok := updatedByPolicy["trust-vector"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(ErrNoTV, updatedByPolicy)
	}

	for k, v := range updatedTV {
		if v != "" {
			resultMap["trust-vector"].(map[string]interface{})[k] = v
		}
	}

	evalBytes, err := json.Marshal(resultMap)
	if err != nil {
		return nil, fmt.Errorf("could not marshal updated result: %w", err)
	}

	var evaluatedResult proto.AttestationResult

	if err = evaluatedResult.UnmarshalJSON(evalBytes); err != nil {
		return nil, fmt.Errorf(ErrBadResult, err, evalBytes)
	}

	evaluatedResult.AppraisalPolicyID = policy.ID

	return &evaluatedResult, nil
}

func (o *Agent) GetBackend() IBackend {
	return o.Backend
}

func (o *Agent) Close() {
	o.Backend.Close()
}
