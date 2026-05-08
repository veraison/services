// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package arm_cca

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/veraison/ccatoken"
	"github.com/veraison/ccatoken/platform"
	"github.com/veraison/ccatoken/realm"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/profiles/cca"
	"github.com/veraison/ear"
	"github.com/veraison/psatoken"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/scheme/common"
	"github.com/veraison/services/vts/appraisal"
	"go.uber.org/zap"
)

const NUM_REMS = 4

var Descriptor = handler.SchemeDescriptor{
	Name:         "ARM_CCA",
	VersionMajor: 1,
	VersionMinor: 0,
	CorimProfiles: []string{
		cca.PlatformProfileURI,
		cca.RealmProfileURI,
	},
	EvidenceMediaTypes: []string{
		`application/eat-collection; profile="http://arm.com/CCA-SSD/1.0.0"`,
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
	ccaToken, err := ccatoken.DecodeAndValidateEvidenceFromCBOR(evidence.Data)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	implIDbytes, err := ccaToken.PlatformClaims.GetImplID()
	if err != nil {
		return nil, err
	}

	instIDbytes, err := ccaToken.PlatformClaims.GetInstID()
	if err != nil {
		return nil, err
	}

	classID, err := cca.NewPlatformImplIDClassID(implIDbytes)
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

	realmEntry, ok := claims["realm"]
	if !ok {
		return nil, errors.New(`no "realm" entry in claims`)
	}

	realmClaims, err := convertToRealmClaims(realmEntry)
	if err != nil {
		return nil, err
	}

	rimValue, err := realmClaims.GetInitialMeasurement()
	if err != nil {
		return nil, err
	}

	return []*comid.Environment{
		{
			Class: trustAnchors[0].Environment.Class,
		},
		{
			Class: &comid.Class{
				ClassID: comid.MustNewBytesClassID(rimValue),
			},
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
	ccaToken, err := ccatoken.DecodeAndValidateEvidenceFromCBOR(evidence.Data)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	platformClaims, err := common.ToMapViaJSON(ccaToken.PlatformClaims)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	realmClaims, err := common.ToMapViaJSON(ccaToken.RealmClaims)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	claims := map[string]any{
		"platform": platformClaims,
		"realm":    realmClaims,
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

	ccaToken, err := ccatoken.DecodeAndValidateEvidenceFromCBOR(evidence.Data)
	if err != nil {
		return handler.BadEvidence(err)
	}

	realmChallenge, err := ccaToken.RealmClaims.GetChallenge()
	if err != nil {
		return handler.BadEvidence(err)
	}

	// If the provided challenge was less than 64 bytes long, the RMM will
	// zero-pad pad it when generating the attestation token, so do the
	// same to the session nonce.
	sessionNonce := make([]byte, 64)
	copy(sessionNonce, evidence.Nonce)

	if !bytes.Equal(realmChallenge, sessionNonce) {
		return handler.BadEvidence(
			"freshness: realm challenge (%s) does not match session nonce (%s)",
			hex.EncodeToString(realmChallenge),
			hex.EncodeToString(evidence.Nonce),
		)
	}

	if err = ccaToken.Verify(pk); err != nil {
		return handler.BadEvidence(err)
	}
	o.logger.Info("Token signature verified.")

	return nil
}

func (o *Implementation) AppraiseClaims(
	claims map[string]any,
	endorsements []*comid.ValueTriple,
) (*ear.AttestationResult, error) {
	result := handler.CreateAttestationResult("CCA_SSD_PLATFORM")

	platformEntry, ok := claims["platform"]
	if !ok {
		return result, errors.New(`no "platform" entry in claims`)
	}

	err := AppraisePlatform(o.logger, result.Submods["CCA_SSD_PLATFORM"], platformEntry, endorsements)
	if err != nil {
		return result, err
	}

	realmEntry, ok := claims["realm"]
	if !ok {
		return result, errors.New(`no "realm" entry in claims`)
	}

	appraisal := ear.NewAppraisal()
	err = AppraiseRealm(o.logger, appraisal, realmEntry, endorsements)
	if err != nil {
		return result, err
	}
	result.Submods["CCA_REALM"] = appraisal

	return result, nil
}

func AppraisePlatform(
	logger *zap.SugaredLogger,
	appraisal *ear.Appraisal,
	entry any,
	endorsements []*comid.ValueTriple,
) error {
	claims, err := convertToPlatformClaims(entry)
	if err != nil {
		return handler.BadEvidence(err)
	}

	// once the signature on the token is verified, we can claim the HW is
	// authentic
	appraisal.TrustVector.Hardware = ear.GenuineHardwareClaim

	rawLifeCycle, err := claims.GetSecurityLifeCycle()
	if err != nil {
		return handler.BadEvidence(err)
	}

	lifeCycle := platform.LifeCycleToState(rawLifeCycle)
	if lifeCycle == platform.StateSecured ||
		lifeCycle == platform.StateNonCCAPlatformDebug {
		appraisal.TrustVector.InstanceIdentity = ear.TrustworthyInstanceClaim
		appraisal.TrustVector.RuntimeOpaque = ear.ApprovedRuntimeClaim
		appraisal.TrustVector.StorageOpaque = ear.HwKeysEncryptedSecretsClaim
	} else {
		appraisal.TrustVector.InstanceIdentity = ear.UntrustworthyInstanceClaim
		appraisal.TrustVector.RuntimeOpaque = ear.VisibleMemoryRuntimeClaim
		appraisal.TrustVector.StorageOpaque = ear.UnencryptedSecretsClaim
	}

	platformMatched, swMatched, err := matchPlatformClaimsToReferenceValues(logger, claims, endorsements)
	if err != nil {
		return err
	}

	if platformMatched {
		appraisal.TrustVector.Configuration = ear.ApprovedConfigClaim

	} else {
		appraisal.TrustVector.Configuration = ear.UnsafeConfigClaim
	}

	if swMatched {
		appraisal.TrustVector.Executables = ear.ApprovedRuntimeClaim

	} else {
		appraisal.TrustVector.Executables = ear.UnrecognizedRuntimeClaim
	}

	claimsMap, err := common.ToMapViaJSON(claims)
	if err != nil {
		return err
	}

	appraisal.UpdateStatusFromTrustVector()
	appraisal.VeraisonAnnotatedEvidence = &claimsMap
	return nil
}

type realmReference struct {
	PersonalizationValue   []byte
	InitialMeasurement     []byte
	ExtensibleMeasurements [][]byte
}

func AppraiseRealm(
	logger *zap.SugaredLogger,
	appraisal *ear.Appraisal,
	entry any,
	endorsements []*comid.ValueTriple,
) error {
	claims, err := convertToRealmClaims(entry)
	if err != nil {
		return handler.BadEvidence(err)
	}

	evidenceRIM, err := claims.GetInitialMeasurement()
	if err != nil {
		return handler.BadEvidence(err)
	}

	evidenceREMs, err := claims.GetExtensibleMeasurements()
	if err != nil {
		return handler.BadEvidence(err)
	}

	evidencePV, err := claims.GetPersonalizationValue()
	if err != nil {
		return handler.BadEvidence(err)
	}

	// If crypto verification (including chaining) completes correctly,
	// we can safely assume the Realm instance to be trustworthy
	appraisal.TrustVector.InstanceIdentity = ear.TrustworthyInstanceClaim
	appraisal.TrustVector.Executables = ear.UnrecognizedRuntimeClaim

	logger.Debug("collecting realm reference values...")
	referenceValues := make([]realmReference, 0, len(endorsements))
	for _, triple := range endorsements {
		refVal := realmReference{
			ExtensibleMeasurements: make([][]byte, NUM_REMS),
		}

		for _, measurement := range triple.Measurements.Values {
			mkey, err := readMeasurementKey(&measurement)
			if err != nil {
				return err
			}

			switch mkey {
			case cca.CCARealmInitialMeasurementMkey:
				digest, err := readMeasurementDigestBytes(&measurement)
				if err != nil {
					return fmt.Errorf("%s: %w", cca.CCARealmInitialMeasurementMkey, err)
				}
				refVal.InitialMeasurement = digest
			case cca.CCARealmExtendedMeasurement0Mkey:
				digest, err := readMeasurementDigestBytes(&measurement)
				if err != nil {
					return fmt.Errorf("%s: %w", cca.CCARealmExtendedMeasurement0Mkey, err)
				}
				refVal.ExtensibleMeasurements[0] = digest
			case cca.CCARealmExtendedMeasurement1Mkey:
				digest, err := readMeasurementDigestBytes(&measurement)
				if err != nil {
					return fmt.Errorf("%s: %w", cca.CCARealmExtendedMeasurement1Mkey, err)
				}
				refVal.ExtensibleMeasurements[1] = digest
			case cca.CCARealmExtendedMeasurement2Mkey:
				digest, err := readMeasurementDigestBytes(&measurement)
				if err != nil {
					return fmt.Errorf("%s: %w", cca.CCARealmExtendedMeasurement2Mkey, err)
				}
				refVal.ExtensibleMeasurements[2] = digest
			case cca.CCARealmExtendedMeasurement3Mkey:
				digest, err := readMeasurementDigestBytes(&measurement)
				if err != nil {
					return fmt.Errorf("%s: %w", cca.CCARealmExtendedMeasurement3Mkey, err)
				}
				refVal.ExtensibleMeasurements[3] = digest
			case cca.CCARealmPersonalizationMkey:
				if measurement.Val.RawValue == nil {
					return fmt.Errorf("raw-value not set for %s",
						cca.CCARealmPersonalizationMkey)
				}

				refVal.PersonalizationValue, err = measurement.Val.RawValue.GetBytes()
				if err != nil {
					return fmt.Errorf("%s: %w", cca.CCARealmPersonalizationMkey, err)
				}
			default:
				logger.Debugw("skipping non-realm measurement", "mkey", mkey)
			}

		}

		if refVal.InitialMeasurement != nil {
			referenceValues = append(referenceValues, refVal)
		}
	}
	logger.Debug("collected realm reference values", "values", referenceValues)

	for _, refVal := range referenceValues {
		if !bytes.Equal(refVal.InitialMeasurement, evidenceRIM) {
			// For this CCA Realm scheme, as RIM fetches all the Endorsements, for now
			// failure to match RIM means some serious issue with the Verifier
			appraisal.TrustVector.SetAll(ear.VerifierMalfunctionClaim)
			break
		}

		// Note: if an Endorser does not use RPV it indicates one Realm per RIM, which is a match
		if refVal.PersonalizationValue != nil && !bytes.Equal(refVal.PersonalizationValue, evidencePV) {
			appraisal.TrustVector.Executables = ear.ContraindicatedRuntimeClaim
			continue // continue looking for other RPVs matching the same RIM
		}

		appraisal.TrustVector.Executables = ear.ApprovedBootClaim
		logger.Debug("RIM & RPV matched")

		if allMatch(refVal.ExtensibleMeasurements, evidenceREMs) {
			appraisal.TrustVector.Executables = ear.ApprovedRuntimeClaim
			logger.Debug("REMs matched")
		} else {
			logger.Debug("REMs failed to match")
		}

		break
	}

	claimsMap, err := common.ToMapViaJSON(claims)
	if err != nil {
		return err
	}

	appraisal.UpdateStatusFromTrustVector()
	appraisal.VeraisonAnnotatedEvidence = &claimsMap

	return nil
}

func convertToRealmClaims(v any) (realm.IClaims, error) {
	encoded, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return realm.DecodeClaimsFromJSON(encoded)
}

func convertToPlatformClaims(v any) (platform.IClaims, error) {
	encoded, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return platform.DecodeClaimsFromJSON(encoded)
}

func readMeasurementKey(measurement *comid.Measurement) (string, error) {
	if measurement.Key == nil || !measurement.Key.IsSet() {
		return "", errors.New(" measurement missing mkey")
	}

	if measurement.Key.Type() != comid.StringType {
		return "", fmt.Errorf(
			"measurement mkey must be string, got %s",
			measurement.Key.Type(),
		)
	}

	return measurement.Key.Value.String(), nil
}

func readMeasurementDigestBytes(measurement *comid.Measurement) ([]byte, error) {
	if measurement.Val.Digests == nil {
		return nil, errors.New("no digests in reference value measurement")
	}

	numDigests := len(*measurement.Val.Digests)
	if numDigests != 1 {
		return nil, fmt.Errorf(
			"expected exactly 1 digest in measurement; found %d",
			numDigests,
		)
	}

	return (*measurement.Val.Digests)[0].HashValue, nil
}

func matchPlatformClaimsToReferenceValues(
	logger *zap.SugaredLogger,
	claims platform.IClaims,
	endorsements []*comid.ValueTriple,
) (bool, bool, error) {
	var err error
	var referenceConfigValue []byte

	logger.Debug("building platform reference values map...")
	referenceValues := make(map[string][2]string)
	for _, triple := range endorsements {
		for _, measurement := range triple.Measurements.Values {
			mkey, err := readMeasurementKey(&measurement)
			if err != nil {
				return false, false, err
			}

			// Check if this is a platform config measurement.
			switch mkey {
			case cca.CCAPlatformConfigMkey:
				if measurement.Val.RawValue == nil {
					return false, false,
						errors.New("no raw value in platform config measurement")
				}

				referenceConfigValue, err = measurement.Val.RawValue.GetBytes()
				if err != nil {
					return false, false, err
				}

				continue
			case cca.CCASoftwareComponentMkey:
				digest, err := readMeasurementDigestBytes(&measurement)
				if err != nil {
					return false, false, err
				}

				encoded := base64.StdEncoding.EncodeToString(digest)
				// Extract label (mtype) and version from measurement value
				var label, version string

				if measurement.Val.Name != nil {
					label = *measurement.Val.Name
				}
				if measurement.Val.Ver != nil {
					version = measurement.Val.Ver.Version
				}

				referenceValues[encoded] = [2]string{label, version}
			default:
				logger.Debugw("skipping non-platform measurement", "mkey", mkey)
			}
		}
	}
	logger.Debugw("platform reference values", "map", referenceValues)

	evidenceConfigValue, err := claims.GetConfig()
	if err != nil {
		return false, false, handler.BadEvidence(err)
	}

	configMatched := false
	if bytes.Equal(evidenceConfigValue, referenceConfigValue) {
		logger.Debug("platform config matched")
		configMatched = true
	} else {
		logger.Debug("platform config failed to match")
	}

	swComponents, err := claims.GetSoftwareComponents()
	if err != nil {
		return false, false, handler.BadEvidence(err)
	}

	for i, swComp := range swComponents {
		mval, err := swComp.GetMeasurementValue()
		if err != nil {
			return false, false, handler.BadEvidence(fmt.Errorf("S/W comp. %d value: %w", i, err))
		}
		mvalEncoded := base64.StdEncoding.EncodeToString(mval)

		mtype, err := swComp.GetMeasurementType()
		if err != nil {
			return false, false, handler.BadEvidence(fmt.Errorf("S/W comp. %d type: %w", i, err))
		}

		mversion, err := swComp.GetVersion()
		if err != nil && !errors.Is(err, psatoken.ErrOptionalFieldMissing) {
			return false, false, handler.BadEvidence(fmt.Errorf("S/W comp. %d version: %w", i, err))
		}

		rvInfo, matched := referenceValues[mvalEncoded]
		if !matched {
			logger.Debugf("S/W comp. %d measurement (%s) failed to match", i, mvalEncoded)
			return configMatched, false, nil
		}
		logger.Debugf("S/W comp. %d measurement (%s) matched", i, mvalEncoded)
		refValLabel := rvInfo[0]
		refValVersion := rvInfo[1]

		typeMatched := refValLabel == "" || mtype == refValLabel
		versionMatched := refValVersion == "" || mversion == refValVersion
		logger.Debugf("S/W comp. %d type matched: %t (%s), version matched: %t (%s)",
			i, typeMatched, mtype, versionMatched, mversion)

		if !typeMatched || !versionMatched {
			return configMatched, false, nil
		}
	}

	return configMatched, true, nil
}

func allMatch(lhs, rhs [][]byte) bool {
	if len(lhs) != len(rhs) {
		return false
	}

	for i, lhsV := range lhs {
		if len(lhsV) > 0 && !bytes.Equal(lhsV, rhs[i]) {
			return false
		}
	}

	return true
}
