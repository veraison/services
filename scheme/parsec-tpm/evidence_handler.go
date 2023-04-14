// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/veraison/ear"
	structpb "google.golang.org/protobuf/types/known/structpb"

	"github.com/veraison/services/handler"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme/common"
)

const (
	ScopeTrustAnchor = "trust anchor"
	ScopeRefValues   = "ref values"
)

type EvidenceHandler struct{}

func (s EvidenceHandler) GetName() string {
	return "parsec-tpm-evidence-handler"
}

func (s EvidenceHandler) GetAttestationScheme() string {
	return SchemeName
}

func (s EvidenceHandler) GetSupportedMediaTypes() []string {
	return EvidenceMediaTypes
}

func (s EvidenceHandler) SynthKeysFromRefValue(tenantID string, refVals *proto.Endorsement) ([]string, error) {
	return synthKeysFromParts(ScopeRefValues, tenantID, refVals.GetAttributes())
}

func (s EvidenceHandler) SynthKeysFromTrustAnchor(tenantID string, ta *proto.Endorsement) ([]string, error) {
	return synthKeysFromParts(ScopeTrustAnchor, tenantID, ta.GetAttributes())
}

func (s EvidenceHandler) GetTrustAnchorID(token *proto.AttestationToken) (string, error) {
	return "", errors.New("TODO(tho)")
}

func (s EvidenceHandler) ExtractClaims(token *proto.AttestationToken, trustAnchor string) (*handler.ExtractedClaims, error) {
	return nil, errors.New("TODO(tho)")
}

func (s EvidenceHandler) ValidateEvidenceIntegrity(token *proto.AttestationToken, trustAnchor string, endorsements []string) error {
	return errors.New("TODO(tho)")
}
func (s EvidenceHandler) AppraiseEvidence(ec *proto.EvidenceContext, endorsementStrings []string) (*ear.AttestationResult, error) {
	return nil, errors.New("TODO(tho)")
}

func synthKeysFromParts(scope, tenantID string, parts *structpb.Struct) ([]string, error) {
	var (
		instance string
		class    string
		fields   map[string]*structpb.Value
		err      error
	)

	fields, err = common.GetFieldsFromParts(parts)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize %s abs-path: %w", scope, err)
	}

	if scope == ScopeTrustAnchor {
		instance, err = common.GetMandatoryPathSegment("parsec-tpm.instance-id", fields)
		if err != nil {
			return nil, fmt.Errorf("unable to synthesize %s abs-path: %w", scope, err)
		}
	}

	class, err = common.GetMandatoryPathSegment("parsec-tpm.class-id", fields)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize %s abs-path: %w", scope, err)
	}

	return []string{parsecTpmLookupKey(scope, tenantID, class, instance)}, nil
}

func parsecTpmLookupKey(scope, tenantID, class, instance string) string {
	var absPath []string

	switch scope {
	case ScopeTrustAnchor:
		absPath = []string{class, instance}
	case ScopeRefValues:
		absPath = []string{class}
	}

	u := url.URL{
		Scheme: SchemeName,
		Host:   tenantID,
		Path:   strings.Join(absPath, "/"),
	}

	return u.String()
}
