// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package trustedservices

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/ear"
	"github.com/veraison/services/handler"
	handlermod "github.com/veraison/services/handler"
	"github.com/veraison/services/kvstore"
	"github.com/veraison/services/log"
	"github.com/veraison/services/proto"
	"google.golang.org/protobuf/types/known/structpb"

	mock_deps "github.com/veraison/services/vts/trustedservices/mocks"
)

type mocks struct {
	evidencePluginManager    *mock_deps.MockIManager[handler.IEvidenceHandler]
	storePluginManager       *mock_deps.MockIManager[handler.IStoreHandler]
	endorsementPluginManager *mock_deps.MockIManager[handler.IEndorsementHandler]
	coservProxyPluginManager *mock_deps.MockIManager[handler.ICoservProxyHandler]
	policyManager            *mock_deps.MockIPolicyManager
	earSigner                *mock_deps.MockIEarSigner
	evidenceHandler          *mock_deps.MockIEvidenceHandler
	storeHandler             *mock_deps.MockIStoreHandler
	taStore                  *mock_deps.MockIKVStore
	enStore                  *mock_deps.MockIKVStore
}

func prepareMockGRPC(t *testing.T) (ITrustedServices, *mocks) {
	ctrl := gomock.NewController(t)

	evidencePluginManager := mock_deps.NewMockIManager[handler.IEvidenceHandler](ctrl)
	storePluginManager := mock_deps.NewMockIManager[handler.IStoreHandler](ctrl)
	endorsementPluginManager := mock_deps.NewMockIManager[handler.IEndorsementHandler](ctrl)
	coservProxyPluginManager := mock_deps.NewMockIManager[handler.ICoservProxyHandler](ctrl)
	policyManager := mock_deps.NewMockIPolicyManager(ctrl)
	earSigner := mock_deps.NewMockIEarSigner(ctrl)
	evidenceHandler := mock_deps.NewMockIEvidenceHandler(ctrl)
	storeHandler := mock_deps.NewMockIStoreHandler(ctrl)
	taStore := mock_deps.NewMockIKVStore(ctrl)
	enStore := mock_deps.NewMockIKVStore(ctrl)

	return NewGRPC(
			taStore,
			enStore,
			evidencePluginManager,
			endorsementPluginManager,
			storePluginManager,
			coservProxyPluginManager,
			policyManager,
			earSigner,
			nil, // coservSigner
			log.Named("test"),
		), &mocks{
			evidencePluginManager:    evidencePluginManager,
			storePluginManager:       storePluginManager,
			endorsementPluginManager: endorsementPluginManager,
			coservProxyPluginManager: coservProxyPluginManager,
			policyManager:            policyManager,
			earSigner:                earSigner,
			evidenceHandler:          evidenceHandler,
			storeHandler:             storeHandler,
			taStore:                  taStore,
			enStore:                  enStore,
		}
}

var (
	testMediaType = "application/attestation-token"
	testTenantID  = "tenant-abc"
	testToken     = &proto.AttestationToken{
		TenantId:  testTenantID,
		Data:      []byte("dummy-attestation-token-data"),
		MediaType: testMediaType,
		Nonce:     []byte("random-nonce-value"),
	}
	testSignedEAR    = []byte("signed-ear")
	testScheme       = "test-scheme"
	testTAID         = []string{"test-trust-anchor-ID"}
	testTA           = []string{"ta-xyz"}
	testClaims       = map[string]interface{}{"a": "b"}
	testClaimsSet, _ = structpb.NewStruct(testClaims)
	testRefValID     = []string{"test-refval-id-1"}
	testEndorsement  = []string{"test-endorsement"}
	testEAR          = ear.NewAttestationResult("a-submod", "a-verifier-build", "a-verifier-developer")
)

func TestGRPC_getAttestation_internal_error_evidence_handler_not_found(t *testing.T) {
	grpc, mocks := prepareMockGRPC(t)

	mocks.evidencePluginManager.
		EXPECT().LookupByMediaType(testMediaType).Return(nil, errors.New("evidence handler not found"))

	mocks.earSigner.
		EXPECT().Sign(gomock.Any()).Return(testSignedEAR, nil)

	actual, err := grpc.GetAttestation(context.Background(), testToken)

	assert.EqualError(t, err, "evidence handler not found")

	expected := &proto.AppraisalContext{
		Evidence: &proto.EvidenceContext{
			TenantId: testTenantID,
		},
		Result: testSignedEAR,
	}

	assert.Equal(t, expected, actual)
}

