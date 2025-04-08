// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package sevsnp

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/proto/sevsnp"
	"github.com/google/go-sev-guest/verify"
	"github.com/google/go-sev-guest/verify/trust"
	sevsnpParser "github.com/jraman567/go-gen-ref/cmd/sevsnp"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/ear"
	"github.com/veraison/ratsd/tokens"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/proto"
)

// EvidenceHandler implements the IEvidenceHandler interface for SEVSNP
type EvidenceHandler struct {
}

// GetName returns the name of this evidence handler instance
func (o EvidenceHandler) GetName() string {
	return "sevsnp-evidence-handler"
}

// GetAttestationScheme returns the attestation scheme
func (o EvidenceHandler) GetAttestationScheme() string {
	return SchemeName
}

// GetSupportedMediaTypes returns the supported media types for the SEVSNP scheme
func (o EvidenceHandler) GetSupportedMediaTypes() []string {
	return EvidenceMediaTypes
}

// ExtractClaims converts evidence in tsm-report format to our
// "internal representation", which is in CoRIM format.
func (o EvidenceHandler) ExtractClaims(
	token *proto.AttestationToken,
	_ []string,
) (map[string]interface{}, error) {
	var claimsSet map[string]interface{}
	var tsm tokens.TSMReport

	err := tsm.FromCBOR(token.Data)
	if err != nil {
		return nil, err
	}

	reportProto, err := abi.ReportToProto(tsm.OutBlob)
	if err != nil {
		return nil, err
	}

	refValComid, err := sevsnpParser.ReportToComid(reportProto, 0)
	if err != nil {
		return nil, err
	}

	err = refValComid.Valid()
	if err != nil {
		return nil, err
	}

	refValCorim := corim.UnsignedCorim{}
	refValCorim.SetProfile("http://amd.com/2024/snp-corim-profile")
	refValCorim.AddComid(refValComid)

	refValJson, err := refValCorim.ToJSON()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(refValJson, &claimsSet)
	if err != nil {
		return nil, err
	}

	return claimsSet, nil
}

// snpAttestationOptions parameter for verifying certificate chain
func snpAttestationOptions() *verify.Options {
	return &verify.Options{
		Getter:              trust.DefaultHTTPSGetter(),
		Now:                 time.Now(),
		DisableCertFetching: true,
	}
}

// readCert helper function to read a certificate from a blob
func readCert(cert []byte) ([]byte, error) {
	block, _ := pem.Decode(cert)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("failed to read certificate")
	}
	return block.Bytes, nil
}

// compareTAs compares two X509 certificates for equality
func compareTAs(provisionedArk []byte, evidenceArk []byte) (bool, error) {
	pArk, err := readCert(provisionedArk)
	if err != nil {
		return false, err
	}

	pCert, err := x509.ParseCertificate(pArk)
	if err != nil {
		return false, err
	}

	eArk, err := readCert(evidenceArk)
	if err != nil {
		return false, err
	}

	eCert, err := x509.ParseCertificate(eArk)
	if err != nil {
		return false, err
	}

	return pCert.Equal(eCert), nil
}

