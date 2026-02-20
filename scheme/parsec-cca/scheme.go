// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_cca

import (
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/ear"
	"github.com/veraison/go-cose"
	parsec_cca "github.com/veraison/parsec/cca"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	cca_scheme "github.com/veraison/services/scheme/arm-cca"
	"github.com/veraison/services/scheme/common"
	"github.com/veraison/services/vts/appraisal"
	"go.uber.org/zap"
)

var Descriptor = handler.SchemeDescriptor{
	Name:         "PARSEC_CCA",
	VersionMajor: 1,
	VersionMinor: 0,
	CorimProfiles: []string{
		ProfileString,
	},
	EvidenceMediaTypes: []string{
		"application/vnd.parallaxsecond.key-attestation.cca",
	},
}

type Implementation struct {
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
	var parsecEvidence parsec_cca.Evidence

	err := parsecEvidence.FromCBOR(evidence.Data)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	implIDbytes, err := parsecEvidence.Pat.PlatformClaims.GetImplID()
	if err != nil {
		return nil, err
	}

	instIDbytes, err := parsecEvidence.Pat.PlatformClaims.GetInstID()
	if err != nil {
		return nil, err
	}

	classID, err := comid.NewImplIDClassID(implIDbytes)
	if err != nil {
		return nil, err
	}

	instanceID, err := comid.NewUEIDInstance(instIDbytes)
	if err != nil {
		return nil, err
	}

	return []*comid.Environment{
		{
			Class:    &comid.Class{ClassID: classID},
			Instance: instanceID,
		},
	}, nil
}

func (o *Implementation) GetReferenceValueIDs(
	trustAnchors []*comid.KeyTriple,
	claims map[string]any,
) ([]*comid.Environment, error) {
	numTAs := len(trustAnchors)
	if numTAs != 1 {
		return nil, fmt.Errorf("expected exactly 1 trust anchor; got %d", numTAs)
	}

	return []*comid.Environment{
		{
			Class: trustAnchors[0].Environment.Class,
		},
	}, nil
}

func (o *Implementation) ValidateComid(c *comid.Comid) error {
	return nil
}

func (o *Implementation) ExtractClaims(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
) (map[string]any, error) {
	var parsecEvidence parsec_cca.Evidence

	err := parsecEvidence.FromCBOR(evidence.Data)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	ck, err := parsecEvidence.Kat.Cnf.COSEKey.MarshalCBOR()
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	katClaims := map[string]any{
		"nonce": parsecEvidence.Kat.Nonce.GetI(0),
		"akpub": ck,
	}

	platformClaims, err := common.ToMapViaJSON(parsecEvidence.Pat.PlatformClaims)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	realmClaims, err := common.ToMapViaJSON(parsecEvidence.Pat.RealmClaims)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	claims := map[string]any{
		"kat":          katClaims,
		"cca.platform": platformClaims,
		"cca.realm":    realmClaims,
	}

	return claims, nil
}

func (o *Implementation) ValidateEvidenceIntegrity(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
	endorsements []*comid.ValueTriple,
) error {
	var parsecEvidence parsec_cca.Evidence

	err := parsecEvidence.FromCBOR(evidence.Data)
	if err != nil {
		return handler.BadEvidence(err)
	}

	pk, err := common.ExtractPublicKeyFromTrustAnchors(trustAnchors)
	if err != nil {
		return fmt.Errorf("could not get public key from trust anchors: %w", err)
	}

	if err = parsecEvidence.Verify(pk); err != nil {
		return handler.BadEvidence("failed to verify signature: %w", err)
	}

	o.logger.Debug("Parsec CCA token signature verified")
	return nil
}

func (o *Implementation) AppraiseClaims(
	claims map[string]any,
	endorsements []*comid.ValueTriple,
) (*ear.AttestationResult, error) {
	result := handler.CreateAttestationResult(Descriptor.Name)
	appraisal := result.Submods[Descriptor.Name]

	// once the signature on the token is verified, we can claim the HW is
	// authentic
	appraisal.TrustVector.Hardware = ear.GenuineHardwareClaim

	katClaims := claims["kat"].(map[string]any)

	var coseKey cose.Key
	if err := coseKey.UnmarshalCBOR(katClaims["akpub"].([]byte)); err != nil {
		return result, handler.BadEvidence(err)
	}

	pk, err := coseKey.PublicKey()
	if err != nil {
		return result, handler.BadEvidence(err)
	}

	if err := appraisal.SetKeyAttestation(pk); err != nil {
		return result, fmt.Errorf("setting extracted public key: %w", err)
	}

	err = cca_scheme.AppraisePlatform(o.logger, appraisal, claims["cca.platform"], endorsements)
	if err != nil {
		return result, err
	}

	return result, nil
}
