// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policymanager

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/corim/comid"
	"github.com/veraison/ear"
	"github.com/veraison/services/kvstore"
	"github.com/veraison/services/log"
	"github.com/veraison/services/policy"
	"github.com/veraison/services/vts/appraisal"
	mock_deps "github.com/veraison/services/vts/policymanager/mocks"
)

func TestPolicyMgr_getPolicy_not_found(t *testing.T) {
	ctrl := gomock.NewController(t)

	store := mock_deps.NewMockIKVStore(ctrl)
	store.EXPECT().
		Get(gomock.Eq("0:TPM_ENACTTRUST:opa")).
		Return(nil, kvstore.ErrKeyNotFound)

	// Get the Mock Agent here
	agent := mock_deps.NewMockIAgent(ctrl)
	agent.EXPECT().GetBackendName().Return("opa")

	appraisal := &appraisal.Context{
		Scheme: "TPM_ENACTTRUST",
		Evidence: &appraisal.Evidence{
			TenantID: "0",
		},
	}

	pm := &PolicyManager{Store: &policy.Store{KVStore: store, Logger: log.Named("test")},
		Agent: agent}

	polKey := pm.getPolicyKey(appraisal)
	assert.Equal(t, "0:TPM_ENACTTRUST:opa", polKey.String())

	pol, err := pm.getPolicy(polKey)
	assert.Nil(t, pol)
	assert.ErrorIs(t, err, policy.ErrNoPolicy)
}

func TestPolicyMgr_getPolicy_OK(t *testing.T) {
	ctrl := gomock.NewController(t)

	store := mock_deps.NewMockIKVStore(ctrl)
	store.EXPECT().
		Get(gomock.Eq("0:TPM_ENACTTRUST:opa")).
		Return([]string{`{"uuid": "7df7714e-aa04-4638-bcbf-434b1dd720f1", "active": true}`}, nil)

	agent := mock_deps.NewMockIAgent(ctrl)
	agent.EXPECT().GetBackendName().Return("opa")

	appraisal := &appraisal.Context{
		Scheme: "TPM_ENACTTRUST",
		Evidence: &appraisal.Evidence{
			TenantID: "0",
		},
	}

	pm := &PolicyManager{Store: &policy.Store{KVStore: store}, Agent: agent}

	polKey := pm.getPolicyKey(appraisal)
	assert.Equal(t, "0:TPM_ENACTTRUST:opa", polKey.String())

	_, err := pm.getPolicy(polKey)
	require.NoError(t, err)
}

func TestPolicyMgr_New_policyAgent_OK(t *testing.T) {
	ctrl := gomock.NewController(t)

	store := mock_deps.NewMockIKVStore(ctrl)
	v := viper.New()
	v.Set("backend", "opa")

	_, err := New(v, &policy.Store{KVStore: store}, log.Named("test"))
	require.NoError(t, err)
}

func TestPolicyMgr_New_policyAgent_NOK(t *testing.T) {
	ctrl := gomock.NewController(t)

	store := mock_deps.NewMockIKVStore(ctrl)
	v := viper.New()
	v.Set("backend", "nope")

	_, err := New(v, &policy.Store{KVStore: store}, log.Named("test"))
	assert.EqualError(t, err, `backend "nope" is not supported`)
}

func TestPolicyMgr_Evaluate_OK(t *testing.T) {
	ctrl := gomock.NewController(t)

	polID := "policy:TPM_ENACTTRUST"
	endorsements := []*comid.ValueTriple{}
	ar := ear.NewAttestationResult("test", "test", "test")
	tier := ear.TrustTierAffirming
	earAp := ear.Appraisal{Status: &tier, AppraisalPolicyID: &polID}
	appraisalContext := &appraisal.Context{
		Scheme: "TPM_ENACTTRUST",
		Evidence: &appraisal.Evidence{
			TenantID: "0",
		},
		Result: ar,
	}

	store := mock_deps.NewMockIKVStore(ctrl)
	store.EXPECT().
		Get(gomock.Eq("0:TPM_ENACTTRUST:opa")).
		Return([]string{`{"uuid": "7df7714e-aa04-4638-bcbf-434b1dd720f1", "active": true}`}, nil)

	agent := mock_deps.NewMockIAgent(ctrl)
	agent.EXPECT().GetBackendName().Return("opa")
	agent.EXPECT().
		Evaluate(
			context.TODO(),
			gomock.Any(),
			appraisalContext,
			gomock.Any(),
			"test",
			ar.Submods["test"],
			endorsements,
		).
		Return(&earAp, nil)

	pm := &PolicyManager{
		Store:  &policy.Store{KVStore: store, Logger: log.Named("store")},
		Agent:  agent,
		logger: log.Named("manager"),
	}
	err := pm.Evaluate(context.TODO(), appraisalContext, endorsements)
	require.NoError(t, err)
}

func TestPolicyMgr_Evaluate_NOK(t *testing.T) {
	ctrl := gomock.NewController(t)

	store := mock_deps.NewMockIKVStore(ctrl)
	store.EXPECT().
		Get(gomock.Eq("0:TPM_ENACTTRUST:opa")).
		Return([]string{`{"uuid": "7df7714e-aa04-4638-bcbf-434b1dd720f1", "active": true}`}, nil)

	ar := ear.NewAttestationResult("test", "test", "test")
	expectedErr := errors.New("could not evaluate policy: policy returned bad update")
	agent := mock_deps.NewMockIAgent(ctrl)
	agent.EXPECT().GetBackendName().Return("opa")
	endorsements := []*comid.ValueTriple{}
	appraisalContext := &appraisal.Context{
		Scheme: "TPM_ENACTTRUST",
		Evidence: &appraisal.Evidence{
			TenantID: "0",
		},
		Result: ar,
	}

	agent.EXPECT().Evaluate(
		context.TODO(),
		gomock.Any(),
		appraisalContext,
		gomock.Any(),
		"test",
		ar.Submods["test"],
		endorsements,
	).Return(nil, expectedErr)

	pm := &PolicyManager{
		Store:  &policy.Store{KVStore: store, Logger: log.Named("store")},
		Agent:  agent,
		logger: log.Named("manager"),
	}
	err := pm.Evaluate(context.TODO(), appraisalContext, endorsements)
	assert.ErrorIs(t, err, expectedErr)

}
