// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package arm_cca

import (
	"encoding/base64"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/ear"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/vts/appraisal"
)

// Helper: testNonce matches the realm challenge in test/evidence/cca-good.cbor
var testNonce = func() []byte {
	b, _ := base64.StdEncoding.DecodeString("byTWuWNaLIu/WOkIuU4Ewb+zroDN6+gyQkV4SZ/jF2Hn9eHYvOASGET1Sr36UobaiPU6ZXsVM1yTlrQyklS8XA==")
	return b
}()

// Helper: loadEvidenceToken loads the CCA token CBOR from test/evidence/cca-good.cbor
func loadEvidenceToken(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("test/evidence/cca-good.cbor")
	require.NoError(t, err, "failed to read evidence token file")
	return data
}

// Helper: loadTrustAnchors extracts KeyTriple from a platform CoRIM's attester-verification-keys.
func loadTrustAnchors(t *testing.T, corimData []byte) []*comid.KeyTriple {
	t.Helper()
	var corimUnsigned corim.UnsignedCorim
	err := corimUnsigned.FromCBOR(corimData)
	assert.Nil(t, err)

	var kt []*comid.KeyTriple

	for i := range corimUnsigned.Tags {
		var cor comid.Comid
		err = cor.FromCBOR(corimUnsigned.Tags[i].Content)
		require.NoError(t, err)

		if cor.Triples.AttestVerifKeys == nil {
			continue
		}

		for k := range *cor.Triples.AttestVerifKeys {
			kt = append(kt, &(*cor.Triples.AttestVerifKeys)[k])
		}
	}
	return kt
}

// Helper: loadReferenceValues extracts ValueTriple from a CoRIM (reference values).
func loadReferenceValues(t *testing.T, corimData []byte) []*comid.ValueTriple {
	t.Helper()
	var corimUnsigned corim.UnsignedCorim
	err := corimUnsigned.FromCBOR(corimData)
	assert.Nil(t, err)

	var vt []*comid.ValueTriple

	for i := range corimUnsigned.Tags {
		var cor comid.Comid
		err = cor.FromCBOR(corimUnsigned.Tags[i].Content)
		require.NoError(t, err)

		if cor.Triples.ReferenceValues == nil {
			continue
		}

		for v := range cor.Triples.ReferenceValues.Values {
			vt = append(vt, &cor.Triples.ReferenceValues.Values[v])
		}
	}
	return vt
}

func TestNewImplementation(t *testing.T) {
	impl := NewImplementation()
	assert.NotNil(t, impl)
	assert.Equal(t, "ARM_CCA", Descriptor.Name)
}

func TestImplementation_GetTrustAnchorIDs(t *testing.T) {
	impl := NewImplementation()
	evidence := &appraisal.Evidence{
		Data:  loadEvidenceToken(t),
		Nonce: testNonce,
	}
	envs, err := impl.GetTrustAnchorIDs(evidence)
	require.NoError(t, err)
	require.Len(t, envs, 1)
	assert.NotNil(t, envs[0].Class)
	assert.NotNil(t, envs[0].Instance)
}

func TestImplementation_ExtractClaims(t *testing.T) {
	impl := NewImplementation()
	evidence := &appraisal.Evidence{
		Data:  loadEvidenceToken(t),
		Nonce: testNonce,
	}
	claims, err := impl.ExtractClaims(evidence, nil)
	require.NoError(t, err)
	assert.Contains(t, claims, "platform")
	assert.Contains(t, claims, "realm")
}

func TestImplementation_ExtractClaims_Empty_Evidence(t *testing.T) {
	impl := NewImplementation()
	evidence := &appraisal.Evidence{
		Data:  []byte("bad evidence"),
		Nonce: testNonce,
	}
	_, err := impl.ExtractClaims(evidence, nil)
	assert.Error(t, err)
	_, ok := err.(handler.BadEvidenceError)
	assert.True(t, ok)
}

