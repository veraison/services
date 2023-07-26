// Copyright 2022-2023 Contributors to the Veraison project.
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
	"github.com/veraison/ear"
)

type TestResult struct {
	Error   string                 `json:"error"`
	Outcome *ear.AttestationResult `json:"outcome"`
}

type EvaluateTestVector struct {
	Title            string     `json:"title"`
	Scheme           string     `json:"scheme"`
	ResultPath       string     `json:"result"`
	EvidencePath     string     `json:"evidence"`
	EndorsementsPath string     `json:"endorsements"`
	PolicyPath       string     `json:"policy"`
	Expected         TestResult `json:"expected"`
}

func (o EvaluateTestVector) Run(t *testing.T, ctx context.Context, pa *OPA) {
	resultMap, err := jsonFileToResultMap(o.ResultPath)
	require.NoError(t, err)

	evidenceMap, err := jsonFileToMap(o.EvidencePath)
	require.NoError(t, err)

	endorsements, err := jsonFileToStringSlice(o.EndorsementsPath)
	require.NoError(t, err)

	policy, err := os.ReadFile(o.PolicyPath)
	require.NoError(t, err)

	res, err := pa.Evaluate(ctx, o.Scheme, string(policy), resultMap, evidenceMap, endorsements)
	if o.Expected.Error == "" {
		require.NoError(t, err)
	} else {
		assert.ErrorContains(t, err, o.Expected.Error)
	}

	expected := getUpdateMap(o.Expected.Outcome)
	assert.Equal(t, expected, res)
}

type ValidateTestVector struct {
	Title      string `json:"title"`
	PolicyPath string `json:"policy"`
	Error      string `json:"error"`
}

func (o ValidateTestVector) Run(t *testing.T, ctx context.Context, pa *OPA) {
	policy, err := os.ReadFile(o.PolicyPath)
	require.NoError(t, err)

	err = pa.Validate(ctx, string(policy))
	if o.Error == "" {
		assert.NoError(t, err)
	} else {
		assert.EqualError(t, err, o.Error)
	}
}

func Test_OPA_GetName(t *testing.T) {
	pa, err := NewOPA(nil)
	require.NoError(t, err)
	defer pa.Close()

	assert.Equal(t, "opa", pa.GetName())
}

func Test_OPA_Evaluate(t *testing.T) {
	bytes, err := os.ReadFile("test/evaluate-vectors.json")
	require.NoError(t, err)

	ctx := context.Background()

	pa, err := NewOPA(nil)
	require.NoError(t, err)
	defer pa.Close()

	var vectors []EvaluateTestVector

	err = json.Unmarshal(bytes, &vectors)
	require.NoError(t, err)

	// XXX(setrofim): currently, rego package spits outs messages about unsafe vars and
	// "no index vars", but despite that, seems to be working. This does not happen when
	// running exactly the same policies via stand-alone opa executable.
	// The messages are connected to array comprehensions, but I haven't been able to figure
	// out exactly whats wrong with them. I suspect it might be somehow related to
	// https://github.com/open-policy-agent/opa/issues/3557 (though that is
	// about set comprehensions). In any case, I'm silencing the log for
	// these tests for now to avoid confusion when parsing the tests' results.
	log.SetOutput(nil)

	for _, v := range vectors {
		fmt.Printf("running %q\n", v.Title)
		v.Run(t, ctx, pa)
	}

}

func Test_OPA_Validate(t *testing.T) {
	bytes, err := os.ReadFile("test/validate-vectors.json")
	require.NoError(t, err)

	ctx := context.Background()

	pa, err := NewOPA(nil)
	require.NoError(t, err)
	defer pa.Close()

	var vectors []ValidateTestVector

	err = json.Unmarshal(bytes, &vectors)
	require.NoError(t, err)

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

	var result map[string]interface{}

	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func jsonFileToResultMap(path string) (map[string]interface{}, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result ear.AttestationResult
	err = result.UnmarshalJSON(bytes)
	if err != nil {
		return nil, err
	}

	return result.AsMap(), nil
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

func getUpdateMap(ar *ear.AttestationResult) map[string]interface{} {
	if ar == nil {
		return nil
	}

	status := ear.TrustTierNone

	app := ar.Submods["test"]

	return map[string]interface{}{
		"ear.status": &status,
		"ear.trustworthiness-vector": map[string]interface{}{
			"instance-identity": app.TrustVector.InstanceIdentity,
			"configuration":     app.TrustVector.Configuration,
			"executables":       app.TrustVector.Executables,
			"file-system":       app.TrustVector.FileSystem,
			"hardware":          app.TrustVector.Hardware,
			"runtime-opaque":    app.TrustVector.RuntimeOpaque,
			"storage-opaque":    app.TrustVector.StorageOpaque,
			"sourced-data":      app.TrustVector.SourcedData,
		},
		"ear.veraison.policy-claims": app.VeraisonPolicyClaims,
	}
}