func TestGRPC_GetAttestation_internal_error_no_store_handler(t *testing.T) {
	grpc, mocks := prepareMockGRPC(t)

	mocks.evidencePluginManager.
		EXPECT().LookupByMediaType(testMediaType).Return(mocks.evidenceHandler, nil)

	mocks.evidenceHandler.
		EXPECT().GetAttestationScheme().Return(testScheme)

	mocks.storePluginManager.
		EXPECT().LookupByAttestationScheme(testScheme).Return(nil, errors.New("store handler not found"))

	mocks.earSigner.
		EXPECT().Sign(gomock.Any()).Return(testSignedEAR, nil)

	actual, err := grpc.GetAttestation(context.Background(), testToken)

	assert.EqualError(t, err, "store handler not found")

	expected := &proto.AppraisalContext{
		Evidence: &proto.EvidenceContext{
			TenantId: testTenantID,
		},
		Result: testSignedEAR,
	}

	assert.Equal(t, expected, actual)
}

func TestGRPC_GetAttestation_no_trust_anchor_ids(t *testing.T) {
	grpc, mocks := prepareMockGRPC(t)

	mocks.evidencePluginManager.
		EXPECT().LookupByMediaType(testMediaType).Return(mocks.evidenceHandler, nil)

	mocks.evidenceHandler.
		EXPECT().GetAttestationScheme().Return(testScheme)

	mocks.storePluginManager.
		EXPECT().LookupByAttestationScheme(testScheme).Return(mocks.storeHandler, nil)

	mocks.storeHandler.
		EXPECT().GetTrustAnchorIDs(testToken).Return(nil, errors.New("TA IDs not found"))

	mocks.earSigner.
		EXPECT().Sign(gomock.Any()).Return(testSignedEAR, nil)

	actual, err := grpc.GetAttestation(context.Background(), testToken)

	assert.EqualError(t, err, "TA IDs not found")

	expected := &proto.AppraisalContext{
		Evidence: &proto.EvidenceContext{
			TenantId: testTenantID,
		},
		Result: testSignedEAR,
	}

	assert.Equal(t, expected, actual)
}

func TestGRPC_GetAttestation_ta_store_lookup_failure(t *testing.T) {
	grpc, mocks := prepareMockGRPC(t)

	mocks.evidencePluginManager.
		EXPECT().LookupByMediaType(testMediaType).Return(mocks.evidenceHandler, nil)

	mocks.evidenceHandler.
		EXPECT().GetAttestationScheme().Return(testScheme)

	mocks.storePluginManager.
		EXPECT().LookupByAttestationScheme(testScheme).Return(mocks.storeHandler, nil)

	mocks.storeHandler.
		EXPECT().GetTrustAnchorIDs(testToken).Return(testTAID, nil)

	mocks.taStore.
		EXPECT().Get(testTAID[0]).Return(nil, errors.New("TA not found"))

	mocks.earSigner.
		EXPECT().Sign(gomock.Any()).Return(testSignedEAR, nil)

	actual, err := grpc.GetAttestation(context.Background(), testToken)

	assert.EqualError(t, err, "TA not found")

	expected := &proto.AppraisalContext{
		Evidence: &proto.EvidenceContext{
			TenantId: testTenantID,
		},
		Result: testSignedEAR,
	}

	assert.Equal(t, expected, actual)
}

func TestGRPC_GetAttestation_trust_anchor_not_found(t *testing.T) {
	grpc, mocks := prepareMockGRPC(t)

	mocks.evidencePluginManager.
		EXPECT().LookupByMediaType(testMediaType).Return(mocks.evidenceHandler, nil)

	mocks.evidenceHandler.
		EXPECT().GetAttestationScheme().Return(testScheme)

	mocks.storePluginManager.
		EXPECT().LookupByAttestationScheme(testScheme).Return(mocks.storeHandler, nil)

	mocks.storeHandler.
		EXPECT().GetTrustAnchorIDs(testToken).Return(testTAID, nil)

	mocks.taStore.
		EXPECT().Get(testTAID[0]).Return(nil, kvstore.ErrKeyNotFound)

	mocks.earSigner.
		EXPECT().Sign(gomock.Any()).Return(testSignedEAR, nil)

	actual, err := grpc.GetAttestation(context.Background(), testToken)

	// this is not an internal error, so a nil error is returned to the caller
	assert.NoError(t, err)

	expected := &proto.AppraisalContext{
		Evidence: &proto.EvidenceContext{
			TenantId: testTenantID,
		},
		Result: testSignedEAR,
	}

	assert.Equal(t, expected, actual)
}

