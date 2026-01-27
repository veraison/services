// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/open-policy-agent/opa/rego"
	"github.com/spf13/viper"
	"github.com/veraison/ear"
	"github.com/veraison/services/log"
)

var ErrBadOPAResult = errors.New("bad result update from policy")

//go:embed opa.rego
var preambleText string

type OPA struct {
}

func NewOPA(v *viper.Viper) (*OPA, error) {
	var o OPA
	if err := o.Init(v); err != nil {
		return nil, err
	}
	return &o, nil
}

func (o *OPA) Init(v *viper.Viper) error {
	return nil
}

func (o *OPA) GetName() string {
	return "opa"
}

func (o *OPA) Evaluate(
	ctx context.Context,
	sessionContext map[string]any,
	scheme string,
	policy string,
	result map[string]any,
	evidence map[string]any,
	endorsements []map[string]any,
) (map[string]interface{}, error) {

	input := map[string]any{
		"scheme":       scheme,
		"session":      sessionContext,
		"result":       result,
		"evidence":     evidence,
		"endorsements": endorsements,
	}

	rego := rego.New(
		rego.Package("policy"),
		rego.Module("opa.rego", preambleText),
		rego.Module("policy.rego", policy),
		rego.Input(input),
		rego.Query("outcome"),
		rego.Dump(log.NamedWriter("opa", log.DebugLevel)),
	)

	resultSet, err := rego.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not Eval policy: %w", err)
	}

	value := resultSet[0].Expressions[0].Value

	resultUpdate, err := processUpdateValue(value)

	if err != nil {
		return nil, fmt.Errorf("policy returned bad update: %w", err)
	}

	return resultUpdate, nil
}

func (o *OPA) Validate(ctx context.Context, policy string) error {
	rego := rego.New(
		rego.Package("policy"),
		rego.Module("opa.rego", preambleText),
		rego.Module("policy.rego", policy),
		rego.Query("outcome"),
		rego.Dump(log.NamedWriter("opa", log.DebugLevel)),
	)

	_, err := rego.Compile(ctx)
	return err
}

func (o *OPA) Close() {
}

func processUpdateValue(value interface{}) (map[string]interface{}, error) {
	rawUpdate, ok := value.(map[string]interface{})
	if !ok {
		err := fmt.Errorf(
			"%w: expected map[string]interface{}, but got %T",
			ErrBadOPAResult, value)
		return nil, err
	}

	updateTv := map[string]interface{}{
		"instance-identity": 0,
		"configuration":     0,
		"executables":       0,
		"file-system":       0,
		"hardware":          0,
		"runtime-opaque":    0,
		"storage-opaque":    0,
		"sourced-data":      0,
	}

	updatedStatus, err := ear.ToTrustTier(rawUpdate["status"])
	if err != nil {
		return nil, err
	}

	rawTv, ok := rawUpdate["trust-vector"].(map[string]interface{})
	if !ok {
		err := fmt.Errorf(
			"%w: \"trust-vector\" value should be map[string]interface{}, but got %T",
			ErrBadOPAResult, value)
		return nil, err
	}

	for claim, rawValue := range rawTv {
		if _, ok := updateTv[claim]; !ok {
			err := fmt.Errorf("%w: unexpected claim %q ", ErrBadOPAResult, claim)
			return nil, err
		}

		value, err := ear.ToTrustClaim(rawValue)
		if err != nil {
			err := fmt.Errorf("%w: bad value %q for %q: %v",
				ErrBadOPAResult, rawValue, claim, err)
			return nil, err
		}

		updateTv[claim] = *value
	}

	addedClaims, ok := rawUpdate["added-claims"].(map[string]interface{})
	if !ok {
		err := fmt.Errorf(
			`%w: "added-claims" value should be map[string]interface{}, but got %T`,
			ErrBadOPAResult, value)
		return nil, err
	}

	update := map[string]interface{}{
		"ear.status":                 updatedStatus,
		"ear.trustworthiness-vector": updateTv,
		"ear.veraison.policy-claims": &addedClaims,
	}

	return update, nil
}
