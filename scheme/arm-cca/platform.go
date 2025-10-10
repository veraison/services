// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package arm_cca

import (
	"fmt"

	"github.com/veraison/ccatoken/platform"
	"github.com/veraison/ear"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/scheme/common"
	"github.com/veraison/services/scheme/common/arm"
)

func platformAppraisal(
	claimsMap map[string]interface{},
	endorsements []handler.Endorsement,
) (*ear.Appraisal, error) {
	claims, err := common.MapToCCAPlatformClaims(claimsMap)
	if err != nil {
		return nil, fmt.Errorf("unable to get claims from platform claims map: %w", err)
	}

	trustVector := ear.TrustVector{}
	// once the signature on the token is verified, we can claim the HW is
	// authentic
	trustVector.Hardware = ear.GenuineHardwareClaim
	rawLifeCycle, err := claims.GetSecurityLifeCycle()
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	lifeCycle := platform.LifeCycleToState(rawLifeCycle)
	if lifeCycle == platform.StateSecured ||
		lifeCycle == platform.StateNonCCAPlatformDebug {
		trustVector.InstanceIdentity = ear.TrustworthyInstanceClaim
		trustVector.RuntimeOpaque = ear.ApprovedRuntimeClaim
		trustVector.StorageOpaque = ear.HwKeysEncryptedSecretsClaim
	} else {
		trustVector.InstanceIdentity = ear.UntrustworthyInstanceClaim
		trustVector.RuntimeOpaque = ear.VisibleMemoryRuntimeClaim
		trustVector.StorageOpaque = ear.UnencryptedSecretsClaim
	}

	swComps := arm.FilterRefVal(endorsements, "platform.sw-component")
	match := arm.MatchSoftware(SchemeName, claims, swComps)
	if match {
		trustVector.Executables = ear.ApprovedRuntimeClaim

	} else {
		trustVector.Executables = ear.UnrecognizedRuntimeClaim
	}

	platformConfig := arm.FilterRefVal(endorsements, "platform.config")
	match = arm.MatchPlatformConfig(SchemeName, claims, platformConfig)

	if match {
		trustVector.Configuration = ear.ApprovedConfigClaim

	} else {
		trustVector.Configuration = ear.UnsafeConfigClaim
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
