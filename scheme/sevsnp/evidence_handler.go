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
	"github.com/veraison/ratsd/tokens"
	"github.com/veraison/services/handler"
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
	tsm, err := parseEvidence(token)
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

	if tsm, err = parseEvidence(token); err != nil {
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
