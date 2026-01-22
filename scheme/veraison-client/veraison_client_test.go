// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package veraisonclient

import (
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/veraison/apiclient/verification"
	mock_deps "github.com/veraison/services/scheme/veraison-client/mocks"
)

var (
	testEvidence       = []byte("test-evidence")
	testNonce          = []byte("test-nonce")
	testSessionURI     = "http://veraison.example/challenge-response/v1/newSession"
	testDiscoveryURI   = "http://veraison.example/.well-known/veraison/verification"
	testECDSAPublicKey = `{
	"kty": "EC",
	"crv": "P-256",
	"x": "usWxHK2PmfnHKwXPS54m0kTcGJ90UiglWiGahtagnv8",
	"y": "IBOL-C3BttVivg-lSreASjpkttcsz-1rb7btKLv8EX4"
}`
	testECDSAPublicKeyJwk, _ = jwk.ParseKey([]byte(testECDSAPublicKey))
	testEAR                  = `eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJlYXIudmVyaWZpZXItaWQiOnsiYnVpbGQiOiJycnRyYXAtdjEuMC4wIiwiZGV2ZWxvcGVyIjoiQWNtZSBJbmMuIn0sImVhdF9wcm9maWxlIjoidGFnOmdpdGh1Yi5jb20sMjAyMzp2ZXJhaXNvbi9lYXIiLCJpYXQiOjE2NjYwOTEzNzMsInN1Ym1vZHMiOnsidGVzdCI6eyJlYXIuYXBwcmFpc2FsLXBvbGljeS1pZCI6InBvbGljeTovL3Rlc3QvMDEyMzQiLCJlYXIuc3RhdHVzIjoiYWZmaXJtaW5nIiwiZWFyLnZlcmFpc29uLmFubm90YXRlZC1ldmlkZW5jZSI6eyJrMSI6InYxIiwiazIiOiJ2MiJ9LCJlYXIudmVyYWlzb24ua2V5LWF0dGVzdGF0aW9uIjp7ImFrcHViIjoiWVd0d2RXSUsifSwiZWFyLnZlcmFpc29uLnBvbGljeS1jbGFpbXMiOnsiYmFyIjoiYmF6IiwiZm9vIjoiYmFyIn19fX0.gTuJrH5Ctf6sAXlaFu1NvHAtI4H0iSqsp2ZtxPPhSfZJBkyeWmZi62lTBw644JDRI0DY9X7Wk7CBWWE6dmBVAA`
	testInvalidEAR           = `a.b.c`
	testEARAppraisals        = `{
	"test": {
		"ear.appraisal-policy-id": "policy://test/01234",
		"ear.status": "affirming",
		"ear.veraison.annotated-evidence": {
			"k1": "v1",
			"k2": "v2"
		},
		"ear.veraison.key-attestation": {
			"akpub": "YWtwdWIK"
		},
		"ear.veraison.policy-claims": {
			"bar": "baz",
			"foo": "bar"
		}
	}
}`
)

func Test_unpackConfig_ok_all_fields(t *testing.T) {
	expectedCfg := &ClientConfig{
		DiscoveryURL: testDiscoveryURI,
		CACerts:      []string{"/path/to/ca1.pem", "/path/to/ca2.pem"},
		Insecure:     true,
	}

	tv := fmt.Sprintf(`{
		"url": %q,
		"ca_certs": ["/path/to/ca1.pem", "/path/to/ca2.pem"],
		"insecure": true
	}`, testDiscoveryURI)

	cfg, err := unpackConfig([]byte(tv))
	assert.NoError(t, err)

	assert.Equal(t, expectedCfg, cfg)
}

func Test_unpackConfig_ok_mandatory_fields_only(t *testing.T) {
	expectedCfg := &ClientConfig{
		DiscoveryURL: testDiscoveryURI,
		CACerts:      nil,
		Insecure:     false,
	}

	tv := fmt.Sprintf(`{ "url": %q }`, testDiscoveryURI)

	cfg, err := unpackConfig([]byte(tv))
	assert.NoError(t, err)

	assert.Equal(t, expectedCfg, cfg)
}

func Test_unpackConfig_err_missing_mandatory_field(t *testing.T) {
	tv := `{
		"ca_certs": ["/path/to/ca1.pem", "/path/to/ca2.pem"],
		"insecure": true
	}`

	_, err := unpackConfig([]byte(tv))
	assert.EqualError(t, err, "missing mandatory URL")
}

func Test_unpackConfig_err_invalid_json(t *testing.T) {
	tv := fmt.Sprintf(`{
		"url": %q,
		"ca_certs": ["/path/to/ca1.pem", "/path/to/ca2.pem"],
		"insecure": true,  // trailing comma causes invalid JSON
	}`, testDiscoveryURI)

	_, err := unpackConfig([]byte(tv))
	assert.Error(t, err)
}

func Test_verifyAttestationResult_ok(t *testing.T) {
	ar, err := verifyAttestationResult([]byte(testEAR), &testECDSAPublicKeyJwk)
	assert.NoError(t, err)
	assert.JSONEq(t, testEARAppraisals, string(ar))
}

func Test_verifyAttestationResult_fail_invalid_jwt(t *testing.T) {
	_, err := verifyAttestationResult([]byte(testInvalidEAR), &testECDSAPublicKeyJwk)
	assert.ErrorContains(t, err, "verifying attestation result signature: failed to parse jws")
}

