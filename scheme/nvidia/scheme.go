// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package nvidia

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/veraison/cmw"
	"github.com/veraison/corim/comid"
	"github.com/veraison/ear"
	ratsdtoken "github.com/veraison/ratsd/ratsd-token"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/vts/appraisal"
	"go.uber.org/zap"
)

const (
	EvidenceMediaTypeRATSd     = `application/eat-ucs+json; eat_profile="` + ratsdtoken.LegacyProfile + `"`
	EvidenceMediaTypeNVJson    = "application/vnd.veraison.nvidia-gpu-evidence+json"
	nvidiaGPUEvidenceCMWKey    = "nv-gpu-evidence"
	attestationReportClaimName = "attestation-report"
	nonceClaimName             = "nonce"
)

var Descriptor = handler.SchemeDescriptor{
	Name:         "NVIDIA",
	VersionMajor: 1,
	VersionMinor: 0,
	CorimProfiles: []string{
		ProfileString,
	},
	EvidenceMediaTypes: []string{
		EvidenceMediaTypeRATSd,
	},
}

var (
	ErrEvidenceMissingCMW     = errors.New("evidence does not contain a CMW collection")
	ErrInvalidEvidenceCMWKind = errors.New("evidence CMW must be a collection")
	ErrInvalidGPUCMWKind      = errors.New("NVIDIA GPU CMW entry must be a monad")
)

type Implementation struct {
	logger *zap.SugaredLogger
	nvat   gpuEvidenceVerifier
}

type gpuEvidenceVerifier interface {
	VerifyEvidence(nonceHex string, evidenceJSON []byte) error
}

func NewImplementation() *Implementation {
	return &Implementation{
		logger: log.Named(Descriptor.Name),
	}
}

func (o *Implementation) Init(params *plugin.Parameters) error {
	opts, err := buildNvatOptions(params)
	if err != nil {
		log.Errorw("failed to build Nvat options", "error", err)
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
	var (
		token  ratsdtoken.Evidence
		claims map[string]any
	)
	if err := token.UnmarshalJSON(evidence.Data); err != nil {
		return nil, handler.BadEvidence(err)
	}

	tokenClaims, err := token.GetClaims()
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	cmwCollection := tokenClaims.GetCMW()
	if cmwCollection == nil {
		return nil, handler.BadEvidence(ErrEvidenceMissingCMW)
	}
	if cmwCollection.GetKind() != cmw.KindCollection {
		return nil, handler.BadEvidence(ErrInvalidEvidenceCMWKind)
	}

	nvGPUCMW, err := cmwCollection.GetCollectionItem(nvidiaGPUEvidenceCMWKey)
	if err != nil {
		return nil, handler.BadEvidence("could not retrieve NVIDIA GPU CMW entry: %w", err)
	}

	if nvGPUCMW.GetKind() != cmw.KindMonad {
		return nil, handler.BadEvidence(ErrInvalidGPUCMWKind)
	}

	mediaType, err := nvGPUCMW.GetMonadType()
	if err != nil {
		return nil, handler.BadEvidence(err)
	}
	if mediaType != EvidenceMediaTypeNVJson {
		return nil, handler.BadEvidence(
			"unexpected NVIDIA GPU CMW media type: got %q, want %q",
			mediaType,
			EvidenceMediaTypeNVJson,
		)
	}

	attestationReport, err := nvGPUCMW.GetMonadValue()
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	claims = make(map[string]any)
	claims[attestationReportClaimName] = attestationReport
	claims[nonceClaimName] = evidence.Nonce

	return claims, nil
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
	result := handler.CreateAttestationResult(Descriptor.Name)
	a := result.Submods[Descriptor.Name]

	a.TrustVector.Hardware = ear.UnsafeHardwareClaim
	a.TrustVector.RuntimeOpaque = ear.VisibleMemoryRuntimeClaim

	attestationReport, ok := claims[attestationReportClaimName].([]byte)
	if !ok {
		return nil, handler.BadEvidence(fmt.Errorf("attestation report has unexpected type %T",
			claims[attestationReportClaimName]))
	}

	nonce, ok := claims[nonceClaimName].([]byte)
	if !ok {
		return nil, handler.BadEvidence(fmt.Errorf("nonce has unexpected type %T",
			claims[nonceClaimName]))
	}

	if o.nvat != nil {
		err := o.nvat.VerifyEvidence(hex.EncodeToString(nonce), attestationReport)
		if err != nil {
			return result, handler.BadEvidence(fmt.Errorf("failed to verify NVIDIA GPU evidence: %w", err))
		}

		a.TrustVector.Hardware = ear.GenuineHardwareClaim
		a.TrustVector.RuntimeOpaque = ear.EncryptedMemoryRuntimeClaim
	}

	a.UpdateStatusFromTrustVector()
	a.VeraisonAnnotatedEvidence = &claims

	return result, nil
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
