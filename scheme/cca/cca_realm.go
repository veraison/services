// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cca

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/veraison/ccatoken"
	"github.com/veraison/ear"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/scheme/common/arm/cca"
)

type Cca_realm_attester struct {
}

func (Cca_realm_attester) PerformAppraisal(
	appraisal *ear.Appraisal,
	claimsMap map[string]interface{},
	endorsements []handler.Endorsement) error {

	claims, err := cca.MapToRealmClaims(claimsMap)
	if err != nil {
		return fmt.Errorf("unable to get realm claims from Realm claims map: %w", err)
	}

	// If crypto verification (including chaining) completes correctly,
	// we can safely assume the Realm instance to be trustworthy
	appraisal.TrustVector.InstanceIdentity = ear.TrustworthyInstanceClaim

	for _, endorsement := range endorsements {
		rimRpvMatch := matchRimRpv(claims, &endorsement)
		if rimRpvMatch {
			appraisal.TrustVector.Executables = ear.ApprovedBootClaim
			// Match REM's
			remMatch := matchRem(claims, &endorsement)
			if remMatch {
				appraisal.TrustVector.Executables = ear.ApprovedRuntimeClaim
			}
			break
		} else {
			// TO DO “UNRECOGNIZED_BOOT" is missing
			appraisal.TrustVector.Executables = ear.ContraindicatedRuntimeClaim
		}
	}
	appraisal.UpdateStatusFromTrustVector()
	appraisal.VeraisonAnnotatedEvidence = &claimsMap

	return nil
}

func matchRimRpv(claims ccatoken.IClaims, endorsement *handler.Endorsement) bool {
	// Compare RIM and PV (if provided)
	rimClaim, err := claims.GetInitialMeasurement()
	if err != nil {
		log.Errorf("matchRimRpv failed: %w", handler.BadEvidence(err))
		return false
	}
	pvClaim, err := claims.GetPersonalizationValue()
	if err != nil {
		log.Errorf("matchRimRpv failed: %w", handler.BadEvidence(err))
		return false
	}

	r, err := cca.GetRim(endorsement.SubScheme, endorsement.Attributes)
	if err != nil {
		log.Errorf("unable to get rim endorsements: %w", err)
		return false
	}
	rim, err := base64.StdEncoding.DecodeString(r)
	if err != nil {
		log.Errorf("matchRimRpv failed: %w", err)
		return false
	}
	pv, err := cca.GetRpv(endorsement.SubScheme, endorsement.Attributes)
	if err != nil {
		log.Errorf("unable to get rpv endorsements: %w", err)
		return false
	}
	if bytes.Equal(rimClaim, rim) {
		if (pv == nil) || bytes.Equal(pvClaim, pv) {
			log.Debug("Realm Initial Measurements and RPV match")
			return true
		}
	}
	return false
}

func matchRem(claims ccatoken.IClaims, endorsement *handler.Endorsement) bool {
	remMatch := false
	remsClaim, err := claims.GetExtensibleMeasurements()
	if err != nil {
		log.Errorf("unable to get realm extensible measurements from claims: %w", handler.BadEvidence(err))
		return remMatch
	}
	rems, err := cca.GetRems(endorsement.SubScheme, endorsement.Attributes)
	if err != nil {
		log.Errorf("unable to get REM endorsements: %w", err)
		return remMatch
	}
	for i, rem := range rems {
		if bytes.Equal(remsClaim[i], rem) {
			log.Debugf("Realm Extended Measurement match at index: %d", i)
			remMatch = true
		} else {
			remMatch = false
			break /* the rem loop */
		}
	}
	return remMatch
}
