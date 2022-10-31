// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/ear"
	"github.com/veraison/services/log"
	mock_deps "github.com/veraison/services/policy/mocks"
	"github.com/veraison/services/proto"
)

func Test_CreateAgent(t *testing.T) {
	v := viper.New()
	v.Set("backend", "opa")

	agent, err := CreateAgent(v, log.Named("test"))
	require.Nil(t, err)

	assert.Equal(t, "opa", agent.GetBackendName())

	v.Set("backend", "nope")

	agent, err = CreateAgent(v, log.Named("test"))
	assert.Nil(t, agent)
	assert.EqualError(t, err, `backend "nope" is not supported`)
}

type AgentEvaluateTestVector struct {
	Name           string
	ExpectedError  string
	ReturnResult   map[string]interface{}
	ReturnError    error
	ExpectedResult *ear.AttestationResult
}

func Test_Agent_Evaluate(t *testing.T) {
	affirmingStatus := ear.TrustTierAffirming
	profile := ear.EatProfile
	timestamp := int64(1666091373)

	vectors := []AgentEvaluateTestVector{
		{
			Name: "success",
			ReturnResult: map[string]interface{}{
				"ear.status":  2,
				"eat_profile": "tag:github.com,2022:veraison/ear",
				"iat":         1666091373,
				"ear.trustworthiness-vector": map[string]interface{}{
					"instance-identity": 0,
					"configuration":     0,
					"executables":       2, // AFFIRMING
					"file-system":       0,
					"hardware":          0,
					"runtime-opaque":    0,
					"storage-opaque":    0,
					"sourced-data":      0,
				},
			},
			ReturnError:   nil,
			ExpectedError: "",
			ExpectedResult: &ear.AttestationResult{
				Status:   &affirmingStatus,
				Profile:  &profile,
				IssuedAt: &timestamp,
				TrustVector: &ear.TrustVector{
					Executables: ear.ApprovedRuntimeClaim,
				},
			},
		},
		{
			Name: "bad status",
			ReturnResult: map[string]interface{}{
				"ear.status":  "MEH",
				"eat_profile": "tag:github.com,2022:veraison/ear",
				"iat":         1666091373,
				"ear.trustworthiness-vector": map[string]interface{}{
					"instance-identity": 0,
					"configuration":     0,
					"executables":       2, // AFFIRMING
					"file-system":       0,
					"hardware":          0,
					"runtime-opaque":    0,
					"storage-opaque":    0,
					"sourced-data":      0,
				},
			},
			ReturnError:    nil,
			ExpectedError:  "invalid values(s) for  'ear.status' from JSON",
			ExpectedResult: nil,
		},
		{
			Name: "bad result, no status",
			ReturnResult: map[string]interface{}{
				"ear.trustworthiness-vector": map[string]interface{}{
					"instance-identity": 0,
					"configuration":     0,
					"executables":       2, // AFFIRMING
					"file-system":       0,
					"hardware":          0,
					"runtime-opaque":    0,
					"storage-opaque":    0,
					"sourced-data":      0,
				},
			},
			ReturnError:    nil,
			ExpectedError:  "backend returned outcome with no status field",
			ExpectedResult: nil,
		},
		{
			Name: "bad result, no trust vector",
			ReturnResult: map[string]interface{}{
				"ear.status":  "affirming",
				"eat_profile": "tag:github.com,2022:veraison/ear",
				"iat":         1666091373,
			},
			ReturnError:    nil,
			ExpectedError:  "backend returned no trust-vector field, or its not a map[string]interface{}",
			ExpectedResult: nil,
		},
		{
			Name: "bad result, bad trust vector",
			ReturnResult: map[string]interface{}{
				"ear.status":  2,
				"eat_profile": "tag:github.com,2022:veraison/ear",
				"iat":         1666091373,
				"ear.trustworthiness-vector": map[string]interface{}{
					"instance-identity": 0,
					"configuration":     0,
					"executables":       2, // AFFIRMING
					"file-system":       0,
					"hardware":          0,
					"runtime-opaque":    0,
					"storage-opaque":    0,
					"sourced-data":      0,
					"wrong-field":       0,
				},
			},
			ReturnError:    nil,
			ExpectedError:  "found unexpected fields: wrong-field",
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
	contraStatus := ear.TrustTierContraindicated
	result := &ear.AttestationResult{
		Status:      &contraStatus,
		Profile:     &profile,
		IssuedAt:    &timestamp,
		TrustVector: &ear.TrustVector{},
	}
	evidence := &proto.EvidenceContext{}

	logger := log.Named("test")

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

		agent := &Agent{Backend: backend, logger: logger}
		res, err := agent.Evaluate(ctx, policy, result, evidence, endorsements)

		if v.ExpectedError == "" {
			require.NoError(t, err)
		} else {
			assert.ErrorContains(t, err, v.ExpectedError)
		}

		if v.ExpectedResult == nil {
			assert.Nil(t, res)
		} else {
			assert.Equal(t, policy.ID, *res.AppraisalPolicyID)
			assert.Equal(t, *v.ExpectedResult.Status, *res.Status)
			assert.Equal(t, v.ExpectedResult.TrustVector.InstanceIdentity,
				res.TrustVector.InstanceIdentity)
			assert.Equal(t, v.ExpectedResult.TrustVector.Configuration,
				res.TrustVector.Configuration)
			assert.Equal(t, v.ExpectedResult.TrustVector.Executables,
				res.TrustVector.Executables)
			assert.Equal(t, v.ExpectedResult.TrustVector.FileSystem,
				res.TrustVector.FileSystem)
			assert.Equal(t, v.ExpectedResult.TrustVector.Hardware,
				res.TrustVector.Hardware)
			assert.Equal(t, v.ExpectedResult.TrustVector.RuntimeOpaque,
				res.TrustVector.RuntimeOpaque)
			assert.Equal(t, v.ExpectedResult.TrustVector.StorageOpaque,
				res.TrustVector.StorageOpaque)
			assert.Equal(t, v.ExpectedResult.TrustVector.SourcedData,
				res.TrustVector.SourcedData)
		}
	}
}
