// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package riot

import (
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/dice"
	"github.com/veraison/ear"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/vts/appraisal"
	"go.uber.org/zap"
)

var altNameID = asn1.ObjectIdentifier{2, 5, 29, 17}

var Descriptor = handler.SchemeDescriptor{
	Name: "RIOT",
	VersionMajor: 1,
	VersionMinor: 0,
	CorimProfiles: []string{
		ProfileString,
	},
	EvidenceMediaTypes: []string{
		"application/pem-certificate-chain",
	},
}

type Implementation struct{
	logger *zap.SugaredLogger
}

func NewImplementation() *Implementation {
	return &Implementation{
		logger: log.Named(Descriptor.Name),
	}
}

func (o *Implementation) GetTrustAnchorIDs(
	evidence *appraisal.Evidence,
) ([]*comid.Environment, error) {
	vendor := "Veraison Project"
	model := "RIOT"

	return []*comid.Environment{
		{
			Class: &comid.Class{
				Vendor: &vendor,
				Model: &model,
			},
		},
	}, nil
}

func (o *Implementation) ExtractClaims(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
) (map[string]any, error) {
	roots := x509.NewCertPool()
	intermediates := x509.NewCertPool()

	if err := parseTrustAnchors(trustAnchors, roots, intermediates); err != nil {
		return nil, err
	}

	aliasCert, err := parseEvidenceCerts(evidence.Data, intermediates, roots)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	opts := x509.VerifyOptions{
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		Roots:         roots,
		Intermediates: intermediates,
	}

	claims, err := extractEvidenceClaims(aliasCert)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	// note: must verify this after extracting claims so that the Subject Alternative Name
	// gets processed; otherwise, it will be raised an unhandled critical extension.
	if _, err = aliasCert.Verify(opts); err != nil {
		return nil, handler.BadEvidence(
			"failed to verify alias cert: " + err.Error())
	}

	return claims, nil
}

func (o *Implementation) ValidateEvidenceIntegrity(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
	endorsements []*comid.ValueTriple,
) error {
	// Cert verified earlier when extracting claims -- see note inside ExtractClaims above.
	return nil
}

func (o *Implementation) AppraiseClaims(
	claims map[string]any,
	endorsements []*comid.ValueTriple,
) (*ear.AttestationResult, error) {
	result := handler.CreateAttestationResult(Descriptor.Name)
	appraisal := result.Submods[Descriptor.Name]

	// If we got this far, this means the cert chain has been verfied, and
	// thus, the identity has been established as valid.
	appraisal.TrustVector.InstanceIdentity = ear.TrustworthyInstanceClaim

	appraisal.UpdateStatusFromTrustVector()
	appraisal.VeraisonPolicyClaims = &claims

	return result, nil
}

func extractEvidenceClaims(cert *x509.Certificate) (map[string]any, error) {
	claims := make(map[string]any)

	for _, ext := range cert.Extensions {
		if ext.Id.Equal(altNameID) {
			if err := processAltName(ext.Value, &claims); err != nil {
				return nil, err
			}
			break
		}
	}

	// Remove Subject Alternative Name from Unhandled critical extensions list, as
	// we've now "handled" it. This will allow the cert to be verified.
	altNameIdx := -1
	for i, extOID := range cert.UnhandledCriticalExtensions {
		if extOID.Equal(altNameID) {
			altNameIdx = i
			break
		}
	}

	if altNameIdx != -1 {
		cert.UnhandledCriticalExtensions = append(cert.UnhandledCriticalExtensions[:altNameIdx],
			cert.UnhandledCriticalExtensions[altNameIdx+1:]...)
	}

	return claims, nil
}

func processAltName(data []byte, claims *map[string]any) error {

	var dice dice.DiceExtension

	rest, err := dice.UnmarshalDER(data)
	if err != nil {
		return err
	}
	if len(rest) != 0 {
		return errors.New("trailing data after DICE extension")
	}

	(*claims)["FWID"] = dice.CompositeDeviceID.Fwid.Fwid
	(*claims)["DeviceID"] = dice.CompositeDeviceID.DeviceID.SubjectPublicKey.Bytes

	return nil
}

func parseEvidenceCerts(
	evidence []byte,
	intermediates *x509.CertPool,
	roots *x509.CertPool,
) (*x509.Certificate, error) {
	block, rest := pem.Decode(evidence)
	if block == nil {
		return nil, errors.New("problem extracting token cert PEM block")
	}

	aliasCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	block, rest = pem.Decode(rest)
	if block == nil {
		return nil, errors.New("problem extracting token cert PEM block")
	}

	deviceCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	// self signed cert should not have any intermediates presented with it.
	if deviceCert.Subject.String() == deviceCert.Issuer.String() {
		if len(rest) != 0 {
			return nil, errors.New("additional data found alongside a self-signed Cert")
		}

		roots.AddCert(deviceCert)

		return aliasCert, nil
	}

	// Device cert is not self-signed. Add it as an intermediate and process
	// the rest of the certs if any.

	intermediates.AddCert(deviceCert)

	for len(rest) != 0 {
		block, rest = pem.Decode(rest)
		if block == nil {
			return nil, errors.New("problem extracting token intermediate PEM block")
		}

		intCert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}

		intermediates.AddCert(intCert)
	}

	return aliasCert, nil
}

func parseTrustAnchors(
	trustAnchors []*comid.KeyTriple,
	roots *x509.CertPool,
	intermediates *x509.CertPool,
) error {
	for _, keyTriple := range trustAnchors {
		for _, verifKey := range keyTriple.VerifKeys {
			switch verifKey.Type() {
			case comid.PKIXBase64CertType, comid.PKIXBase64CertPathType:
				var block *pem.Block
				rest := []byte(verifKey.Value.String())
				for len(rest) != 0 {
					block, rest = pem.Decode(rest)
					if block == nil {
						return errors.New("problem extracting trust anchor PEM block")
					}

					cert, err := x509.ParseCertificate(block.Bytes)
					if err != nil {
						return err
					}

					if cert.Subject.String() == cert.Issuer.String() {
						// self-signed
						roots.AddCert(cert)
					} else {
						intermediates.AddCert(cert)
					}
				}
			default:
				return fmt.Errorf("unexpected verif. key type: %s", verifKey.Type())
			}
		}
	}

	return nil
}
