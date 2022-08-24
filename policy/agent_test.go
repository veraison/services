// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/services/config"
	mock_deps "github.com/veraison/services/policy/mocks"
	"github.com/veraison/services/proto"
)

func Test_CreateAgent(t *testing.T) {
	cfg := config.Store{
		DirectiveBackend: "opa",
	}

	agent, err := CreateAgent(cfg)
	require.Nil(t, err)

	_, ok := agent.Backend.(*OPA)
	if !ok {
		t.Errorf("expected agent to be an instance of OPA, but found %T", agent)
	}

	assert.Equal(t, "opa", agent.GetBackendName())

	cfg = config.Store{} // use default
	agent, err = CreateAgent(cfg)
	require.Nil(t, err)

	_, ok = agent.Backend.(*OPA)
	if !ok {
		t.Errorf("expected agent to be an instance of OPA, but found %T", agent)
	}

	cfg = config.Store{
		DirectiveBackend: "nope",
	}
	agent, err = CreateAgent(cfg)
	assert.Nil(t, agent)
	assert.EqualError(t, err, `backend "nope" is not supported`)

	cfg = config.Store{
		DirectiveBackend: nil,
	}
	agent, err = CreateAgent(cfg)
	assert.Nil(t, agent)
	assert.EqualError(t, err, `loading backend from config: invalidly specified directive "policy.backend": want string, got <nil>`)

}

type AgentEvaluateTestVector struct {
	Name           string
	ExpectedError  string
	ReturnResult   map[string]interface{}
	ReturnError    error
	ExpectedResult *proto.AttestationResult
}

func Test_Agent_Evaluate(t *testing.T) {
	vectors := []AgentEvaluateTestVector{
		{
			Name: "success",
			ReturnResult: map[string]interface{}{
				"status": "SUCCESS",
				"trust-vector": map[string]interface{}{
					"certification-status": "",
					"config-integrity":     "",
					"hw-authenticity":      "",
					"runtime-integrity":    "",
					"sw-integrity":         "SUCCESS",
					"sw-up-to-dateness":    "",
				},
			},
			ReturnError:   nil,
			ExpectedError: "",
			ExpectedResult: &proto.AttestationResult{
				Status: proto.AR_Status_SUCCESS,
				TrustVector: &proto.TrustVector{
					SoftwareIntegrity: proto.AR_Status_SUCCESS,
				},
			},
		},
		{
			Name: "bad status",
			ReturnResult: map[string]interface{}{
				"status": "MEH",
				"trust-vector": map[string]interface{}{
					"certification-status": "",
					"config-integrity":     "",
					"hw-authenticity":      "",
					"runtime-integrity":    "",
					"sw-integrity":         "SUCCESS",
					"sw-up-to-dateness":    "",
				},
			},
			ReturnError:    nil,
			ExpectedError:  "invalid value for enum type: \"MEH\" from JSON {\"status\":\"MEH\",\"trust-vector\":{\"sw-integrity\":\"SUCCESS\"}}",
			ExpectedResult: nil,
		},
		{
			Name: "bad result, no status",
			ReturnResult: map[string]interface{}{
				"trust-vector": map[string]interface{}{
					"certification-status": "",
					"config-integrity":     "",
					"hw-authenticity":      "",
					"runtime-integrity":    "",
					"sw-integrity":         "SUCCESS",
					"sw-up-to-dateness":    "",
				},
			},
			ReturnError:    nil,
			ExpectedError:  "backend returned outcome with no status field: map[trust-vector:map[certification-status: config-integrity: hw-authenticity: runtime-integrity: sw-integrity:SUCCESS sw-up-to-dateness:]]",
			ExpectedResult: nil,
		},
		{
			Name: "bad result, no trust vector",
			ReturnResult: map[string]interface{}{
				"status": "SUCCESS",
			},
			ReturnError:    nil,
			ExpectedError:  "backend returned no trust-vector field, or its not a map[string]interface{}: map[status:SUCCESS]",
			ExpectedResult: nil,
		},
		{
			Name: "bad result, bad trust vector",
			ReturnResult: map[string]interface{}{
				"status": "SUCCESS",
				"trust-vector": map[string]interface{}{
					"certification-status": "",
					"config-integrity":     "",
					"hw-authenticity":      "",
					"wrong-field":          7,
					"sw-integrity":         "SUCCESS",
				},
			},
			ReturnError:    nil,
			ExpectedError:  `unknown field "wrong-field"`,
			ExpectedResult: nil,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	policy := &Policy{
		ID:    "test-policy",
		Rules: "",
	}
	var endorsements []string
	result := &proto.AttestationResult{
		Status:      proto.AR_Status_FAILURE,
		TrustVector: &proto.TrustVector{},
	}
	evidence := &proto.EvidenceContext{}

	for _, v := range vectors {
		fmt.Printf("running %q\n", v.Name)

		backend := mock_deps.NewMockIBackend(ctrl)
		backend.EXPECT().
			Evaluate(gomock.Eq(ctx),
				gomock.Eq(policy.Rules),
				gomock.Any(),
				gomock.Any(),
				gomock.Eq(endorsements)).
			AnyTimes().
			Return(v.ReturnResult, v.ReturnError)

		agent := NewAgent(backend)

		res, err := agent.Evaluate(ctx, policy, result, evidence, endorsements)

		if v.ExpectedError == "" {
			require.NoError(t, err)
		} else {
			assert.ErrorContains(t, err, v.ExpectedError)
		}

		if v.ExpectedResult == nil {
			assert.Nil(t, res)
		} else {
			assert.Equal(t, policy.ID, res.AppraisalPolicyID)
			assert.Equal(t, v.ExpectedResult.Status, res.Status)
			assert.Equal(t, v.ExpectedResult.TrustVector.SoftwareIntegrity,
				res.TrustVector.SoftwareIntegrity)
			assert.Equal(t, v.ExpectedResult.TrustVector.SoftwareUpToDateness,
				res.TrustVector.SoftwareUpToDateness)
			assert.Equal(t, v.ExpectedResult.TrustVector.HardwareAuthenticity,
				res.TrustVector.HardwareAuthenticity)
			assert.Equal(t, v.ExpectedResult.TrustVector.CertificationStatus,
				res.TrustVector.CertificationStatus)
			assert.Equal(t, v.ExpectedResult.TrustVector.ConfigIntegrity,
				res.TrustVector.ConfigIntegrity)
		}
	}
}
