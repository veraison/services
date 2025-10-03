// Copyright 2023-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cca

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"

	ccatokenrealm "github.com/veraison/ccatoken/realm"
	"github.com/veraison/ear"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/scheme/common/arm"
	"github.com/veraison/services/scheme/common/cca/realm"
)

var (
	ErrKeyNotFound    = errors.New("key not found")
	ErrValuesMismatch = errors.New("values mismatch")
)

func realmAppraisal(
	claimsMap map[string]interface{},
	endorsements []handler.Endorsement) (*ear.Appraisal, error) {
	realmEndorsements := arm.FilterRefVal(endorsements, "realm.reference-value")
	claims, err := realm.MapToRealmClaims(claimsMap)
	if err != nil {
		return nil, fmt.Errorf("unable to get claims from realm claims map: %w", err)
	}

	trustVector := ear.TrustVector{}
	// If crypto verification (including chaining) completes correctly,
	// we can safely assume the Realm instance to be trustworthy
	trustVector.InstanceIdentity = ear.TrustworthyInstanceClaim
	trustVector.Executables = ear.UnrecognizedRuntimeClaim

	for _, endorsement := range realmEndorsements {
		if matchRim(claims, &endorsement) {
			err := matchRpv(claims, &endorsement)
			switch err {
			// Note, If an Endorser does not use RPV it indicates, one Realm per RIM, which is a match
			case nil, ErrKeyNotFound:
				trustVector.Executables = ear.ApprovedBootClaim
			case ErrValuesMismatch:
				trustVector.Executables = ear.ContraindicatedRuntimeClaim
				continue // continue looking for other RPVs matching the same RIM
			default:
				// Some serious error has happened, report it as VerifierMalfunction
				log.Errorf("fatal error in matchRpv %w", err)
				trustVector.SetAll(ear.VerifierMalfunctionClaim)
			}

			// Only match REMs when RIM and RPV match
			if trustVector.Executables == ear.ApprovedBootClaim {
				// Match REM's
				if matchREMs(claims, &endorsement) {
					trustVector.Executables = ear.ApprovedRuntimeClaim
				}
			}
			break
		} else {
			// For this CCA Realm scheme, as RIM fetches all the Endorsements, for now
			// failure to match RIM means some serious issue with the Verifier
			trustVector.SetAll(ear.VerifierMalfunctionClaim)
		}
	}
	var status ear.TrustTier
	appraisal := ear.Appraisal{
		Status:      &status,
		TrustVector: &trustVector,
	}

	appraisal.UpdateStatusFromTrustVector()
	appraisal.VeraisonAnnotatedEvidence = &claimsMap

	return &appraisal, nil
}

func matchRim(claims ccatokenrealm.IClaims, endorsement *handler.Endorsement) bool {
	// get RIM Claim from Evidence Claims
	rimClaim, err := claims.GetInitialMeasurement()
	if err != nil {
		log.Errorf("failed to extract rim measurements: %w", err)
		return false
	}
	// get RIM from endorsements
	r, err := realm.GetRIM(endorsement.Attributes)
	if err != nil {
		log.Errorf("unable to get rim endorsements: %w", err)
		return false
	}
	rim, err := base64.StdEncoding.DecodeString(r)
	if err != nil {
		log.Errorf("base64 decode failed: %w", err)
		return false
	}
	if !bytes.Equal(rimClaim, rim) {
		log.Errorf("fatal error in Rim matching, evidence rim %v, endorsement rim %v", rimClaim, rim)
		return false
	}
	return true
}

func matchRpv(claims ccatokenrealm.IClaims, endorsement *handler.Endorsement) error {
	pvClaim, err := claims.GetPersonalizationValue()
	if err != nil {
		return fmt.Errorf("matchRpv failed: %w", err)
	}
	rpv, err := realm.GetRPV(endorsement.Attributes)
	if err != nil {
		return fmt.Errorf("unable to get rpv endorsements: %w", err)
	}
	if rpv == nil {
		return ErrKeyNotFound
	}
	if !bytes.Equal(pvClaim, rpv) {
		return ErrValuesMismatch
	}
	return nil
}

func matchREMs(claims ccatokenrealm.IClaims, endorsement *handler.Endorsement) bool {
	remMatch := false
	remsClaim, err := claims.GetExtensibleMeasurements()
	if err != nil {
		log.Errorf("unable to get realm extensible measurements from claims: %w", err)
		return remMatch
	}
	rems, err := realm.GetREMs(endorsement.Attributes)
	if err != nil {
		log.Errorf("unable to get REM endorsements: %w", err)
		return remMatch
	}
	for i, rem := range rems {
		if bytes.Equal(remsClaim[i], rem) {
			log.Debugf("Realm Extended Measurement match at index: %d", i)
			remMatch = true
		} else {
			log.Debugf("Realm Extended Measurement does not match at index: %d", i)
			remMatch = false
			break /* the rem loop */
		}
	}
	return remMatch
}