func TestGRPC_GetAttestation_extract_claims_failure(t *testing.T) {
	grpc, mocks := prepareMockGRPC(t)

	mocks.evidencePluginManager.
		EXPECT().LookupByMediaType(testMediaType).Return(mocks.evidenceHandler, nil)

	mocks.evidenceHandler.
		EXPECT().GetAttestationScheme().Return(testScheme)

	mocks.storePluginManager.
		EXPECT().LookupByAttestationScheme(testScheme).Return(mocks.storeHandler, nil)

	mocks.storeHandler.
		EXPECT().GetTrustAnchorIDs(testToken).Return(testTAID, nil)

	mocks.taStore.
		EXPECT().Get(testTAID[0]).Return(testTA, nil)

	mocks.evidenceHandler.
		EXPECT().ExtractClaims(testToken, testTA).Return(nil, handlermod.BadEvidenceError{})

	mocks.earSigner.
		EXPECT().Sign(gomock.Any()).Return(testSignedEAR, nil)

	actual, err := grpc.GetAttestation(context.Background(), testToken)

	// this is not an internal error, so a nil error is returned to the caller
	assert.NoError(t, err)

	expected := &proto.AppraisalContext{
		Evidence: &proto.EvidenceContext{
			TenantId:       testTenantID,
			TrustAnchorIds: testTAID,
		},
		Result: testSignedEAR,
	}

	assert.Equal(t, expected, actual)
}

func TestGRPC_GetAttestation_get_reference_value_ids_failure(t *testing.T) {
	grpc, mocks := prepareMockGRPC(t)

	mocks.evidencePluginManager.
		EXPECT().LookupByMediaType(testMediaType).Return(mocks.evidenceHandler, nil)

	mocks.evidenceHandler.
		EXPECT().GetAttestationScheme().Return(testScheme)

	mocks.storePluginManager.
		EXPECT().LookupByAttestationScheme(testScheme).Return(mocks.storeHandler, nil)

	mocks.storeHandler.
		EXPECT().GetTrustAnchorIDs(testToken).Return(testTAID, nil)

	mocks.taStore.
		EXPECT().Get(testTAID[0]).Return(testTA, nil)

	mocks.evidenceHandler.
		EXPECT().ExtractClaims(testToken, testTA).Return(testClaims, nil)

	mocks.storeHandler.
		EXPECT().GetRefValueIDs(testTenantID, testTA, testClaims).Return(nil, errors.New("error retrieving ref value ids"))

	mocks.earSigner.
		EXPECT().Sign(gomock.Any()).Return(testSignedEAR, nil)

	actual, err := grpc.GetAttestation(context.Background(), testToken)

	assert.EqualError(t, err, "error retrieving ref value ids")

	expected := &proto.AppraisalContext{
		Evidence: &proto.EvidenceContext{
			TenantId:       testTenantID,
			TrustAnchorIds: testTAID,
		},
		Result: testSignedEAR,
	}

	assert.Equal(t, expected, actual)
}

func TestGRPC_GetAttestation_get_reference_values_failure(t *testing.T) {
	grpc, mocks := prepareMockGRPC(t)

	mocks.evidencePluginManager.
		EXPECT().LookupByMediaType(testMediaType).Return(mocks.evidenceHandler, nil)

	mocks.evidenceHandler.
		EXPECT().GetAttestationScheme().Return(testScheme)

	mocks.storePluginManager.
		EXPECT().LookupByAttestationScheme(testScheme).Return(mocks.storeHandler, nil)

	mocks.storeHandler.
		EXPECT().GetTrustAnchorIDs(testToken).Return(testTAID, nil)

	mocks.taStore.
		EXPECT().Get(testTAID[0]).Return(testTA, nil)

	mocks.evidenceHandler.
		EXPECT().ExtractClaims(testToken, testTA).Return(testClaims, nil)

	claimsSet, err := structpb.NewStruct(testClaims)
	require.NoError(t, err)

	mocks.storeHandler.
		EXPECT().GetRefValueIDs(testTenantID, testTA, testClaims).Return(testRefValID, nil)

	mocks.enStore.EXPECT().Get(testRefValID[0]).Return(nil, errors.New("store lookup operation failed"))

	mocks.earSigner.
		EXPECT().Sign(gomock.Any()).Return(testSignedEAR, nil)

	actual, err := grpc.GetAttestation(context.Background(), testToken)

	assert.EqualError(t, err, "store lookup operation failed")

	expected := &proto.AppraisalContext{
		Evidence: &proto.EvidenceContext{
			TenantId:       testTenantID,
			TrustAnchorIds: testTAID,
			ReferenceIds:   testRefValID,
			Evidence:       claimsSet,
		},
		Result: testSignedEAR,
	}

	assert.Equal(t, expected, actual)
}

