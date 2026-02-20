// Copyright 2024-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

import (
	"bytes"
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/veraison/corim/comid"
	"github.com/veraison/ear"
	"github.com/veraison/parsec/tpm"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/scheme/common"
	"github.com/veraison/services/vts/appraisal"
	"github.com/veraison/swid"
	"go.uber.org/zap"
)

var Descriptor = handler.SchemeDescriptor{
	Name:         "PARSEC_TPM",
	VersionMajor: 1,
	VersionMinor: 0,
	CorimProfiles: []string{
		ProfileString,
	},
	EvidenceMediaTypes: []string{
		"application/vnd.parallaxsecond.key-attestation.tpm",
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
	var parsecEvidence tpm.Evidence
	err := parsecEvidence.FromCBOR(evidence.Data)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	kat := parsecEvidence.Kat
	if kat == nil {
		return nil, errors.New("no key attestation token to fetch Key ID")
	}

	instanceID, err := comid.NewBytesInstance(*kat.KID)
	if err != nil {
		return nil, err
	}

	return []*comid.Environment{
		{
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
		return nil, fmt.Errorf("expected exactly one trust anchor, got %d", numTAs)
	}

	return []*comid.Environment{
		{
			Class: trustAnchors[0].Environment.Class,
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
	var parsecEvidence tpm.Evidence

	err := parsecEvidence.FromCBOR(evidence.Data)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	claims, err := common.ToMapViaJSON(parsecEvidence)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	return claims, nil
}

func (o *Implementation) ValidateEvidenceIntegrity(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
	endorsements []*comid.ValueTriple,
) error {
	var ev tpm.Evidence

	if err := ev.FromCBOR(evidence.Data); err != nil {
		return handler.BadEvidence(err)
	}

	numKeys := len(trustAnchors[0].VerifKeys)
	if numKeys != 1 {
		return fmt.Errorf("expected exactly one key in trust anchor, found %d", numKeys)
	}

	pk, err := trustAnchors[0].VerifKeys[0].PublicKey()
	if err != nil {
		return fmt.Errorf("could not get public key from trust anchor: %w", err)
	}

	if err := ev.Verify(pk); err != nil {
		return handler.BadEvidence(err)
	}

	o.logger.Debug("Token Signature Verified")
	return nil
}

func (o *Implementation) AppraiseClaims(
	claims map[string]any,
	endorsements []*comid.ValueTriple,
) (*ear.AttestationResult, error) {
	result := handler.CreateAttestationResult(Descriptor.Name)
	appraisal := result.Submods[Descriptor.Name]

	// once the signature on the token is verified, we can claim the HW is
	// authentic
	appraisal.TrustVector.Hardware = ear.GenuineHardwareClaim

	parsecEvidence, err := convertMapToTPMEvidence(claims)
	if err != nil {
		return result, handler.BadEvidence(err)
	}

	attInfo, err := parsecEvidence.Pat.GetAttestationInfo()
	if err != nil {
		return result, handler.BadEvidence(err)
	}

	pcrs := attInfo.PCR.PCRinfo.PCRs
	hashAlgID := attInfo.PCR.PCRinfo.HashAlgID
	pcrDigest := attInfo.PCR.PCRDigest

	sort.Ints(pcrs)

	matched := false
	for i, end := range endorsements {
		o.logger.Debugf("attempting to match endorsement %d...", i)
		endorsedDigest, ok := computeEndorsedHash(o.logger, pcrs, hashAlgID, end.Measurements.Values)
		if !ok {
			continue
		}

		if bytes.Equal(pcrDigest, endorsedDigest) {
			matched = true
			break
		} else {
			o.logger.Debug("failed to match digest")
		}
	}

	if !matched {
		appraisal.TrustVector.Executables = ear.UnrecognizedRuntimeClaim
		return result, handler.BadEvidence("failed to match PCRs")
	}

	o.logger.Debug("PCR digests matched")
	appraisal.TrustVector.Executables = ear.ApprovedRuntimeClaim
	appraisal.TrustVector.InstanceIdentity = ear.TrustworthyInstanceClaim

	// Populate Veraison Key Attestation Extension
	key, err := parsecEvidence.Kat.DecodePubArea()
	if err != nil {
		return result, handler.BadEvidence(err)
	}

	if err := appraisal.SetKeyAttestation(key); err != nil {
		return result, fmt.Errorf("setting extracted public key: %w", err)
	}

	appraisal.UpdateStatusFromTrustVector()
	appraisal.VeraisonAnnotatedEvidence = &claims

	return result, nil
}

func convertMapToTPMEvidence(in map[string]any) (*tpm.Evidence, error) {
	evidence := &tpm.Evidence{}
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	err = evidence.FromJSON(data)
	if err != nil {
		return nil, fmt.Errorf("unable to map to evidence: %w", err)
	}
	return evidence, err
}

func computeEndorsedHash(
	logger *zap.SugaredLogger,
	pcrs []int,
	hashAlgID uint64,
	measurements []comid.Measurement,
) ([]byte, bool) {
	digests := make(map[int][]byte)
	for i, mea := range measurements {
		endPcr, err := mea.Key.GetKeyUint()
		if err != nil {
			logger.Errorf("measurement key at index %d: %w", i, err)
			continue
		}

		if mea.Val.Digests == nil {
			logger.Errorf("no digests in measurement at index %d", i)
			continue
		}

		for _, digest := range *mea.Val.Digests {
			if digest.HashAlgID == hashAlgID {
				digests[int(endPcr)] = digest.HashValue
				break
			}
		}
	}

	var concatHashes []byte
	for _, pcr := range pcrs {
		pcrHash, ok := digests[pcr]
		if !ok {
			logger.Debugf("failed to match PCR %d", pcr)
			return nil, false
		}

		concatHashes = append(concatHashes, pcrHash...)
	}

	hash, err := computeHash(concatHashes, hashAlgID)
	if err != nil {
		logger.Errorf("unable to compute digest: %w", err)
		return nil, false
	}

	return hash, true
}

func computeHash(in []byte, algID uint64) ([]byte, error) {
	h := hashFunc(algID)
	if !h.Available() {
		return nil, fmt.Errorf("unavailable hash function for algID: %d", algID)
	}
	hh := h.New()
	if _, err := hh.Write(in); err != nil {
		return nil, err
	}
	return hh.Sum(nil), nil
}

// hashFunc returns the hash associated with the algorithms supported
// within tpm library
func hashFunc(alg uint64) crypto.Hash {
	switch alg {
	case swid.Sha256:
		return crypto.SHA256
	case swid.Sha384:
		return crypto.SHA384
	case swid.Sha512:
		return crypto.SHA512
	default:
		return 0
	}
}
