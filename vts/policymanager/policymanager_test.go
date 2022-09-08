// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policymanager

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/services/kvstore"
	"github.com/veraison/services/policy"
	"github.com/veraison/services/proto"
	mock_deps "github.com/veraison/services/vts/policymanager/mocks"
	"google.golang.org/protobuf/types/known/structpb"
)

func Test_getPolicy_not_found(t *testing.T) {
	ctrl := gomock.NewController(t)

	store := mock_deps.NewMockIKVStore(ctrl)
	store.EXPECT().
		Get(gomock.Eq("opa://0/TPM_ENACTTRUST")).
		Return(nil, kvstore.ErrKeyNotFound)

	backend := mock_deps.NewMockIBackend(ctrl)
	backend.EXPECT().
		GetName().
		Return("opa")

	agent := policy.NewAgent(backend)

	evStruct, err := structpb.NewStruct(nil)
	require.NoError(t, err)

	ec := &proto.EvidenceContext{
		Format:        proto.AttestationFormat_TPM_ENACTTRUST,
		TenantId:      "0",
		TrustAnchorId: "TPM_ENACTTRUST://0/7df7714e-aa04-4638-bcbf-434b1dd720f1",
		SoftwareId:    "TPM_ENACTTRUST://0/7df7714e-aa04-4638-bcbf-434b1dd720f1",
		Evidence:      evStruct,
	}

	pm := &PolicyManager{Store: store, Agent: agent}

	pol, err := pm.getPolicy(ec)
	assert.Nil(t, pol)
	assert.ErrorIs(t, err, ErrNoPolicy)
}
