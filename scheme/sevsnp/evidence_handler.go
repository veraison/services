// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package sevsnp

import (
	"bytes"
	"encoding/json"
	"errors"
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

var (
	ErrNoARK           = errors.New("missing ARK certificate in evidence")
	ErrNoASK           = errors.New("missing ASK certificate in evidence")
	ErrNoVEK           = errors.New("evidence must supply VLEK or VCEK")
	ErrNoVCEK          = errors.New("VCEK is missing")
	ErrNoVLEK          = errors.New("VLEK is missing")
	ErrTAMismatch      = errors.New("evidence Trust Anchor (ARK) doesn't match the provisioned one")
	ErrNoProvisionedTA = errors.New("missing provisioned Trust Anchor")
	ErrNoProvisionedRV = errors.New("unrecognized evidence")
	ErrBadSigningKey   = errors.New("bad signing key in attestation report")
)

const (
	ReportSigningKeyVcek = 0
	ReportSigningKeyVlek = 1
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

func transformEvidenceToCorim(token *proto.AttestationToken) (*corim.UnsignedCorim, error) {
	tsm, err := parseAttestationToken(token)
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
	refValCorim.SetProfile(EndorsementMediaTypeRV)
	refValCorim.AddComid(refValComid)

	return &refValCorim, nil
}

// ExtractClaims converts evidence in tsm-report format to our
// "internal representation", which is in CoRIM format.
func (o EvidenceHandler) ExtractClaims(
	token *proto.AttestationToken,
	_ []string,
) (map[string]interface{}, error) {
	var claimsSet map[string]interface{}

	refValCorim, err := transformEvidenceToCorim(token)
	if err != nil {
		return nil, err
	}

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

func extractProvisionedTA(trustAnchors []string) (*comid.CryptoKey, error) {
	var (
		taEndorsement *handler.Endorsement
		avk           comid.KeyTriple
	)

	for i, t := range trustAnchors {
		var endorsement handler.Endorsement

		if err := json.Unmarshal([]byte(t), &endorsement); err != nil {
			return nil, fmt.Errorf("could not decode endorsement at index %d: %w", i, err)
		}

		if endorsement.Type == handler.EndorsementType_VERIFICATION_KEY {
			taEndorsement = &endorsement
			break
		}
	}

	if taEndorsement == nil {
		return nil, ErrNoProvisionedTA
	}

	err := json.Unmarshal(taEndorsement.Attributes, &avk)
	if err != nil {
		return nil, err
	}

	// The StoreHandler takes care of ensuring that only one TA is
	// supplied, we don't have to re-check it here.
	provisionedArk := avk.VerifKeys[0]

	return provisionedArk, nil
}

func validateCertificateChain(certChain *sevsnp.CertificateChain) error {
	if len(certChain.GetArkCert()) == 0 {
		return ErrNoARK
	}

	if len(certChain.GetAskCert()) == 0 {
		return ErrNoASK
	}

	if len(certChain.GetVcekCert()) == 0 && len(certChain.GetVlekCert()) == 0 {
		return ErrNoVEK
	}

	return nil
}

func validateTA(certChain *sevsnp.CertificateChain, provisionedArk *comid.CryptoKey) error {
	if !bytes.Equal(certChain.GetArkCert(), []byte(provisionedArk.String())) {
		return ErrTAMismatch
	}

	return nil
}

func validateReportIntegrity(tsm *tokens.TSMReport, certChain *sevsnp.CertificateChain) error {
	var (
		ark, ask, vcek, vlek []byte
		attestation          sevsnp.Attestation
	)

	options := verify.Options{
		Getter:              trust.DefaultHTTPSGetter(),
		Now:                 time.Now(),
		DisableCertFetching: true,
	}

	protoReport, err := abi.ReportToProto(tsm.OutBlob)
	if err != nil {
		return err
	}
	attestation.Report = protoReport

	if ark, err = readCert(certChain.GetArkCert()); err != nil {
		return fmt.Errorf("can't read ARK to validate cert chain: %w", err)
	}

	if ask, err = readCert(certChain.GetAskCert()); err != nil {
		return fmt.Errorf("can't read ASK to validate cert chain: %w", err)
	}

	signerInfo, err := abi.ParseSignerInfo(protoReport.GetSignerInfo())
	if err != nil {
		return err
	}

	switch signerInfo.SigningKey {
	case ReportSigningKeyVlek:
		if len(certChain.GetVlekCert()) == 0 {
			return ErrNoVLEK
		}
		if vlek, err = readCert(certChain.GetVlekCert()); err != nil {
			return fmt.Errorf("can't read VLEK to validate cert chain: %w", err)
		}
		attestation.CertificateChain = &sevsnp.CertificateChain{VlekCert: vlek, AskCert: ask, ArkCert: ark}
	case ReportSigningKeyVcek:
		if len(certChain.GetVcekCert()) == 0 {
			return ErrNoVCEK
		}
		if vcek, err = readCert(certChain.GetVcekCert()); err != nil {
			return fmt.Errorf("can't read VCEK to validate cert chain: %w", err)
		}
		attestation.CertificateChain = &sevsnp.CertificateChain{VcekCert: vcek, AskCert: ask, ArkCert: ark}
	default:
		return ErrBadSigningKey
	}

	return verify.SnpAttestation(&attestation, &options)
}

// ValidateEvidenceIntegrity confirms the integrity of evidence by doing the following:
//   - verifies that the TA in the evidence matches the provisioned TA
//   - confirms the integrity of the certificate chain
//   - validates the integrity of evidence by checking its signature
func (o EvidenceHandler) ValidateEvidenceIntegrity(
	token *proto.AttestationToken,
	trustAnchors []string,
	_ []string,
) error {
	var (
		tsm            *tokens.TSMReport
		provisionedArk *comid.CryptoKey
		certChain      *sevsnp.CertificateChain
		err            error
	)

	if tsm, err = parseAttestationToken(token); err != nil {
		return err
	}

	if provisionedArk, err = extractProvisionedTA(trustAnchors); err != nil {
		return err
	}

	if certChain, err = parseCertificateChainFromEvidence(tsm); err != nil {
		return err
	}

	if err := validateCertificateChain(certChain); err != nil {
		return err
	}

	if err := validateTA(certChain, provisionedArk); err != nil {
		return err
	}

	return validateReportIntegrity(tsm, certChain)
}

// refvalToComidTriple converts extracted reference values to CoMID value triple
func refvalToComidTriple(endorsementsStrings []string) (*comid.ValueTriple, error) {
	var (
		refValEndorsement *handler.Endorsement
		rv                comid.ValueTriple
	)

	for i, e := range endorsementsStrings {
		var endorsement handler.Endorsement

		if err := json.Unmarshal([]byte(e), &endorsement); err != nil {
			return nil, fmt.Errorf("could not decode endorsement at index %d: %w", i, err)
		}

		if endorsement.Type == handler.EndorsementType_REFERENCE_VALUE {
			refValEndorsement = &endorsement
			break
		}
	}

	if refValEndorsement == nil {
		return nil, ErrNoProvisionedRV
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

func compareTcb(refM comid.Measurement, evM comid.Measurement) bool {
	if refM.Val.SVN == nil {
		log.Errorf("reference doesn't have SVN")
		return false
	}

	if evM.Val.SVN == nil {
		log.Errorf("evidence doesn't have SVN")
		return false
	}

	refTcbParts, err := transformSVNtoTCB(*refM.Val.SVN)
	if err != nil {
		log.Errorf("could not transform reference SVN to TCB parts: %v", err)
		return false
	}

	evTcbParts, err := transformSVNtoTCB(*evM.Val.SVN)
	if err != nil {
		log.Errorf("could not transform evidence SVN to TCB parts: %v", err)
	}

	if evTcbParts.BlSpl < refTcbParts.BlSpl ||
		evTcbParts.SnpSpl < refTcbParts.SnpSpl ||
		evTcbParts.TeeSpl < refTcbParts.TeeSpl ||
		evTcbParts.UcodeSpl < refTcbParts.UcodeSpl {
		return false
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

claimsLoop:
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

		switch k {
		case mKeyReportedTcb:
			if !compareTcb(m, *em) {
				err = fmt.Errorf("reported TCB in evidence doesn't match reference")
				break claimsLoop
			}
		default:
			if !compareMeasurements(m, *em) {
				err = fmt.Errorf("MKey %d in reference value doesn't match with evidence", k)
				break claimsLoop
			}
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
