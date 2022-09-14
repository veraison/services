// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestResult struct {
	Error   string                 `json:"error"`
	Outcome map[string]interface{} `json:"outcome"`
}

type TestVector struct {
	Title            string     `json:"title"`
	ResultPath       string     `json:"result"`
	EvidencePath     string     `json:"evidence"`
	EndorsementsPath string     `json:"endorsements"`
	PolicyPath       string     `json:"policy"`
	Expected         TestResult `json:"expected"`
}

func (o TestVector) Run(t *testing.T, ctx context.Context, pa *OPA) {
	resultMap, err := jsonFileToMap(o.ResultPath)
	require.NoError(t, err)

	evidenceMap, err := jsonFileToMap(o.EvidencePath)
	require.NoError(t, err)

	endorsements, err := jsonFileToStringSlice(o.EndorsementsPath)
	require.NoError(t, err)

	policy, err := os.ReadFile(o.PolicyPath)
	require.NoError(t, err)

	res, err := pa.Evaluate(ctx, string(policy), resultMap, evidenceMap, endorsements)
	if o.Expected.Error == "" {
		require.NoError(t, err)
	} else {
		assert.EqualError(t, err, o.Expected.Error)
	}

	assert.Equal(t, o.Expected.Outcome, res)
}

func Test_OPA_GetName(t *testing.T) {
	pa, err := NewOPA(nil)
	require.NoError(t, err)
	defer pa.Close()

	assert.Equal(t, "opa", pa.GetName())
}

func Test_OPA_Evaluate(t *testing.T) {
	bytes, err := os.ReadFile("test/vectors.json")
	require.NoError(t, err)

	ctx := context.Background()

	pa, err := NewOPA(nil)
	require.NoError(t, err)
	defer pa.Close()

	var vectors []TestVector

	err = json.Unmarshal(bytes, &vectors)
	require.NoError(t, err)

	// XXX(setrofim): currently, rego packages spits outs messages about unsafe vars and
	// "no index vars", but despite that, seems to be working. This does not happend when
	// running exactly the same policies via stand-alone opa executable.
	// The messages are connected to array comprehensions, but I haven't been able to figure
	// out exactly whats wrong with them. I suspect it might be somehow related to
	// https://github.com/open-policy-agent/opa/issues/3557 (though that is
	// about set comprehensions). In any case, I'm silencing the log for
	// these tests for now to avoid confusiong when parsing the tests'
	// results.
	log.SetOutput(nil)

	for _, v := range vectors {
		fmt.Printf("running %q\n", v.Title)
		v.Run(t, ctx, pa)
	}

}

func jsonFileToMap(path string) (map[string]interface{}, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func jsonFileToStringSlice(path string) ([]string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result []string
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
