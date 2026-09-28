// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sevsnp

import (
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"testing"

	"github.com/google/go-sev-guest/verify/trust"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/ear"
	"github.com/veraison/ratsd/tokens"
	"github.com/veraison/services/vts/appraisal"
)

//go:embed test/evidence/sevsnp-ratsd-token
var sevsnpRatsdToken []byte

//go:embed test/evidence/sevsnp-ratsd-claims.json
var sevsnpRatsdClaimsMapJson []byte

//go:embed test/evidence/sevsnp-ratsd-env.json
var sevsnpRatsdEnv []byte

//go:embed test/corim/corim-sevsnp-valid.cbor
var sevsnpCorimValidEndorsements []byte

func Test_GetTrustAnchorIDs_ok(t *testing.T) {
	Impl := NewImplementation()
	evidence := appraisal.Evidence{Data: sevsnpRatsdToken, MediaType: `application/eat+cwt; eat_profile="tag:github.com,2025:veraison/ratsd/cmw"`}
	env, err := Impl.GetTrustAnchorIDs(&evidence)
	require.NoError(t, err)

	expectedVendor := "Advanced Micro Devices"
	expectedModel := "ARK-Genoa"
	expectedEnv := comid.Environment{Class: &comid.Class{Vendor: &expectedVendor, Model: &expectedModel}}

	require.Len(t, env, 1)
	assert.Equal(t, *env[0], expectedEnv)
}

func Test_GetReferenceValueIDs_ok(t *testing.T) {
	var (
		claims      map[string]any
		expectedEnv comid.Environment
	)

	err := json.Unmarshal(sevsnpRatsdClaimsMapJson, &claims)
	require.NoError(t, err)

	Impl := NewImplementation()
	env, err := Impl.GetReferenceValueIDs(nil, claims)
	require.NoError(t, err)
	require.Len(t, env, 1)

	err = json.Unmarshal(sevsnpRatsdEnv, &expectedEnv)
	require.NoError(t, err)

	assert.Equal(t, *env[0], expectedEnv)
}

func Test_ExtractClaims_ok(t *testing.T) {
	var expectedClaims map[string]any

	err := json.Unmarshal(sevsnpRatsdClaimsMapJson, &expectedClaims)
	require.NoError(t, err)

	expectedMeasurementMap, err := transformClaimsToMeasurementsMap(expectedClaims)
	require.NoError(t, err)

	Impl := NewImplementation()
	evidence := appraisal.Evidence{Data: sevsnpRatsdToken, MediaType: `application/eat+cwt; eat_profile="tag:github.com,2025:veraison/ratsd/cmw"`}
	claims, err := Impl.ExtractClaims(&evidence, nil)
	require.NoError(t, err)
	measurementMap, err := transformClaimsToMeasurementsMap(claims)
	require.NoError(t, err)

	assert.Equal(t, expectedMeasurementMap, measurementMap)
}