func TestGRPC_GetAttestation_validate_evidence_integrity_failure(t *testing.T) {
	grpc, mocks := prepareMockGRPC(t)

	mocks.evidencePluginManager.
		EXPECT().LookupByMediaType(testMediaType).Return(mocks.evidenceHandler, nil)

	mocks.evidenceHandler.
		EXPECT().GetAttestationScheme().Return(testScheme)

	mocks.storePluginManager.
		EXPECT().LookupByAttestationScheme(testScheme).Return(mocks.storeHandler, nil)

	mocks.storeHandler.
		EXPECT().GetTrustAnchorIDs(testToken).Return(testTAID, nil)

	mocks.taStore.
		EXPECT().Get(testTAID[0]).Return(testTA, nil)

	mocks.evidenceHandler.
		EXPECT().ExtractClaims(testToken, testTA).Return(testClaims, nil)

	mocks.storeHandler.
		EXPECT().GetRefValueIDs(testTenantID, testTA, testClaims).Return(testRefValID, nil)

	mocks.enStore.EXPECT().Get(testRefValID[0]).Return(testEndorsement, nil)

	mocks.evidenceHandler.
		EXPECT().ValidateEvidenceIntegrity(testToken, testTA, testEndorsement).Return(errors.New("evidence validation failure"))

	mocks.earSigner.
		EXPECT().Sign(gomock.Any()).Return(testSignedEAR, nil)

	actual, err := grpc.GetAttestation(context.Background(), testToken)

	assert.EqualError(t, err, "evidence validation failure")

	expected := &proto.AppraisalContext{
		Evidence: &proto.EvidenceContext{
			TenantId:       testTenantID,
			TrustAnchorIds: testTAID,
			ReferenceIds:   testRefValID,
			Evidence:       testClaimsSet,
		},
		Result: testSignedEAR,
	}

	assert.Equal(t, expected, actual)
}

func TestGRPC_GetAttestation_appraise_evidence_failure(t *testing.T) {
	grpc, mocks := prepareMockGRPC(t)

	mocks.evidencePluginManager.
		EXPECT().LookupByMediaType(testMediaType).Return(mocks.evidenceHandler, nil)

	mocks.evidenceHandler.
		EXPECT().GetAttestationScheme().Return(testScheme)

	mocks.storePluginManager.
		EXPECT().LookupByAttestationScheme(testScheme).Return(mocks.storeHandler, nil)

	mocks.storeHandler.
		EXPECT().GetTrustAnchorIDs(testToken).Return(testTAID, nil)

	mocks.taStore.
		EXPECT().Get(testTAID[0]).Return(testTA, nil)

	mocks.evidenceHandler.
		EXPECT().ExtractClaims(testToken, testTA).Return(testClaims, nil)

	mocks.storeHandler.
		EXPECT().GetRefValueIDs(testTenantID, testTA, testClaims).Return(testRefValID, nil)

	mocks.enStore.EXPECT().Get(testRefValID[0]).Return(testEndorsement, nil)

	mocks.evidenceHandler.
		EXPECT().ValidateEvidenceIntegrity(testToken, testTA, testEndorsement).Return(nil)

	evidenceCtx := &proto.EvidenceContext{
		TenantId:       testTenantID,
		TrustAnchorIds: testTAID,
		ReferenceIds:   testRefValID,
		Evidence:       testClaimsSet,
	}

	mocks.evidenceHandler.
		EXPECT().AppraiseEvidence(evidenceCtx, testEndorsement).Return(nil, errors.New("appraise evidence failure"))

	mocks.earSigner.
		EXPECT().Sign(gomock.Any()).Return(testSignedEAR, nil)

	actual, err := grpc.GetAttestation(context.Background(), testToken)

	assert.EqualError(t, err, "appraise evidence failure")

	expected := &proto.AppraisalContext{
		Evidence: evidenceCtx,
		Result:   testSignedEAR,
	}

	assert.Equal(t, expected, actual)
}