// ValidateEvidenceIntegrity verifies that the ARK in the
// evidence matches the provisioned ARK, confirms the
// integrity of the certificate chain, and validates
// the signature of the evidence.
//
// The "auxblob" in the evidence contains a certificate chain.
// The Trust Anchor in this chain is AMD Root Key (ARK).
func (o EvidenceHandler) ValidateEvidenceIntegrity(
	token *proto.AttestationToken,
	trustAnchors []string,
	endorsementsStrings []string,
) error {
	var (
		taEndorsement handler.Endorsement
		avk           comid.KeyTriple
		tsm           tokens.TSMReport
	)

	// Get the ARK TA
	for i, t := range trustAnchors {
		var endorsement handler.Endorsement

		if err := json.Unmarshal([]byte(t), &endorsement); err != nil {
			return fmt.Errorf("could not decode endorsement at index %d: %w", i, err)
		}

		if endorsement.Type == handler.EndorsementType_VERIFICATION_KEY {
			taEndorsement = endorsement
			break
		}
	}

	if taEndorsement.Type != handler.EndorsementType_VERIFICATION_KEY {
		return fmt.Errorf("trust anchors unavailable")
	}

	err := json.Unmarshal(taEndorsement.Attributes, &avk)
	if err != nil {
		return err
	}

	provisionedArk := avk.VerifKeys[0]

	// Parse certificate chain in evidence (auxblob)
	err = tsm.FromCBOR(token.Data)
	if err != nil {
		return err
	}

	protoReport, err := abi.ReportToProto(tsm.OutBlob)
	if err != nil {
		return err
	}

	ark, err := getARK(tsm.AuxBlob)
	if err != nil {
		return err
	}

	arkBlock, err := readCert(ark)
	if err != nil {
		return err
	}

	ask, err := getASK(tsm.AuxBlob)
	if err != nil {
		return err
	}

	askBlock, err := readCert(ask)
	if err != nil {
		return err
	}

	vcek, err := getVCEK(tsm.AuxBlob)
	if err != nil {
		return err
	}

	vcekBlock, err := readCert(vcek)
	if err != nil {
		return err
	}

	// Test if TA matches with the one supplied in evidence
	match, err := compareTAs([]byte(provisionedArk.String()), ark)
	if err != nil {
		return err
	}
	if !match {
		return fmt.Errorf("ARK in evidence does not match provisioned ARK")
	}

	// Validate the integrity of evidence by ensuring the
	// certificate chain is intact, and the signature is valid
	var attestation sevsnp.Attestation
	attestation.Report = protoReport
	attestation.CertificateChain = &sevsnp.CertificateChain{VcekCert: vcekBlock, AskCert: askBlock, ArkCert: arkBlock}
	err = verify.SnpAttestation(&attestation, snpAttestationOptions())

	if err != nil {
		log.Errorf("failed to validate certificate chain: %+v\n", err)
	}

	return err
}

// refvalToComidTriple converts extracted reference values to CoMID value triple
func refvalToComidTriple(endorsementsStrings []string) (*comid.ValueTriple, error) {
	var (
		refValEndorsement handler.Endorsement
		rv                comid.ValueTriple
	)

	for i, e := range endorsementsStrings {
		var endorsement handler.Endorsement

		if err := json.Unmarshal([]byte(e), &endorsement); err != nil {
			return nil, fmt.Errorf("could not decode endorsement at index %d: %w", i, err)
		}

		if endorsement.Type == handler.EndorsementType_REFERENCE_VALUE {
			refValEndorsement = endorsement
			break
		}
	}

	if refValEndorsement.Type != handler.EndorsementType_REFERENCE_VALUE {
		return nil, fmt.Errorf("reference values unavailable")
	}

	err := json.Unmarshal(refValEndorsement.Attributes, &rv)
	if err != nil {
		return nil, err
	}

	return &rv, nil
}

// evidenceToComidTriple converts claim set to CoMID value triple
func evidenceToComidTriple(ec *proto.EvidenceContext) (*comid.ValueTriple, error) {
	evCorimJson, err := json.Marshal(ec.Evidence.AsMap())
	if err != nil {
		return nil, err
	}

	evComid, err := comidFromJson(evCorimJson)
	if err != nil {
		return nil, err
	}

	return &evComid.Triples.ReferenceValues.Values[0], nil
}

