// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cca

import (
	"bytes"
	"encoding/base64"
	"fmt"

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

	// Compare RIM and PV (if provided)
	rimClaim, err := claims.GetInitialMeasurement()
	if err != nil {
		handler.BadEvidence(err)
	}
	pvClaim, err := claims.GetPersonalizationValue()
	if err != nil {
		handler.BadEvidence(err)
	}
	for _, endorsement := range endorsements {
		rimRpvMatch := false
		r, err := cca.GetRim(endorsement.SubScheme, endorsement.Attributes)
		if err != nil {
			return fmt.Errorf("unable to get rim endorsements: %w", err)
		}
		rim, err := base64.StdEncoding.DecodeString(r)
		if err != nil {
			return err
		}
		pv, err := cca.GetRpv(endorsement.SubScheme, endorsement.Attributes)
		if err != nil {
			return fmt.Errorf("unable to get rpv endorsements: %w", err)
		}
		if bytes.Equal(rimClaim, rim) {
			if (pv == nil) || bytes.Equal(pvClaim, pv) {
				log.Debug("Realm Initial Measurements and RPV match")
				rimRpvMatch = true
			}
		}
		if rimRpvMatch {
			appraisal.TrustVector.Executables = ear.ApprovedBootClaim
			remMatch := false
			// Match REM's
			remsClaim, err := claims.GetExtensibleMeasurements()
			if err != nil {
				return handler.BadEvidence(err)
			}
			rems, err := cca.GetRems(endorsement.SubScheme, endorsement.Attributes)
			if err != nil {
				return err
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
			if remMatch {
				appraisal.TrustVector.Executables = ear.ApprovedRuntimeClaim
			} else {
				appraisal.TrustVector.Executables = ear.ContraindicatedRuntimeClaim
			}
			// break from the Endorsement loop once Rim, Rpv match
			break
		} else {
			// TO DO “UNRECOGNIZED_BOOT" is missing
			appraisal.TrustVector.Executables = ear.UnrecognizedRuntimeClaim
		}
	}
	return nil
}
