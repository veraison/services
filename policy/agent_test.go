// Copyright 2022-2026 Contributors to the Veraison project.
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
	"github.com/veraison/corim/comid"
	"github.com/veraison/ear"
	"github.com/veraison/services/log"
	mock_deps "github.com/veraison/services/policy/mocks"
	"github.com/veraison/services/vts/appraisal"
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
	Name              string
	ExpectedError     string
	ReturnAppraisal   map[string]any
	ReturnError       error
	ExpectedAppraisal *ear.Appraisal
}

func Test_Agent_Evaluate(t *testing.T) {
	affirmingStatus := ear.TrustTierAffirming

	vectors := []AgentEvaluateTestVector{
		{
			Name: "success",
			ReturnAppraisal: map[string]any{
				"ear.status": 2,
				"ear.trustworthiness-vector": map[string]any{
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
			ExpectedAppraisal: &ear.Appraisal{
				Status: &affirmingStatus,
				TrustVector: &ear.TrustVector{
					Executables: ear.ApprovedRuntimeClaim,
				},
			},
		},
		{
			Name: "bad status",
			ReturnAppraisal: map[string]any{
				"ear.status": "MEH",
				"ear.trustworthiness-vector": map[string]any{
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
			ReturnError:       nil,
			ExpectedError:     "invalid value(s) for 'ear.status' (not a valid TrustTier name: \"MEH\")",
			ExpectedAppraisal: nil,
		},
		{
			Name: "bad result, no status",
			ReturnAppraisal: map[string]any{
				"ear.trustworthiness-vector": map[string]any{
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
			ReturnError:       nil,
			ExpectedError:     "backend returned outcome with no status field",
			ExpectedAppraisal: nil,
		},
		{
			Name: "bad result, no trust vector",
			ReturnAppraisal: map[string]any{
				"ear.status": "affirming",
			},
			ReturnError:       nil,
			ExpectedError:     "backend returned no trust-vector field, or its not a map[string]interface{}",
			ExpectedAppraisal: nil,
		},
		{
			Name: "bad result, bad trust vector",
			ReturnAppraisal: map[string]any{
				"ear.status": 2,
				"ear.trustworthiness-vector": map[string]any{
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
			ReturnError:       nil,
			ExpectedError:     "invalid value(s) for 'ear.trustworthiness-vector' (unexpected: wrong-field)",
			ExpectedAppraisal: nil,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	policy := &Policy{
		StoreKey: PolicyKey{"test-tenant", "test-scheme", "test-policy"},
		Rules:    "",
	}

	endorsements := []*comid.ValueTriple{}
	contraStatus := ear.TrustTierContraindicated
	polID := "policy:test-scheme"
	appraisalContext := &appraisal.Context{
	}
	appraisal := &ear.Appraisal{
		Status:            &contraStatus,
		TrustVector:       &ear.TrustVector{},
		AppraisalPolicyID: &polID,
	}

	logger := log.Named("test")

	for _, v := range vectors {
		fmt.Printf("running %q\n", v.Name)

		backend := mock_deps.NewMockIBackend(ctrl)
		backend.EXPECT().
			Evaluate(gomock.Eq(ctx),
				gomock.Any(),
				gomock.Any(),
				gomock.Eq(policy.Rules),
				gomock.Any(),
				gomock.Any(),
				gomock.Eq([]map[string]any{})).
			AnyTimes().
			Return(v.ReturnAppraisal, v.ReturnError)

		agent := &Agent{Backend: backend, logger: logger}
		res, err := agent.Evaluate(
			ctx,
			map[string]any{},
			appraisalContext,
			policy,
			"test",
			appraisal,
			endorsements,
		)

		if v.ExpectedError == "" {
			require.NoError(t, err)
		} else {
			assert.ErrorContains(t, err, v.ExpectedError)
		}

		if v.ExpectedAppraisal == nil {
			assert.Nil(t, res)
		} else {
			assert.Equal(t, *appraisal.AppraisalPolicyID, *res.AppraisalPolicyID)
			assert.Equal(t, *v.ExpectedAppraisal.Status, *res.Status)
			assert.Equal(t, v.ExpectedAppraisal.TrustVector.InstanceIdentity,
				res.TrustVector.InstanceIdentity)
			assert.Equal(t, v.ExpectedAppraisal.TrustVector.Configuration,
				res.TrustVector.Configuration)
			assert.Equal(t, v.ExpectedAppraisal.TrustVector.Executables,
				res.TrustVector.Executables)
			assert.Equal(t, v.ExpectedAppraisal.TrustVector.FileSystem,
				res.TrustVector.FileSystem)
			assert.Equal(t, v.ExpectedAppraisal.TrustVector.Hardware,
				res.TrustVector.Hardware)
			assert.Equal(t, v.ExpectedAppraisal.TrustVector.RuntimeOpaque,
				res.TrustVector.RuntimeOpaque)
			assert.Equal(t, v.ExpectedAppraisal.TrustVector.StorageOpaque,
				res.TrustVector.StorageOpaque)
			assert.Equal(t, v.ExpectedAppraisal.TrustVector.SourcedData,
				res.TrustVector.SourcedData)
		}
	}
}
