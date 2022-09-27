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
	mock_deps "github.com/veraison/services/policy/mocks"
	"github.com/veraison/services/proto"
)

func Test_CreateAgent(t *testing.T) {
	v := viper.New()
	v.Set(DirectiveBackend, "opa")

	agent, err := CreateAgent(v)
	require.Nil(t, err)

	assert.Equal(t, "opa", agent.GetBackendName())

	v.Set(DirectiveBackend, "nope")

	agent, err = CreateAgent(v)
	assert.Nil(t, agent)
	assert.EqualError(t, err, `backend "nope" is not supported`)
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
				"status": 2, // AFFIRMING
				"trust-vector": map[string]interface{}{
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
			ExpectedResult: &proto.AttestationResult{
				Status: proto.TrustTier_AFFIRMING,
				TrustVector: &proto.TrustVector{
					Executables: 2, // AFFIRMING
				},
			},
		},
		{
			Name: "bad status",
			ReturnResult: map[string]interface{}{
				"status": "MEH",
				"trust-vector": map[string]interface{}{
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
			ExpectedError:  "invalid value for enum type: \"MEH\"",
			ExpectedResult: nil,
		},
		{
			Name: "bad result, no status",
			ReturnResult: map[string]interface{}{
				"trust-vector": map[string]interface{}{
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
				"status": "SUCCESS",
			},
			ReturnError:    nil,
			ExpectedError:  "backend returned no trust-vector field, or its not a map[string]interface{}: map[status:SUCCESS]",
			ExpectedResult: nil,
		},
		{
			Name: "bad result, bad trust vector",
			ReturnResult: map[string]interface{}{
				"status": 2, // AFFIRMING
				"trust-vector": map[string]interface{}{
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
		Status:      96, // CONTRAINDICATED
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

		agent := &Agent{Backend: backend}
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