func TestImplementation_GetReferenceValueIDs(t *testing.T) {
	impl := NewImplementation()
	evidence := &appraisal.Evidence{
		Data:  loadEvidenceToken(t),
		Nonce: testNonce,
	}
	claims, err := impl.ExtractClaims(evidence, nil)
	require.NoError(t, err)

	// Load trust anchors to get KeyTriple objects
	trustAnchors := loadTrustAnchors(t, corimCcaPlatformValid)
	require.NotEmpty(t, trustAnchors)

	refIDs, err := impl.GetReferenceValueIDs(trustAnchors, claims)
	require.NoError(t, err)
	require.Len(t, refIDs, 2)
	assert.NotNil(t, refIDs[0].Class)
	assert.NotNil(t, refIDs[1].Class)
}

func TestImplementation_ValidateEvidenceIntegrity_Valid(t *testing.T) {
	impl := NewImplementation()
	evidence := &appraisal.Evidence{
		Data:  loadEvidenceToken(t),
		Nonce: testNonce,
	}
	trustAnchors := loadTrustAnchors(t, corimCcaPlatformValid)
	require.NotEmpty(t, trustAnchors)

	err := impl.ValidateEvidenceIntegrity(evidence, trustAnchors, nil)
	assert.NoError(t, err)
}

func TestImplementation_ValidateEvidenceIntegrity_BadKey(t *testing.T) {
	impl := NewImplementation()
	evidence := &appraisal.Evidence{
		Data:  loadEvidenceToken(t),
		Nonce: testNonce,
	}
	trustAnchors := loadTrustAnchors(t, corimCcaPlatformBadTaCert)
	require.NotEmpty(t, trustAnchors)

	err := impl.ValidateEvidenceIntegrity(evidence, trustAnchors, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to verify platform token")
}

func TestImplementation_ValidateEvidenceIntegrity_NonceMismatch(t *testing.T) {
	impl := NewImplementation()
	evidence := &appraisal.Evidence{
		Data:  loadEvidenceToken(t),
		Nonce: []byte("wrongnonce"),
	}
	trustAnchors := loadTrustAnchors(t, corimCcaPlatformValid)
	require.NotEmpty(t, trustAnchors)

	err := impl.ValidateEvidenceIntegrity(evidence, trustAnchors, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "freshness")
}

func TestImplementation_AppraiseClaims_Platform_Valid(t *testing.T) {
	impl := NewImplementation()
	claims, err := extractClaimsFromEvidence(t)
	require.NoError(t, err)

	endorsements := loadReferenceValues(t, corimCcaPlatformValid)
	result, err := impl.AppraiseClaims(claims, endorsements)
	require.NoError(t, err)

	platformAppraisal := result.Submods["CCA_SSD_PLATFORM"]
	assert.NotNil(t, platformAppraisal)
	assert.Equal(t, ear.TrustTierAffirming, *platformAppraisal.Status)
	assert.Equal(t, ear.ApprovedConfigClaim, platformAppraisal.TrustVector.Configuration)
	assert.Equal(t, ear.ApprovedRuntimeClaim, platformAppraisal.TrustVector.Executables)
}

func TestImplementation_AppraiseClaims_Platform_No_Platform_Claims(t *testing.T) {
	impl := NewImplementation()
	var claims map[string]any

	endorsements := loadReferenceValues(t, corimCcaPlatformValid)
	_, err := impl.AppraiseClaims(claims, endorsements)
	assert.Error(t, err)

	assert.Contains(t, err.Error(), "no \"platform\" entry in claims")
}

func TestImplementation_AppraiseClaims_Platform_No_Realm_Claims(t *testing.T) {
	impl := NewImplementation()
	claims, err := extractClaimsFromEvidence(t)
	require.NoError(t, err)

	pclaims := map[string]any{}
	pclaims["platform"] = claims["platform"]

	endorsements := loadReferenceValues(t, corimCcaPlatformValid)
	_, err = impl.AppraiseClaims(pclaims, endorsements)
	assert.Error(t, err)

	assert.Contains(t, err.Error(), "no \"realm\" entry in claims")
}

func TestImplementation_AppraiseClaims_Platform_MismatchRefVal(t *testing.T) {
	impl := NewImplementation()
	claims, err := extractClaimsFromEvidence(t)
	require.NoError(t, err)

	endorsements := loadReferenceValues(t, corimCcaPlatformBadRefvalMkey)
	_, err = impl.AppraiseClaims(claims, endorsements)
	assert.Contains(t, err.Error(), "measurement mkey must be string")
}

func TestImplementation_AppraiseClaims_Platform_No_MKey(t *testing.T) {
	impl := NewImplementation()
	claims, err := extractClaimsFromEvidence(t)
	require.NoError(t, err)

	endorsements := loadReferenceValues(t, corimCcaPlatformBadRefvalNoMkey)
	_, err = impl.AppraiseClaims(claims, endorsements)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "measurement missing mkey")
}