// compareMeasurements checks if two given comid.Measurement variables are the same.
func compareMeasurements(refM comid.Measurement, evM comid.Measurement) bool {
	// RawValue comparison
	if refM.Val.RawValue != nil {
		if evM.Val.RawValue == nil {
			return false
		}

		refDigest, _ := refM.Val.RawValue.GetBytes()
		return evM.Val.RawValue.CompareAgainstReference(refDigest, nil)
	}

	// Digests comparison
	if refM.Val.Digests != nil {
		if evM.Val.Digests == nil {
			return false
		}

		return evM.Val.Digests.CompareAgainstReference(*refM.Val.Digests)
	}

	// SVN comparison
	if refM.Val.SVN != nil {
		if evM.Val.SVN == nil {
			log.Debugf("evidence doesn't have SVN")
			return false
		}

		if c, ok := evM.Val.SVN.Value.(*comid.TaggedSVN); ok {
			if r, ok := refM.Val.SVN.Value.(*comid.TaggedSVN); ok {
				return c.CompareAgainstRefSVN(*r)
			} else if r, ok := refM.Val.SVN.Value.(*comid.TaggedMinSVN); ok {
				return c.CompareAgainstRefMinSVN(*r)
			} else {
				log.Debugf("unknown refVal SVN type")
				return false
			}
		} else if c, ok := evM.Val.SVN.Value.(*comid.TaggedMinSVN); ok {
			if r, ok := refM.Val.SVN.Value.(*comid.TaggedMinSVN); ok {
				return c.Equal(*r)
			}
			log.Debugf("can't compare TaggedMinSVN against TaggedSVN")
			return false
		} else {
			log.Debugf("unknown evidence SVN type")
			return false
		}
	}

	return true
}

// AppraiseEvidence confirms if the claims in the evidence match with the provisioned
// reference values.
//
// Appraisal can confirm if the evidence is genuinely generated by AMD
// hardware and if SEV-SNP enables memory encryption. As such, set the
// "Hardware" and "RuntimeOpaque" values in the trustworthiness vector;
// we can't infer other aspects of the vector from SEV-SNP evidence alone.
func (o EvidenceHandler) AppraiseEvidence(
	ec *proto.EvidenceContext,
	endorsementsStrings []string,
) (*ear.AttestationResult, error) {
	var (
		err         error
		evidenceMap map[string]interface{}
	)

	refVal, err := refvalToComidTriple(endorsementsStrings)
	if err != nil {
		return nil, err
	}

	evidence, err := evidenceToComidTriple(ec)
	if err != nil {
		return nil, err
	}

	result := handler.CreateAttestationResult(SchemeName)

	appraisal := result.Submods[SchemeName]

	appraisal.TrustVector.InstanceIdentity = ear.NoClaim
	appraisal.TrustVector.Executables = ear.NoClaim
	appraisal.TrustVector.Configuration = ear.NoClaim
	appraisal.TrustVector.FileSystem = ear.NoClaim
	appraisal.TrustVector.StorageOpaque = ear.NoClaim
	appraisal.TrustVector.SourcedData = ear.NoClaim
	appraisal.TrustVector.Hardware = ear.UnsafeHardwareClaim
	appraisal.TrustVector.RuntimeOpaque = ear.VisibleMemoryRuntimeClaim

	for _, m := range refVal.Measurements.Values {
		var (
			k  uint64
			em *comid.Measurement
		)

		k, err = m.Key.GetKeyUint()
		if err != nil {
			break
		}

		// REPORT_ID is ephemeral, so we can't use it for verification.
		// REPORT_DATA is client-supplied , which we aren't using for
		// verification in this scheme.
		if k == mKeyReportData || k == mKeyReportID {
			continue
		}

		em, err = measurementByUintKey(*evidence, k)
		if err != nil {
			break
		}

		if em == nil {
			err = fmt.Errorf("MKey %d not found in Evidence", k)
			break
		}

		if !compareMeasurements(m, *em) {
			err = fmt.Errorf("MKey %d in reference value doesn't match with evidence", k)
			break
		}
	}

	if err == nil {
		appraisal.TrustVector.Hardware = ear.GenuineHardwareClaim
		appraisal.TrustVector.RuntimeOpaque = ear.EncryptedMemoryRuntimeClaim
	}

	appraisal.UpdateStatusFromTrustVector()

	evidenceJson, err := json.Marshal(evidence)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(evidenceJson, &evidenceMap)
	if err != nil {
		return nil, err
	}

	appraisal.VeraisonAnnotatedEvidence = &evidenceMap

	return result, err
}
