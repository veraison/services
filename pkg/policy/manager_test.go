// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

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
	"github.com/veraison/services/pkg/kvstore"
	"github.com/veraison/services/pkg/log"
)

func TestPolicyMgr_getPolicy_not_found(t *testing.T) {
	ctrl := gomock.NewController(t)

	store := NewMockIKVStore(ctrl)
	store.EXPECT().
		Get(gomock.Eq("0:TPM_ENACTTRUST:opa")).
		Return(nil, kvstore.ErrKeyNotFound)

	// Get the Mock Agent here
	agent := NewMockIAgent(ctrl)
	agent.EXPECT().GetBackendName().Return("opa")

	pm := &Manager{Store: &Store{KVStore: store, Logger: log.Named("test")},
		Agent: agent}

	polKey := pm.getPolicyKey("0", "TPM_ENACTTRUST")
	assert.Equal(t, "0:TPM_ENACTTRUST:opa", polKey.String())

	pol, err := pm.getPolicy(polKey)
	assert.Nil(t, pol)
	assert.ErrorIs(t, err, ErrNoPolicy)
}

func TestPolicyMgr_getPolicy_OK(t *testing.T) {
	ctrl := gomock.NewController(t)

	store := NewMockIKVStore(ctrl)
	store.EXPECT().
		Get(gomock.Eq("0:TPM_ENACTTRUST:opa")).
		Return([]string{`{"uuid": "7df7714e-aa04-4638-bcbf-434b1dd720f1", "active": true}`}, nil)

	agent := NewMockIAgent(ctrl)
	agent.EXPECT().GetBackendName().Return("opa")

	pm := &Manager{Store: &Store{KVStore: store}, Agent: agent}

	polKey := pm.getPolicyKey("0", "TPM_ENACTTRUST")
	assert.Equal(t, "0:TPM_ENACTTRUST:opa", polKey.String())

	_, err := pm.getPolicy(polKey)
	require.NoError(t, err)
}

func TestPolicyMgr_New_policyAgent_OK(t *testing.T) {
	ctrl := gomock.NewController(t)

	store := NewMockIKVStore(ctrl)
	v := viper.New()
	v.Set("backend", "opa")

	_, err := NewManager(v, &Store{KVStore: store}, log.Named("test"))
	require.NoError(t, err)
}

func TestPolicyMgr_New_policyAgent_NOK(t *testing.T) {
	ctrl := gomock.NewController(t)

	store := NewMockIKVStore(ctrl)
	v := viper.New()
	v.Set("backend", "nope")

	_, err := NewManager(v, &Store{KVStore: store}, log.Named("test"))
	assert.EqualError(t, err, `backend "nope" is not supported`)
}

func TestPolicyMgr_Evaluate_OK(t *testing.T) {
	ctrl := gomock.NewController(t)

	polID := "policy:TPM_ENACTTRUST"
	endorsements := []*comid.ValueTriple{}
	ar := ear.NewAttestationResult("test", "test", "test")
	tier := ear.TrustTierAffirming
	earAp := ear.Appraisal{Status: &tier, AppraisalPolicyID: &polID}

	store := NewMockIKVStore(ctrl)
	store.EXPECT().
		Get(gomock.Eq("0:TPM_ENACTTRUST:opa")).
		Return([]string{`{"uuid": "7df7714e-aa04-4638-bcbf-434b1dd720f1", "active": true}`}, nil)

	agent := NewMockIAgent(ctrl)
	agent.EXPECT().GetBackendName().Return("opa")
	agent.EXPECT().
		Evaluate(
			context.TODO(),
			gomock.Any(),
			"TPM_ENACTTRUST",
			map[string]any{},
			gomock.Any(),
			"test",
			ar.Submods["test"],
			endorsements,
		).
		Return(&earAp, nil)

	pm := &Manager{
		Store:  &Store{KVStore: store, Logger: log.Named("store")},
		Agent:  agent,
		logger: log.Named("manager"),
	}
	err := pm.Evaluate(context.TODO(), "0", "TPM_ENACTTRUST", map[string]any{}, ar, endorsements)
	require.NoError(t, err)
}

func TestPolicyMgr_Evaluate_NOK(t *testing.T) {
	ctrl := gomock.NewController(t)

	store := NewMockIKVStore(ctrl)
	store.EXPECT().
		Get(gomock.Eq("0:TPM_ENACTTRUST:opa")).
		Return([]string{`{"uuid": "7df7714e-aa04-4638-bcbf-434b1dd720f1", "active": true}`}, nil)

	ar := ear.NewAttestationResult("test", "test", "test")
	expectedErr := errors.New("could not evaluate policy: policy returned bad update")
	agent := NewMockIAgent(ctrl)
	agent.EXPECT().GetBackendName().Return("opa")
	endorsements := []*comid.ValueTriple{}

	agent.EXPECT().Evaluate(
		context.TODO(),
		gomock.Any(),
		"TPM_ENACTTRUST",
		map[string]any{},
		gomock.Any(),
		"test",
		ar.Submods["test"],
		endorsements,
	).Return(nil, expectedErr)

	pm := &Manager{
		Store:  &Store{KVStore: store, Logger: log.Named("store")},
		Agent:  agent,
		logger: log.Named("manager"),
	}
	err := pm.Evaluate(context.TODO(), "0", "TPM_ENACTTRUST", map[string]any{}, ar, endorsements)
	assert.ErrorIs(t, err, expectedErr)

}

func TestPolicyMgr_Evaluate_inactive_policy(t *testing.T) {
	ctrl := gomock.NewController(t)

	endorsements := []*comid.ValueTriple{}
	ar := ear.NewAttestationResult("test", "test", "test")
	store := NewMockIKVStore(ctrl)
	store.EXPECT().
		Get(gomock.Eq("0:TPM_ENACTTRUST:opa")).
		Return([]string{`{"uuid": "7df7714e-aa04-4638-bcbf-434b1dd720f1", "active": false}`}, nil)

	agent := NewMockIAgent(ctrl)
	agent.EXPECT().GetBackendName().Return("opa")

	pm := &Manager{
		Store:  &Store{KVStore: store, Logger: log.Named("store")},
		Agent:  agent,
		logger: log.Named("manager"),
	}
	err := pm.Evaluate(context.TODO(), "0", "TPM_ENACTTRUST", map[string]any{}, ar, endorsements)
	require.NoError(t, err)
}
