// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/open-policy-agent/opa/rego"
	"github.com/setrofim/viper"
	"github.com/veraison/services/proto"
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
	policy string,
	result map[string]interface{},
	evidence map[string]interface{},
	endorsements []string,
) (map[string]interface{}, error) {

	input, err := constructInput(result, evidence, endorsements)
	if err != nil {
		return nil, fmt.Errorf("could not construct policy input: %w", err)
	}

	rego := rego.New(
		rego.Package("policy"),
		rego.Module("opa.rego", preambleText),
		rego.Module("policy.rego", string(policy)),
		rego.Input(input),
		rego.Query("outcome"),
		rego.Dump(log.Writer()),
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

func (o *OPA) Close() {
}

func constructInput(
	result map[string]interface{},
	evidence map[string]interface{},
	endorsementStrings []string,
) (map[string]interface{}, error) {
	var endorsements []map[string]interface{}

	for i, es := range endorsementStrings {
		var e map[string]interface{}

		if err := json.Unmarshal([]byte(es), &e); err != nil {
			return nil, fmt.Errorf("endorsement %d is not valid JSON: %w", i, err)
		}

		endorsements = append(endorsements, e)
	}

	return map[string]interface{}{
		"result":       result,
		"evidence":     evidence,
		"endorsements": endorsements,
	}, nil
}

func getInt32Status(v interface{}) (int32, error) {
	number, ok := v.(json.Number)
	if !ok {
		err := fmt.Errorf("expected json.Number, but got %T", v)
		return 0, err
	}

	i64, err := number.Int64()
	if err != nil {
		return 0, err
	}

	if _, err := proto.Int64ToStatus(i64); err != nil {
		return 0, err
	}

	return int32(i64), nil
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

	updatedStatus, err := getInt32Status(rawUpdate["status"])
	if err != nil {
		return nil, err
	}

	if _, ok = proto.TrustTier_name[updatedStatus]; !ok {
		return nil, fmt.Errorf("not a valid TrustTier value: %d", updatedStatus)
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

		value, err := getInt32Status(rawValue)
		if err != nil {
			err := fmt.Errorf("%w: bad value %q for %q: %v",
				ErrBadOPAResult, rawValue, claim, err)
			return nil, err
		}

		updateTv[claim] = value
	}

	update := map[string]interface{}{
		"status":                         updatedStatus,
		"trust-vector":                   updateTv,
		"veraison-verifier-added-claims": rawUpdate["veraison-verifier-added-claims"],
	}

	return update, nil
}
