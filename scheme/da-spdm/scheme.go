// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package da_spdm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/veraison/corim/comid"
	"github.com/veraison/da"
	"github.com/veraison/ear"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/scheme/common"
	"github.com/veraison/services/vts/appraisal"
	"go.uber.org/zap"
)



const (
	TPM_ALG_SHA_256 uint8 = iota + 1
	TPM_ALG_SHA_384
	TPM_ALG_SHA_512
	TPM_ALG_SHA3_256
	TPM_ALG_SHA3_384
	TPM_ALG_SHA3_512
	TPM_ALG_SM3_256
)

var (
	DaDeviceProfile = "tag:linaro.org,2025:device#1.0.0"
	DaDeviceSpdmProfile = "tag:linaro.org,2025:device-spdm#1.0.0"

	SpdmComponentTypes = map[da.ComponentType]string{
		da.ComponentTypeImmutableROM: "immutable-rom",
		da.ComponentTypeMutableFirmware: "mutable-firmware",
		da.ComponentTypeHardwareConfig: "hardware-config",
		da.ComponentTypeFirmwareConfig: "firmware-config",
		da.ComponentTypeFreeformMeasurementManifest: "freeform-manifest",
		da.ComponentTypeDeviceMode: "device-mode",
		da.ComponentTypeMutableFirmwareVersion: "mutable-firmware-version",
		da.ComponentTypeMutableFirmwareSVN: "mutable-firmware-svn",
		da.ComponentTypeHashExtendMeasurement: "hash-extended-measurement",
		da.ComponentTypeInformational: "informational",
		da.ComponentTypeStructuredMeasurementManifest: "structured-measurement-manifest",
	}

	SpdmHashAlgorithms = map[uint8]comid.DigestAlgorithm{
		TPM_ALG_SHA_256: comid.IntDigestAlgorithm(comid.Sha256),
		TPM_ALG_SHA_384: comid.IntDigestAlgorithm(comid.Sha384),
		TPM_ALG_SHA_512: comid.IntDigestAlgorithm(comid.Sha512),
		TPM_ALG_SHA3_256: comid.IntDigestAlgorithm(comid.Sha3_256),
		TPM_ALG_SHA3_384: comid.IntDigestAlgorithm(comid.Sha3_384),
		TPM_ALG_SHA3_512: comid.IntDigestAlgorithm(comid.Sha3_512),
		TPM_ALG_SM3_256: comid.StringDigestAlgorithm("SM3_256"),
	}

	Descriptor = handler.SchemeDescriptor{
		Name: "DA_SPDM",
		VersionMajor: 1,
		VersionMinor: 0,
		CorimProfiles: []string{
			ProfileString,
		},
		EvidenceMediaTypes: []string{
			fmt.Sprintf("application/eat-ucs+json; eat_profile=%q", DaDeviceProfile),
		},
	}
)

type MatchResult struct {
	Matched bool
	Reason string
}

type Implementation struct{
	logger *zap.SugaredLogger
}

func NewImplementation() *Implementation {
	return &Implementation{
		logger: log.Named(Descriptor.Name),
	}
}

func (o *Implementation) GetReferenceValueIDs(
	trustAnchors []*comid.KeyTriple,
	claims map[string]any,
) ([]*comid.Environment, error) {
	triplesMap, err := clamsToTripleMap(claims)
	if err != nil {
		return nil, err
	}

	ret := make([]*comid.Environment, 0, len(triplesMap))
	for _, triple := range triplesMap {
		ret = append(ret, &triple.Environment)
	}

	return ret, nil
}

func (o *Implementation) ExtractClaims(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
) (map[string]any, error) {
	claims, err := evidenceToTriplesMap(evidence)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	return common.ToMapViaJSON(claims)
}