func TestGRPC_GetAttestation_policy_failure(t *testing.T) {
	grpc, mocks := prepareMockGRPC(t)

	mocks.evidencePluginManager.
		EXPECT().LookupByMediaType(testMediaType).Return(mocks.evidenceHandler, nil)

	mocks.evidenceHandler.
		EXPECT().GetAttestationScheme().Return(testScheme)

	mocks.storePluginManager.
		EXPECT().LookupByAttestationScheme(testScheme).Return(mocks.storeHandler, nil)

	mocks.storeHandler.
		EXPECT().GetTrustAnchorIDs(testToken).Return(testTAID, nil)

	mocks.taStore.
		EXPECT().Get(testTAID[0]).Return(testTA, nil)

	mocks.evidenceHandler.
		EXPECT().ExtractClaims(testToken, testTA).Return(testClaims, nil)

	mocks.storeHandler.
		EXPECT().GetRefValueIDs(testTenantID, testTA, testClaims).Return(testRefValID, nil)

	mocks.enStore.
		EXPECT().Get(testRefValID[0]).Return(testEndorsement, nil)

	mocks.evidenceHandler.
		EXPECT().ValidateEvidenceIntegrity(testToken, testTA, testEndorsement).Return(nil)

	evidenceCtx := &proto.EvidenceContext{
		TenantId:       testTenantID,
		TrustAnchorIds: testTAID,
		ReferenceIds:   testRefValID,
		Evidence:       testClaimsSet,
	}

	mocks.evidenceHandler.
		EXPECT().AppraiseEvidence(evidenceCtx, testEndorsement).Return(testEAR, nil)

	mocks.policyManager.
		EXPECT().Evaluate(context.Background(), gomock.Any()).Return(errors.New("policy failure")) // XXX: using Any() is lazy

	mocks.earSigner.
		EXPECT().Sign(gomock.Any()).Return(testSignedEAR, nil)

	actual, err := grpc.GetAttestation(context.Background(), testToken)

	assert.EqualError(t, err, "policy failure")

	expected := &proto.AppraisalContext{
		Evidence: evidenceCtx,
		Result:   testSignedEAR,
	}

	assert.Equal(t, expected, actual)
}

func TestGRPC_GetAttestation_ok(t *testing.T) {
	grpc, mocks := prepareMockGRPC(t)

	mocks.evidencePluginManager.
		EXPECT().LookupByMediaType(testMediaType).Return(mocks.evidenceHandler, nil)

	mocks.evidenceHandler.
		EXPECT().GetAttestationScheme().Return(testScheme)

	mocks.storePluginManager.
		EXPECT().LookupByAttestationScheme(testScheme).Return(mocks.storeHandler, nil)

	mocks.storeHandler.
		EXPECT().GetTrustAnchorIDs(testToken).Return(testTAID, nil)

	mocks.taStore.
		EXPECT().Get(testTAID[0]).Return(testTA, nil)

	mocks.evidenceHandler.
		EXPECT().ExtractClaims(testToken, testTA).Return(testClaims, nil)

	mocks.storeHandler.
		EXPECT().GetRefValueIDs(testTenantID, testTA, testClaims).Return(testRefValID, nil)

	mocks.enStore.
		EXPECT().Get(testRefValID[0]).Return(testEndorsement, nil)

	mocks.evidenceHandler.
		EXPECT().ValidateEvidenceIntegrity(testToken, testTA, testEndorsement).Return(nil)

	evidenceCtx := &proto.EvidenceContext{
		TenantId:       testTenantID,
		TrustAnchorIds: testTAID,
		ReferenceIds:   testRefValID,
		Evidence:       testClaimsSet,
	}

	mocks.evidenceHandler.
		EXPECT().AppraiseEvidence(evidenceCtx, testEndorsement).Return(testEAR, nil)

	mocks.policyManager.
		EXPECT().Evaluate(context.Background(), gomock.Any()).Return(nil)

	mocks.earSigner.
		EXPECT().Sign(gomock.Any()).Return(testSignedEAR, nil)

	actual, err := grpc.GetAttestation(context.Background(), testToken)

	assert.NoError(t, err)

	expected := &proto.AppraisalContext{
		Evidence: evidenceCtx,
		Result:   testSignedEAR,
	}

	assert.Equal(t, expected, actual)
}
