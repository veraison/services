// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package psa_iot

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/ear"
	"github.com/veraison/psatoken"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/scheme/common"
	"github.com/veraison/services/vts/appraisal"
	"go.uber.org/zap"
)

var Descriptor = handler.SchemeDescriptor{
	Name:         "PSA_IOT",
	VersionMajor: 1,
	VersionMinor: 0,
	CorimProfiles: []string{
		ProfileString,
	},
	EvidenceMediaTypes: []string{
		"application/psa-attestation-token",
		`application/eat-cwt; profile="http://arm.com/psa/2.0.0"`,
		`application/eat+cwt; eat_profile="tag:psacertified.org,2023:psa#tfm"`,
		`application/eat+cwt; eat_profile="tag:psacertified.org,2019:psa#legacy"`,
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
	psaToken, err := psatoken.DecodeAndValidateEvidenceFromCOSE(evidence.Data)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	implIDbytes, err := psaToken.Claims.GetImplID()
	if err != nil {
		return nil, err
	}

	instIDbytes, err := psaToken.Claims.GetInstID()
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
	psaToken, err := psatoken.DecodeAndValidateEvidenceFromCOSE(evidence.Data)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	claims, err := common.ToMapViaJSON(psaToken.Claims)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	return claims, nil
}

func (o *Implementation) ValidateEvidenceIntegrity(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
	endorsements []*comid.ValueTriple,
) error {
	pk, err := common.ExtractPublicKeyFromTrustAnchors(trustAnchors)
	if err != nil {
		return fmt.Errorf("could not get public key from trust anchors: %w", err)
	}

	psaToken, err := psatoken.DecodeAndValidateEvidenceFromCOSE(evidence.Data)
	if err != nil {
		return handler.BadEvidence(err)
	}

	psaNonce, err := psaToken.Claims.GetNonce()
	if err != nil {
		return handler.BadEvidence(err)
	}

	if !bytes.Equal(psaNonce, evidence.Nonce) {
		return handler.BadEvidence(
			"freshness: psa-nonce (%s) does not match session nonce (%s)",
			hex.EncodeToString(psaNonce),
			hex.EncodeToString(evidence.Nonce),
		)
	}

	if err = psaToken.Verify(pk); err != nil {
		return handler.BadEvidence(err)
	}
	o.logger.Info("Token signature verified.")

	return nil
}

func (o *Implementation) AppraiseClaims(
	claims map[string]any,
	endorsements []*comid.ValueTriple,
) (*ear.AttestationResult, error) {
	result := handler.CreateAttestationResult(Descriptor.Name)
	appraisal := result.Submods[Descriptor.Name]

	psaClaims, err := convertMapToPSAClaims(claims)
	if err != nil {
		return result, handler.BadEvidence(err)
	}

	rawLifeCycle, err := psaClaims.GetSecurityLifeCycle()
	if err != nil {
		return result, handler.BadEvidence(err)
	}

	// once the signature on the token is verified, we can claim the HW is
	// authentic
	appraisal.TrustVector.Hardware = ear.GenuineHardwareClaim

	lifeCycle := psatoken.LifeCycleToState(rawLifeCycle)
	if lifeCycle == psatoken.StateSecured || lifeCycle == psatoken.StateNonPSAROTDebug {
		appraisal.TrustVector.InstanceIdentity = ear.TrustworthyInstanceClaim
		appraisal.TrustVector.RuntimeOpaque = ear.ApprovedRuntimeClaim
		appraisal.TrustVector.StorageOpaque = ear.HwKeysEncryptedSecretsClaim
	} else {
		appraisal.TrustVector.InstanceIdentity = ear.UntrustworthyInstanceClaim
		appraisal.TrustVector.RuntimeOpaque = ear.VisibleMemoryRuntimeClaim
		appraisal.TrustVector.StorageOpaque = ear.UnencryptedSecretsClaim
	}

	matched, err := matchClaimsToReferenceValues(o.logger, psaClaims, endorsements)
	if err != nil {
		return result, err
	}

	if matched {
		appraisal.TrustVector.Executables = ear.ApprovedRuntimeClaim
		o.logger.Info("Matched software.")
	} else {
		appraisal.TrustVector.Executables = ear.UnrecognizedRuntimeClaim
		o.logger.Info("Failed to match software.")
	}

	appraisal.UpdateStatusFromTrustVector()
	appraisal.VeraisonAnnotatedEvidence = &claims

	return result, nil
}

func convertMapToPSAClaims(m map[string]any) (psatoken.IClaims, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return psatoken.DecodeAndValidateClaimsFromJSON(data)
}

func matchClaimsToReferenceValues(
	logger *zap.SugaredLogger,
	claims psatoken.IClaims,
	endorsements []*comid.ValueTriple,
) (bool, error) {
	referenceValues := make(map[string][2]string)
	for _, triple := range endorsements {
		for _, measurement := range triple.Measurements.Values {
			refValID, err := measurement.Key.GetPSARefValID()
			if err != nil {
				return false, err
			}

			if measurement.Val.Digests == nil {
				return false, errors.New("no digests in reference value measurement")
			}

			numDigests := len(*measurement.Val.Digests)
			if numDigests != 1 {
				return false, fmt.Errorf(
					"expected exactly 1 digest in measurement; found %d",
					numDigests,
				)
			}

			encoded := base64.StdEncoding.EncodeToString((*measurement.Val.Digests)[0].HashValue)
			referenceValues[encoded] = [2]string{*refValID.Label, *refValID.Version}
		}
	}

	swComponents, err := claims.GetSoftwareComponents()
	if err != nil {
		return false, handler.BadEvidence(err)
	}

	for i, swComp := range swComponents {
		mval, err := swComp.GetMeasurementValue()
		if err != nil {
			return false, handler.BadEvidence(fmt.Errorf("S/W comp. %d value: %w", i, err))
		}
		mvalEncoded := base64.StdEncoding.EncodeToString(mval)

		mtype, err := swComp.GetMeasurementType()
		if err != nil {
			return false, handler.BadEvidence(fmt.Errorf("S/W comp. %d type: %w", i, err))
		}

		mversion, err := swComp.GetVersion()
		if err != nil {
			return false, handler.BadEvidence(fmt.Errorf("S/W comp. %d version: %w", i, err))
		}

		rvInfo, matched := referenceValues[mvalEncoded]
		if !matched {
			logger.Debugf("S/W comp. %d measurement failed to match", i)
			return false, nil
		}
		logger.Debugf("S/W comp. %d measurement matched", i)
		refValLabel := rvInfo[0]
		refValVersion := rvInfo[1]

		typeMatched := refValLabel == "" || mtype == refValLabel
		versionMatched := refValVersion == "" || mversion == refValVersion
		logger.Debugf("S/W comp. %d type matched: %t, version matched: %t", i, typeMatched, versionMatched)

		if !typeMatched || !versionMatched {
			return false, nil
		}
	}

	return true, nil
}