func (o *Implementation) AppraiseClaims(
	claims map[string]any,
	endorsements []*comid.ValueTriple,
) (*ear.AttestationResult, error) {
	claimsTripleMap, err := clamsToTripleMap(claims)
	if err != nil {
		return nil, err
	}

	endorsementsTripleMap, err := endorsementsToTripleMap(endorsements)
	if err != nil {
		return nil, err
	}

	result := handler.CreateAttestationResult(Descriptor.Name)
	for submodName, submodClaims := range claimsTripleMap {
		appraisal := ear.NewAppraisal()
		result.Submods[submodName] = appraisal

		endorsement, ok := endorsementsTripleMap[submodName]
		if !ok {
			appraisal.TrustVector.Hardware = ear.UnrecognizedHardwareClaim
			continue
		}

		submodClaimsMap, err := common.ToMapViaJSON(submodClaims)
		if err != nil {
			return nil, fmt.Errorf("submod[%s]: %w", submodName, err)
		}
		appraisal.VeraisonAnnotatedEvidence = &submodClaimsMap

		matchResult, err := matchMeasurements(
			endorsement.Measurements.Values,
			submodClaims.Measurements.Values,
		)
		if err != nil {
			return nil, fmt.Errorf("submod[%s]: %w", submodName, err)
		}

		if matchResult.Matched {
			appraisal.TrustVector.Executables = ear.ApprovedRuntimeClaim
		 } else {
			appraisal.TrustVector.Executables = ear.UnrecognizedRuntimeClaim
		}
	}

	// remove the scheme submod from the result, as it it is not used --
	// each SPDM device claims set in evidence has a corresponding submod
	// in the result.
	delete(result.Submods, Descriptor.Name)

	result.UpdateStatusFromTrustVector()
	return result, nil
}

func evidenceToTriplesMap(evidence *appraisal.Evidence) (map[string]*comid.ValueTriple, error) {
	var token da.Token

	err := token.FromCBOR(evidence.Data)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(token.EatNonce[:], evidence.Nonce) {
		return nil, errors.New("nonce in evidence does not match session nonce")
	}

	return daTokenToTriplesMap(&token)
}

func clamsToTripleMap(claims map[string]any) (map[string]*comid.ValueTriple, error) {
	encoded, err := json.Marshal(claims)
	if err != nil {
		return nil, err
	}

	var triplesMap map[string]*comid.ValueTriple
	if err := json.Unmarshal(encoded, &triplesMap); err != nil {
		return nil, err
	}

	return triplesMap, nil
}

func endorsementsToTripleMap(endorsements []*comid.ValueTriple) (map[string]*comid.ValueTriple, error) {
	ret := make(map[string]*comid.ValueTriple)

	for i, endorsement := range endorsements {
		submodName, err := submodNameFromEnvironment(&endorsement.Environment)
		if err != nil {
			return nil, fmt.Errorf("endrosement[%d]: %w", i, err)
		}

		ret[submodName] = endorsement
	}

	return ret, nil
}

func daTokenToTriplesMap(token *da.Token) (map[string]*comid.ValueTriple, error) {
	if len(token.EatSubmods) == 0 {
		return nil, fmt.Errorf("no submods in DA token")
	}

	ret := make(map[string]*comid.ValueTriple)

	for submodName, spdmClaims := range token.EatSubmods {
		if spdmClaims.EatProfile != DaDeviceSpdmProfile {
			return nil, fmt.Errorf(
				"submod[%q]: unsupported profile %q",
				submodName,
				spdmClaims.EatProfile,
			)
		}

		keys := make([]uint8, 0, len(spdmClaims.Measurements))
		for spdmIndex := range spdmClaims.Measurements {
			keys = append(keys, spdmIndex)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})

		measurements := comid.NewMeasurements()
		for _, spdmIndex := range keys {
			spdmMeasurement := spdmClaims.Measurements[spdmIndex]

			measurement, err := comid.NewMeasurement(uint(spdmIndex), comid.UintType)
			if err != nil {
				return nil, fmt.Errorf(
					"submod[%q]: measurement[%d]: %w",
					submodName,
					spdmIndex,
					err,
				)
			}

			componentType, ok := SpdmComponentTypes[da.ComponentType(spdmMeasurement.ComponentType)]
			if !ok {
				return nil, fmt.Errorf(
					"submod[%q]: measurement[%d]: unexpected component type %d",
					submodName,
					spdmIndex,
					spdmMeasurement.ComponentType,
				)
			}
			measurement.SetName(componentType)

			if spdmMeasurement.RawMeasurement != nil { // nolint:gocritic
				measurement.SetRawValueBytes(*spdmMeasurement.RawMeasurement, nil)
			} else if spdmMeasurement.DigestedMeasurement != nil {
				digest := *spdmMeasurement.DigestedMeasurement

				digestAlgorithm, ok := SpdmHashAlgorithms[digest.Algorithm]
				if !ok {
					return nil, fmt.Errorf(
						"submod[%q]: measurement[%d]: unknown algorithm %d",
						submodName,
						spdmIndex,
						digest.Algorithm,
					)
				}

				measurement.Val.Digests = &comid.Digests{
					*comid.NewDigest(digestAlgorithm, digest.Value),
				}
			} else {
				return nil, fmt.Errorf(
					"submod[%q]: measurement[%d]: neither raw measurement nor digest are set",
					submodName,
					spdmIndex,
				)
			}

			measurements.Add(measurement)
		}

		if len(measurements.Values) == 0 {
			return nil, fmt.Errorf("submod[%q]: no measurements", submodName)
		}

		className, err := classNameFromSubmodName(submodName)
		if err != nil {
			return nil, fmt.Errorf("submod[%q]: %w", submodName, err)
		}

		ret[className] = &comid.ValueTriple{
			Environment: comid.Environment{
				Class: comid.NewClassBytes([]byte(className)),
			},
			Measurements: *measurements,
		}
	}

	return ret, nil
}