func Test_ValidateEvidenceIntegrity_ok(t *testing.T) {
	var (
		endCorim     corim.UnsignedCorim
		endComid     comid.Comid
		trustAnchors []*comid.KeyTriple //nolint:prealloc
	)

	err := endCorim.FromCBOR(sevsnpCorimValidEndorsements)
	require.NoError(t, err)
	require.Len(t, endCorim.Tags, 2)
	err = endComid.FromCBOR(endCorim.Tags[1].Content)
	require.NoError(t, err)
	keyTriples := *endComid.Triples.AttestVerifKeys
	trustAnchors = append(trustAnchors, &keyTriples[0])

	nonce, err := hex.DecodeString("4d4944424e48323869696f69736a5079787878787878787878787878787878784d4944424e48323869696f69736a507978787878787878787878787878787878")
	require.NoError(t, err)

	incorrectNonce := make([]byte, len(nonce))
	copy(incorrectNonce, nonce)
	incorrectNonce[0] ^= 0xff

	Impl := NewImplementation()

	testCases := []struct {
		name    string
		nonce   []byte
		wantErr bool
	}{
		{name: "valid nonce", nonce: nonce},
		{name: "invalid nonce", nonce: incorrectNonce, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			evidence := appraisal.Evidence{
				Data:      sevsnpRatsdToken,
				MediaType: `application/eat+cwt; eat_profile="tag:github.com,2025:veraison/ratsd/cmw"`,
				Nonce:     tc.nonce,
			}

			err := Impl.ValidateEvidenceIntegrity(&evidence, trustAnchors, nil)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func Test_AppraiseClaims_ok(t *testing.T) {
	var (
		claims       map[string]any
		endCorim     corim.UnsignedCorim
		endComid     comid.Comid
		endorsements []*comid.ValueTriple //nolint:prealloc
	)

	err := json.Unmarshal(sevsnpRatsdClaimsMapJson, &claims)
	require.NoError(t, err)

	err = endCorim.FromCBOR(sevsnpCorimValidEndorsements)
	require.NoError(t, err)
	require.Len(t, endCorim.Tags, 2)
	err = endComid.FromCBOR(endCorim.Tags[0].Content)
	require.NoError(t, err)

	Impl := NewImplementation()
	endorsements = append(endorsements, &endComid.Triples.ReferenceValues.Values[0])
	attestationResult, err := Impl.AppraiseClaims(claims, endorsements)
	require.NoError(t, err)

	sevsnpSubmod := attestationResult.Submods["SEVSNP"]

	assert.Equal(t, ear.TrustTierAffirming, *sevsnpSubmod.Status)
	assert.Equal(t, ear.GenuineHardwareClaim, sevsnpSubmod.TrustVector.Hardware)
	assert.Equal(t, ear.EncryptedMemoryRuntimeClaim, sevsnpSubmod.TrustVector.RuntimeOpaque)
}

//go:embed test/evidence/sevsnp-tsm-report.cbor
var sevsnpTSMReportCBOR []byte

// derCertTable rewrites a certificate table whose entries are PEM into the
// same table with DER entries, as Google Cloud hosts supply it.
func derCertTable(t *testing.T, table []byte) []byte {
	t.Helper()

	type entry struct {
		guid []byte
		der  []byte
	}

	var entries []entry
	for i := 0; i+24 <= len(table); i += 24 {
		guid := table[i : i+16]
		if string(guid) == string(make([]byte, 16)) {
			break
		}
		off := binary.LittleEndian.Uint32(table[i+16:])
		n := binary.LittleEndian.Uint32(table[i+20:])
		block, _ := pem.Decode(table[off : off+n])
		require.NotNil(t, block)
		entries = append(entries, entry{guid: guid, der: block.Bytes})
	}
	require.NotEmpty(t, entries)

	off := uint32(24 * (len(entries) + 1))
	var head, body []byte
	for _, e := range entries {
		head = append(head, e.guid...)
		head = binary.LittleEndian.AppendUint32(head, off+uint32(len(body)))
		head = binary.LittleEndian.AppendUint32(head, uint32(len(e.der)))
		body = append(body, e.der...)
	}
	head = append(head, make([]byte, 24)...)

	return append(head, body...)
}

func Test_ValidateEvidenceIntegrity_DERCertTable(t *testing.T) {
	var (
		endCorim corim.UnsignedCorim
		endComid comid.Comid
		tsm      tokens.TSMReport
	)

	require.NoError(t, endCorim.FromCBOR(sevsnpCorimValidEndorsements))
	require.NoError(t, endComid.FromCBOR(endCorim.Tags[1].Content))
	keyTriples := *endComid.Triples.AttestVerifKeys
	trustAnchors := []*comid.KeyTriple{&keyTriples[0]}

	nonce, err := hex.DecodeString("4d4944424e48323869696f69736a5079787878787878787878787878787878784d4944424e48323869696f69736a507978787878787878787878787878787878")
	require.NoError(t, err)

	require.NoError(t, tsm.FromCBOR(sevsnpTSMReportCBOR))
	pemToken, err := tsm.ToCBOR()
	require.NoError(t, err)
	tsm.AuxBlob = derCertTable(t, tsm.AuxBlob)
	derToken, err := tsm.ToCBOR()
	require.NoError(t, err)

	Impl := NewImplementation()

	for _, tc := range []struct {
		name    string
		token   []byte
		nonce   []byte
		wantErr string
	}{
		{name: "PEM certificate table", token: pemToken, nonce: nonce},
		{name: "DER certificate table", token: derToken, nonce: nonce},
		{
			name:    "DER certificate table, other nonce",
			token:   derToken,
			nonce:   append([]byte{nonce[0] ^ 0xff}, nonce[1:]...),
			wantErr: "nonce in the evidence doesn't match",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			evidence := appraisal.Evidence{
				Data:      tc.token,
				MediaType: EvidenceMediaTypeTSMCbor,
				Nonce:     tc.nonce,
			}
			err := Impl.ValidateEvidenceIntegrity(&evidence, trustAnchors, nil)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func Test_ValidateEvidenceIntegrity_DERCertTable_TAMismatch(t *testing.T) {
	var tsm tokens.TSMReport

	require.NoError(t, tsm.FromCBOR(sevsnpTSMReportCBOR))
	tsm.AuxBlob = derCertTable(t, tsm.AuxBlob)
	derToken, err := tsm.ToCBOR()
	require.NoError(t, err)

	// the ASK from the same table stands in for a trust anchor that is not the ARK
	certs, err := parseCertChainFromTSMReport(&tsm)
	require.NoError(t, err)
	askPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certs.GetAskCert()})
	ta, err := comid.NewPKIXBase64Cert(string(askPEM))
	require.NoError(t, err)
	trustAnchors := []*comid.KeyTriple{{VerifKeys: comid.CryptoKeys{ta}}}

	evidence := appraisal.Evidence{Data: derToken, MediaType: EvidenceMediaTypeTSMCbor}
	err = NewImplementation().ValidateEvidenceIntegrity(&evidence, trustAnchors, nil)
	require.ErrorIs(t, err, ErrTAMismatch)
}

// A report from a Google Cloud guest (n2d-standard-2, AMD Milan; 7 October 2026), signed with the chip's VCEK, with
// the certificate table as the host supplied it: DER entries (test/evidence/SOURCES.md).
//
//go:embed test/evidence/sevsnp-gcp-tsm-report.cbor
var sevsnpGCPTSMReportCBOR []byte

// the REPORT_DATA the guest bound into that report
const sevsnpGCPReportData = "f896ba422824bb5bb0201da4599c9124d24ab896cc3cec0a0c3a9ceac9a5208f4cbc7d3aa0f895ccc61cf7811c5b2cc62ea41c70771150e64c09cb7668539524"

// amdRootTrustAnchor is the ARK of a product line as go-sev-guest embeds it, in the form a CoRIM provisions it.
func amdRootTrustAnchor(t *testing.T, productLine string) []*comid.KeyTriple {
	t.Helper()
	roots := trust.DefaultRootCerts[productLine]
	require.NotNil(t, roots)
	arkPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: roots.ProductCerts.Ark.Raw})
	ta, err := comid.NewPKIXBase64Cert(string(arkPEM))
	require.NoError(t, err)
	return []*comid.KeyTriple{{VerifKeys: comid.CryptoKeys{ta}}}
}

// A real report whose host wrote the table as DER validates against AMD's ARK-Milan; another nonce and another
// product line's ARK are still refused for their own reasons.
func Test_ValidateEvidenceIntegrity_GoogleCloudReport(t *testing.T) {
	reportData, err := hex.DecodeString(sevsnpGCPReportData)
	require.NoError(t, err)

	for _, tc := range []struct {
		name         string
		nonce        []byte
		trustAnchors []*comid.KeyTriple
		wantErr      string
		wantErrIs    error
	}{
		{
			name:         "as supplied",
			nonce:        reportData,
			trustAnchors: amdRootTrustAnchor(t, "Milan"),
		},
		{
			name:         "other nonce",
			nonce:        append([]byte{reportData[0] ^ 0xff}, reportData[1:]...),
			trustAnchors: amdRootTrustAnchor(t, "Milan"),
			wantErr:      "nonce in the evidence doesn't match",
		},
		{
			name:         "another product line's ARK as the trust anchor",
			nonce:        reportData,
			trustAnchors: amdRootTrustAnchor(t, "Genoa"),
			wantErrIs:    ErrTAMismatch,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			evidence := appraisal.Evidence{
				Data:      sevsnpGCPTSMReportCBOR,
				MediaType: EvidenceMediaTypeTSMCbor,
				Nonce:     tc.nonce,
			}
			err := NewImplementation().ValidateEvidenceIntegrity(&evidence, tc.trustAnchors, nil)
			switch {
			case tc.wantErrIs != nil:
				require.ErrorIs(t, err, tc.wantErrIs)
			case tc.wantErr != "":
				require.ErrorContains(t, err, tc.wantErr)
			default:
				require.NoError(t, err)
			}
		})
	}
}
