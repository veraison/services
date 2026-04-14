// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package nvidia

import (
	"errors"
	"strings"

	"github.com/veraison/corim/comid"
	"github.com/veraison/ear"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/vts/appraisal"
	"go.uber.org/zap"
)

var (
	EvidenceMediaTypeNVCbor = "application/vnd.veraison.nvidia-attestation-report+cbor"
	EvidenceMediaTypeNVJson = "application/vnd.veraison.nvidia-attestation-report+json"
)

var Descriptor = handler.SchemeDescriptor{
	Name:         "NVIDIA",
	VersionMajor: 1,
	VersionMinor: 0,
	CorimProfiles: []string{
		ProfileString,
	},
	EvidenceMediaTypes: []string{
		EvidenceMediaTypeNVCbor,
		EvidenceMediaTypeNVJson,
	},
}

type Implementation struct {
	logger *zap.SugaredLogger
	nvat   *NvatGpuInterface
}

func NewImplementation() *Implementation {
	return &Implementation{
		logger: log.Named(Descriptor.Name),
	}
}

func (o *Implementation) Init(params *plugin.Parameters) error {
	opts, err := buildNvatOptions(params)
	if err != nil {
		return err
	}

	o.nvat, err = NewNvatGpuInterface(opts)
	return err
}

func (o *Implementation) GetTrustAnchorIDs(
	evidence *appraisal.Evidence,
) ([]*comid.Environment, error) {
	return nil, nil
}

func (o *Implementation) GetReferenceValueIDs(
	trustAnchors []*comid.KeyTriple,
	claims map[string]any,
) ([]*comid.Environment, error) {
	return nil, nil
}

func (o *Implementation) ExtractClaims(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
) (map[string]any, error) {
	return nil, nil
}

func (o *Implementation) ValidateEvidenceIntegrity(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
	endorsements []*comid.ValueTriple,
) error {
	return nil
}

func (o *Implementation) AppraiseClaims(
	claims map[string]any,
	endorsements []*comid.ValueTriple,
) (*ear.AttestationResult, error) {
	if o.nvat != nil {
		_ = o.nvat.VerifyEvidence("", []byte(""))
	}
	return nil, nil
}

func buildNvatOptions(params *plugin.Parameters) (NvatOptions, error) {
	opts := DefaultNvatOptions()

	if value, err := readStringParam(params, "service_token"); err != nil {
		return NvatOptions{}, err
	} else if value != "" {
		opts.ServiceToken = value
	}
	if value, err := readStringParam(params, "verifier_mode"); err != nil {
		return NvatOptions{}, err
	} else if value != "" {
		opts.VerifierMode = value
	}

	if value, err := readStringParam(params, "catrust_file"); err != nil {
		return NvatOptions{}, err
	} else if value != "" {
		opts.CaTrustFile = value
	}
	if value, err := readStringParam(params, "remote_host"); err != nil {
		return NvatOptions{}, err
	} else if value != "" {
		opts.RemoteHost = strings.TrimRight(value, "/")
	}
	if err := opts.Valid(); err != nil {
		return NvatOptions{}, err
	}

	return opts, nil
}

func readStringParam(params *plugin.Parameters, key string) (string, error) {
	value, err := params.GetString(key)
	if err != nil {
		if errors.Is(err, plugin.ErrNotSet) {
			return "", nil
		}
		return "", err
	}
	return value, nil
}