func classNameFromSubmodName(submodName string) (string, error) {
	if !strings.HasPrefix(submodName, "spdm:") {
		return "", fmt.Errorf("submod name must start with spdm namespace")
	}

	if strings.Contains(submodName, ",") { // nolint:gocritic
		// x.509 Distinguished Name
		parts := strings.Split(submodName[5:], ",")
		out := make([]string, 0, len(parts))

		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "CN=") {
				continue
			}
			out = append(out, part)
		}

		out[0] = fmt.Sprintf("spdm:%s", out[0])

		return strings.Join(out, ","), nil
	} else if strings.Count(submodName, ":") == 3 {
		// ub-DMTF-device-info
		i := strings.LastIndex(submodName, ":")
		return submodName[:i], nil
	} else {
		return "", fmt.Errorf("could not derive class from %q", submodName)
	}
}

func submodNameFromEnvironment(env *comid.Environment) (string, error) {
	if err := validateEnvironment(env); err != nil {
		return "", err
	}

	return string(env.Class.ClassID.Bytes()), nil
}

func matchMeasurements(referenceMeasurements, evidenceMeasurements []comid.Measurement) (MatchResult, error) {
	for i, refMeasurement := range referenceMeasurements {
		mkey, err := refMeasurement.Key.GetKeyUint()
		if err != nil {
			return MatchResult{}, fmt.Errorf("reference measurement[%d]: %w", i, err)
		}

		var evidenceMeasurement *comid.Measurement
		for i, measurement := range evidenceMeasurements {
			evidenceMkey, err := measurement.Key.GetKeyUint()
			if err != nil {
				return MatchResult{}, fmt.Errorf("evidence measurement[%d]: %w", i, err)
			}

			if evidenceMkey == mkey {
				evidenceMeasurement = &measurement
				break
			}
		}

		if evidenceMeasurement == nil {
			return MatchResult{
				Matched:  false,
				Reason: fmt.Sprintf("failed to match mkey %d", mkey),
			}, nil
		}

		if *refMeasurement.Val.Name != *evidenceMeasurement.Val.Name {
			return MatchResult{
				Matched:  false,
				Reason: fmt.Sprintf(
					"mkey %d: reference (%q) and evidence (%q) names don't match",
					mkey,
					*refMeasurement.Val.Name,
					*evidenceMeasurement.Val.Name,
				),
			}, nil
		}

		if refMeasurement.Val.RawValue != nil {
			if evidenceMeasurement.Val.RawValue == nil {
				return MatchResult{
					Matched:  false,
					Reason: fmt.Sprintf("mkey %d: evidence does not contain raw-value", mkey),
				}, nil
			}

			if !evidenceMeasurement.Val.RawValue.CompareAgainstReference(
				refMeasurement.Val.RawValue.Bytes(),
				refMeasurement.Val.RawValue.Mask(),
			) {
				return MatchResult{
					Matched:  false,
					Reason: fmt.Sprintf(
						"mkey %d: raw-value (%x) did not match reference (%x)",
						mkey,
						evidenceMeasurement.Val.RawValue.Bytes(),
						refMeasurement.Val.RawValue.Bytes(),
					),
				}, nil
			}
		} else {
			if evidenceMeasurement.Val.Digests == nil {
				return MatchResult{
					Matched:  false,
					Reason: fmt.Sprintf("mkey %d: evidence does not contain digests", mkey),
				}, nil
			}

			if !evidenceMeasurement.Val.Digests.CompareAgainstReference(
				*refMeasurement.Val.Digests,
			) {
				return MatchResult{
					Matched:  false,
					Reason: fmt.Sprintf("mkey %d: digests did not match reference", mkey),
				}, nil
			}
		}
	}

	return MatchResult{Matched: true}, nil
}