func Test_discover_ok(t *testing.T) {
	cfg := &ClientConfig{
		DiscoveryURL: testDiscoveryURI,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mc := mock_deps.NewMockIVeraisonDiscoveryClient(ctrl)

	mc.EXPECT().SetDiscoveryURI(testDiscoveryURI)
	mc.EXPECT().Run().Return(&verification.DiscoveryObject{
		PublicKey: []byte(testECDSAPublicKey),
		ApiEndpoints: map[string]string{
			"newChallengeResponseSession": "/challenge-response/v1/newSession",
		},
	}, nil)

	verificationKey, crURL, err := discover(cfg, mc)

	assert.NoError(t, err)
	assert.Equal(t, &testECDSAPublicKeyJwk, verificationKey)
	assert.Equal(t, testSessionURI, crURL)
}

func Test_discover_fail_run(t *testing.T) {
	cfg := &ClientConfig{
		DiscoveryURL: testDiscoveryURI,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mc := mock_deps.NewMockIVeraisonDiscoveryClient(ctrl)

	mc.EXPECT().SetDiscoveryURI(testDiscoveryURI)
	mc.EXPECT().Run().Return(nil, errors.New("some kind of error"))

	expectedErr := `failed to run discovery: some kind of error`

	_, _, err := discover(cfg, mc)
	assert.EqualError(t, err, expectedErr)
}

func Test_remoteAppraisal_ok(t *testing.T) {
	cfg := &ClientConfig{
		crURL: testSessionURI,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mc := mock_deps.NewMockIVeraisonChallengeResponseClient(ctrl)

	mc.EXPECT().SetSessionURI(testSessionURI)
	mc.EXPECT().SetNonce(testNonce)
	mc.EXPECT().SetDeleteSession(true)
	mc.EXPECT().SetIsInsecure(false)
	mc.EXPECT().SetEvidenceBuilder(gomock.Any())
	mc.EXPECT().Run().Return([]byte(testEAR), nil)

	ar, err := remoteAppraisal(testEvidence, "media/type", testNonce, cfg, mc)
	assert.NoError(t, err)
	assert.Equal(t, []byte(testEAR), ar)
}

func Test_remoteAppraisal_fail_run(t *testing.T) {
	cfg := &ClientConfig{
		crURL: testSessionURI,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mc := mock_deps.NewMockIVeraisonChallengeResponseClient(ctrl)

	mc.EXPECT().SetSessionURI(testSessionURI)
	mc.EXPECT().SetNonce(testNonce)
	mc.EXPECT().SetDeleteSession(true)
	mc.EXPECT().SetIsInsecure(false)
	mc.EXPECT().SetEvidenceBuilder(gomock.Any())
	mc.EXPECT().Run().Return(nil, errors.New("some kind of error"))

	expectedErr := `failed to run challenge-response client: some kind of error`

	_, err := remoteAppraisal(testEvidence, "media/type", testNonce, cfg, mc)
	assert.EqualError(t, err, expectedErr)
}

func Test_appraiseComponentEvidence_ok(t *testing.T) {
	cfg := fmt.Sprintf(`{ "url": %q }`, testDiscoveryURI)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mcDiscovery := mock_deps.NewMockIVeraisonDiscoveryClient(ctrl)
	mcChallengeResponse := mock_deps.NewMockIVeraisonChallengeResponseClient(ctrl)

	mcDiscovery.EXPECT().SetDiscoveryURI(testDiscoveryURI)
	mcDiscovery.EXPECT().Run().Return(&verification.DiscoveryObject{
		PublicKey: []byte(testECDSAPublicKey),
		ApiEndpoints: map[string]string{
			"newChallengeResponseSession": "/challenge-response/v1/newSession",
		},
	}, nil)

	mcChallengeResponse.EXPECT().SetSessionURI(testSessionURI)
	mcChallengeResponse.EXPECT().SetNonce(testNonce)
	mcChallengeResponse.EXPECT().SetDeleteSession(true)
	mcChallengeResponse.EXPECT().SetIsInsecure(false)
	mcChallengeResponse.EXPECT().SetEvidenceBuilder(gomock.Any())
	mcChallengeResponse.EXPECT().Run().Return([]byte(testEAR), nil)

	ar, err := appraiseComponentEvidence(
		testEvidence,
		"media/type",
		testNonce,
		[]byte(cfg),
		mcDiscovery,
		mcChallengeResponse,
	)
	assert.NoError(t, err)
	assert.JSONEq(t, testEARAppraisals, string(ar))
}

func Test_appraiseComponentEvidence_fail_verify_ear(t *testing.T) {
	cfg := fmt.Sprintf(`{ "url": %q }`, testDiscoveryURI)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mcDiscovery := mock_deps.NewMockIVeraisonDiscoveryClient(ctrl)
	mcChallengeResponse := mock_deps.NewMockIVeraisonChallengeResponseClient(ctrl)

	mcDiscovery.EXPECT().SetDiscoveryURI(testDiscoveryURI)
	mcDiscovery.EXPECT().Run().Return(&verification.DiscoveryObject{
		PublicKey: []byte(testECDSAPublicKey),
		ApiEndpoints: map[string]string{
			"newChallengeResponseSession": "/challenge-response/v1/newSession",
		},
	}, nil)

	mcChallengeResponse.EXPECT().SetSessionURI(testSessionURI)
	mcChallengeResponse.EXPECT().SetNonce(testNonce)
	mcChallengeResponse.EXPECT().SetDeleteSession(true)
	mcChallengeResponse.EXPECT().SetIsInsecure(false)
	mcChallengeResponse.EXPECT().SetEvidenceBuilder(gomock.Any())
	mcChallengeResponse.EXPECT().Run().Return([]byte(testInvalidEAR), nil)

	_, err := appraiseComponentEvidence(
		testEvidence,
		"media/type",
		testNonce,
		[]byte(cfg),
		mcDiscovery,
		mcChallengeResponse,
	)
	assert.ErrorContains(t, err, "failed to verify attestation result")
}
