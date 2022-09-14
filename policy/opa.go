// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"

	"github.com/open-policy-agent/opa/rego"
	"github.com/veraison/services/config"
	"github.com/veraison/services/proto"
)

var ErrBadInput = "could not construct policy input: %w"
var ErrBadOPAResult = "wanted map[string]interface{}, but OPA returned: %v"

//go:embed opa.rego
var preambleText string

type OPA struct {
}

func NewOPA(cfg config.Store) (*OPA, error) {
	var o OPA
	if err := o.Init(cfg); err != nil {
		return nil, err
	}
	return &o, nil
}

func (o *OPA) Init(cfg config.Store) error {
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
		return nil, fmt.Errorf(ErrBadInput, err)
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
	resultUpdate, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(ErrBadOPAResult, value)
	}

	if err = validateUpdateValues(resultUpdate); err != nil {
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

func validateUpdateValues(update map[string]interface{}) error {
	if err := checkStatusValue(update["status"]); err != nil {
		return fmt.Errorf("bad \"status\" value: %w", err)
	}

	tv, ok := update["trust-vector"].(map[string]interface{})
	if !ok {
		return fmt.Errorf(
			"bad trust-vector: expected map[string]interface{}, but got %T",
			update["trust-vector"],
		)
	}

	for k, v := range tv {
		if err := checkStatusValue(v); err != nil {
			return fmt.Errorf("bad value for %q: %w", k, err)
		}
	}

	return nil
}

func checkStatusValue(v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("must be a string, but got %T", v)
	}

	// empty string means there was no update to correpsonding key
	if s == "" {
		return nil
	}

	_, ok = proto.AR_Status_value[s]
	if !ok {
		var valid []string
		for i := 0; i < len(proto.AR_Status_name); i++ {
			name := proto.AR_Status_name[int32(i)]
			valid = append(valid, fmt.Sprintf("%q", name))
		}

		return fmt.Errorf("%q is a not a valid status; must be in %v", s, valid)
	}

	return nil
}
