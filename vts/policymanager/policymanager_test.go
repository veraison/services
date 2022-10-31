// Copyright 2022 Contributors to the Veraison project.
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
	"github.com/veraison/ear"
	"github.com/veraison/services/kvstore"
	"github.com/veraison/services/log"
	"github.com/veraison/services/policy"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/vts/appraisal"
	mock_deps "github.com/veraison/services/vts/policymanager/mocks"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestPolicyMgr_getPolicy_not_found(t *testing.T) {
	ctrl := gomock.NewController(t)

	store := mock_deps.NewMockIKVStore(ctrl)
	store.EXPECT().
		Get(gomock.Eq("opa://0")).
		Return(nil, kvstore.ErrKeyNotFound)

	// Get the Mock Agent here
	agent := mock_deps.NewMockIAgent(ctrl)
	agent.EXPECT().GetBackendName().Return("opa")
	evStruct, err := structpb.NewStruct(nil)
	require.NoError(t, err)

	ec := &proto.EvidenceContext{
		Format:        proto.AttestationFormat_TPM_ENACTTRUST,
		TenantId:      "0",
		TrustAnchorId: "TPM_ENACTTRUST://0/7df7714e-aa04-4638-bcbf-434b1dd720f1",
		SoftwareId:    "TPM_ENACTTRUST://0/7df7714e-aa04-4638-bcbf-434b1dd720f1",
		Evidence:      evStruct,
	}

	pm := &PolicyManager{Store: &policy.Store{KVStore: store, Logger: log.Named("test")},
		Agent: agent}

	polID := pm.getPolicyID(ec)
	assert.Equal(t, "opa://0", polID)

	pol, err := pm.getPolicy(polID)
	assert.Nil(t, pol)
	assert.ErrorIs(t, err, policy.ErrNoPolicy)
}

func TestPolicyMgr_getPolicy_OK(t *testing.T) {
	ctrl := gomock.NewController(t)

	store := mock_deps.NewMockIKVStore(ctrl)
	store.EXPECT().
		Get(gomock.Eq("opa://0")).
		Return([]string{"{}"}, nil)

	agent := mock_deps.NewMockIAgent(ctrl)
	agent.EXPECT().GetBackendName().Return("opa")
	evStruct, err := structpb.NewStruct(nil)
	require.NoError(t, err)

	ec := &proto.EvidenceContext{
		Format:        proto.AttestationFormat_TPM_ENACTTRUST,
		TenantId:      "0",
		TrustAnchorId: "TPM_ENACTTRUST://0/7df7714e-aa04-4638-bcbf-434b1dd720f1",
		SoftwareId:    "TPM_ENACTTRUST://0/7df7714e-aa04-4638-bcbf-434b1dd720f1",
		Evidence:      evStruct,
	}

	pm := &PolicyManager{Store: &policy.Store{KVStore: store}, Agent: agent}

	polID := pm.getPolicyID(ec)
	assert.Equal(t, "opa://0", polID)

	_, err = pm.getPolicy(polID)
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
	evStruct, _ := structpb.NewStruct(nil)

	store := mock_deps.NewMockIKVStore(ctrl)
	store.EXPECT().
		Get(gomock.Eq("opa://0")).
		Return([]string{"{}"}, nil)

	ec := &proto.EvidenceContext{
		Format:        proto.AttestationFormat_TPM_ENACTTRUST,
		TenantId:      "0",
		TrustAnchorId: "TPM_ENACTTRUST://0/7df7714e-aa04-4638-bcbf-434b1dd720f1",
		SoftwareId:    "TPM_ENACTTRUST://0/7df7714e-aa04-4638-bcbf-434b1dd720f1",
		Evidence:      evStruct,
	}
	endorsements := []string{"h0KPxSKAPTEGXnvOPPA/5HUJZjHl4Hu9eg/eYMTPJcc="}
	ar := ear.NewAttestationResult()
	ap := &appraisal.Appraisal{EvidenceContext: ec, Result: ar}

	agent := mock_deps.NewMockIAgent(ctrl)
	agent.EXPECT().GetBackendName().Return("opa")
	agent.EXPECT().Evaluate(context.TODO(), gomock.Any(), ar, ec, endorsements)
	pm := &PolicyManager{Store: &policy.Store{KVStore: store}, Agent: agent}
	err := pm.Evaluate(context.TODO(), ap, endorsements)
	require.NoError(t, err)
}

func TestPolicyMgr_Evaluate_NOK(t *testing.T) {
	ctrl := gomock.NewController(t)
	evStruct, _ := structpb.NewStruct(nil)

	store := mock_deps.NewMockIKVStore(ctrl)
	store.EXPECT().
		Get(gomock.Eq("opa://0")).
		Return([]string{"{}"}, nil)

	ec := &proto.EvidenceContext{
		Format:        proto.AttestationFormat_TPM_ENACTTRUST,
		TenantId:      "0",
		TrustAnchorId: "TPM_ENACTTRUST://0/7df7714e-aa04-4638-bcbf-434b1dd720f1",
		SoftwareId:    "TPM_ENACTTRUST://0/7df7714e-aa04-4638-bcbf-434b1dd720f1",
		Evidence:      evStruct,
	}
	endorsements := []string{"h0KPxSKAPTEGXnvOPPA/5HUJZjHl4Hu9eg/eYMTPJcc="}
	ar := ear.NewAttestationResult()
	ap := &appraisal.Appraisal{EvidenceContext: ec, Result: ar}
	expectedErr := errors.New("could not evaluate policy: policy returned bad update")
	agent := mock_deps.NewMockIAgent(ctrl)
	agent.EXPECT().GetBackendName().Return("opa")
	agent.EXPECT().Evaluate(context.TODO(), gomock.Any(), ar, ec, endorsements).Return(nil, expectedErr)
	pm := &PolicyManager{Store: &policy.Store{KVStore: store}, Agent: agent}
	err := pm.Evaluate(context.TODO(), ap, endorsements)
	assert.ErrorIs(t, err, expectedErr)

}
