// Copyright 2021-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package arm

import (
	"fmt"

	"github.com/veraison/ccatoken"
	parsec_cca "github.com/veraison/parsec/cca"
	"github.com/veraison/psatoken"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme/common"
)

func SynthKeysFromRefValue(scheme string, tenantID string,
	refVal *handler.Endorsement,
) ([]string, error) {

	implID, err := common.GetImplID(scheme, refVal.Attributes)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize reference value: %w", err)
	}

	lookupKey := RefValLookupKey(scheme, tenantID, implID)
	log.Debugf("Scheme %s Plugin Reference Value Look Up Key= %s\n", scheme, lookupKey)

	return []string{lookupKey}, nil

}

func SynthKeysFromTrustAnchors(scheme string, tenantID string,
	ta *handler.Endorsement,
) ([]string, error) {

	implID, err := common.GetImplID(scheme, ta.Attributes)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize reference value: %w", err)
	}

	instID, err := common.GetInstID(scheme, ta.Attributes)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize trust anchor abs-path: %w", err)
	}

	lookupKey := TaLookupKey(scheme, tenantID, implID, instID)
	log.Debugf("Scheme %s Plugin TA Look Up Key= %s\n", scheme, lookupKey)
	return []string{lookupKey}, nil
}

func GetTrustAnchorID(scheme string, token *proto.AttestationToken) (string, error) {
	var claims psatoken.IClaims

	switch scheme {
	case "PSA_IOT":
		var psaToken psatoken.Evidence

		err := psaToken.FromCOSE(token.Data)
		if err != nil {
			return "", handler.BadEvidence(err)
		}
		claims = psaToken.Claims

	case "CCA_SSD_PLATFORM":
		var evidence ccatoken.Evidence

		err := evidence.FromCBOR(token.Data)
		if err != nil {
			return "", handler.BadEvidence(err)
		}

		claims = evidence.PlatformClaims

	case "PARSEC_CCA":
		var evidence parsec_cca.Evidence

		err := evidence.FromCBOR(token.Data)
		if err != nil {
			return "", handler.BadEvidence(err)
		}
		claims = evidence.Pat.PlatformClaims
	default:
		return "", fmt.Errorf("invalid scheme argument to GetTrustAnchorID : %s", scheme)

	}

	return TaLookupKey(
		scheme,
		token.TenantId,
		MustImplIDString(claims),
		MustInstIDString(claims),
	), nil
}