func TestImplementation_AppraiseClaims_Realm_Valid(t *testing.T) {
	impl := NewImplementation()
	claims, err := extractClaimsFromEvidence(t)
	require.NoError(t, err)

	endorsements := loadReferenceValues(t, corimCcaRealmValid)
	result, err := impl.AppraiseClaims(claims, endorsements)
	require.NoError(t, err)

	realmAppraisal := result.Submods["CCA_REALM"]
	assert.NotNil(t, realmAppraisal)
	assert.Equal(t, ear.TrustTierAffirming, *realmAppraisal.Status)
	assert.Equal(t, ear.ApprovedRuntimeClaim, realmAppraisal.TrustVector.Executables)
}

func TestImplementation_AppraiseClaims_Realm_No_Raw_Value(t *testing.T) {
	impl := NewImplementation()
	claims, err := extractClaimsFromEvidence(t)
	require.NoError(t, err)

	endorsements := loadReferenceValues(t, corimCcaRealmBadNoRawValue)
	_, err = impl.AppraiseClaims(claims, endorsements)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "raw-value")
}

func TestImplementation_AppraiseClaims_Realm_No_Integ_Regs(t *testing.T) {
	impl := NewImplementation()
	claims, err := extractClaimsFromEvidence(t)
	require.NoError(t, err)

	endorsements := loadReferenceValues(t, corimCcaRealmBadNoIntegRegs)
	_, err = impl.AppraiseClaims(claims, endorsements)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no digests")
}

func TestImplementation_AppraiseClaims_Realm_NoEndorsements(t *testing.T) {
	impl := NewImplementation()
	claims, err := extractClaimsFromEvidence(t)
	require.NoError(t, err)

	result, err := impl.AppraiseClaims(claims, nil)
	require.NoError(t, err)

	realmAppraisal := result.Submods["CCA_REALM"]
	assert.Equal(t, ear.TrustTierWarning, *realmAppraisal.Status)
	assert.Equal(t, ear.UnrecognizedRuntimeClaim, realmAppraisal.TrustVector.Executables)
}

func TestImplementation_AppraiseClaims_Realm_MismatchRIM(t *testing.T) {
	impl := NewImplementation()
	claims, err := extractClaimsFromEvidence(t)
	require.NoError(t, err)

	endorsements := loadReferenceValues(t, corimCcaRealmBadNoRim)
	result, err := impl.AppraiseClaims(claims, endorsements)
	require.NoError(t, err)

	realmAppraisal := result.Submods["CCA_REALM"]
	assert.Equal(t, ear.TrustTierWarning, *realmAppraisal.Status)
	assert.Equal(t, ear.NoClaim, realmAppraisal.TrustVector.Hardware)
	assert.Equal(t, ear.TrustworthyInstanceClaim, realmAppraisal.TrustVector.InstanceIdentity)
	assert.Equal(t, ear.NoClaim, realmAppraisal.TrustVector.RuntimeOpaque)
	assert.Equal(t, ear.NoClaim, realmAppraisal.TrustVector.StorageOpaque)
	assert.Equal(t, ear.NoClaim, realmAppraisal.TrustVector.Configuration)
	assert.Equal(t, ear.UnrecognizedRuntimeClaim, realmAppraisal.TrustVector.Executables)
}

// Helper: extract claims from evidence token once.
func extractClaimsFromEvidence(t *testing.T) (map[string]any, error) {
	t.Helper()
	impl := NewImplementation()
	evidence := &appraisal.Evidence{
		Data:  loadEvidenceToken(t),
		Nonce: testNonce,
	}
	return impl.ExtractClaims(evidence, nil)
}
