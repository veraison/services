// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"mime"

	"go.uber.org/zap"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/ear"
	"github.com/veraison/services/pkg/config"
	"github.com/veraison/services/pkg/endorsement"
	"github.com/veraison/services/pkg/fault"
	"github.com/veraison/services/pkg/policy"
	"github.com/veraison/services/pkg/scheme"
	"github.com/veraison/services/pkg/vts"
)

type TrustedService struct {
	EndorsementStore *endorsement.Store
	SchemeManager    *scheme.Manager
	PolicyManager    *policy.Manager
	EarSigner        IEarSigner
	RootCerts        *x509.CertPool

	logger *zap.SugaredLogger
}

func NewService(
	endorsementStore *endorsement.Store,
	schemeManager *scheme.Manager,
	policyManager *policy.Manager,
	earSigner IEarSigner,
	rootCerts *x509.CertPool,
	logger *zap.SugaredLogger,
) *TrustedService {
	return &TrustedService{
		EndorsementStore: endorsementStore,
		SchemeManager:    schemeManager,
		PolicyManager:    policyManager,
		EarSigner:        earSigner,
		RootCerts:        rootCerts,
		logger:           logger,
	}
}

func (o *TrustedService) GetServiceState(ctx context.Context) (*vts.State, error) {
	o.logger.Info("GetServiceState")
	verificationMediaTypes, err := o.getSupportedVerificationMediaTypes()
	if err != nil {
		return nil, err
	}

	provisioningMediaTypes, err := o.getSupportedProvisioningMediaTypes()
	if err != nil {
		return nil, err
	}

	coservMediaTypes, err := o.getSupportedCoserMediaTypes()
	if err != nil {
		return nil, err
	}

	earPublicKey, err := o.getEARSigningPublicKey()
	if err != nil {
		return nil, err
	}

	coservPublicKey, err := o.getCoservSigningPublicKey()
	if err != nil {
		return nil, err
	}

	serviceState := &vts.State{
		Version:                config.Version,
		ProvisioningMediaTypes: provisioningMediaTypes,
		VerificationMediaTypes: verificationMediaTypes,
		CoservMediaTypes:       coservMediaTypes,
		EarSigningPublicKey:    earPublicKey,
		CoservSigningPublicKey: coservPublicKey,
	}

	o.logger.Infow("obtained service state", "state", serviceState)
	return serviceState, nil
}

func (o *TrustedService) SubmitEndorsements(
	ctx context.Context,
	mediaType string,
	data []byte,
) error {
	tenantID := vts.GetTenantID(ctx)
	o.logger.Infow("SubmitEndorsements", "media-type", mediaType, "tenant-id", tenantID)

	mt, mtParams, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return err
	}
	profile := mtParams["profile"]

	var uc *corim.UnsignedCorim
	switch mt {
	case "application/rim+cose":
		uc, err = o.decodeAndValidateSignedCorim(data)
		if err != nil {
			return err
		}
	case "application/rim+cbor":
		uc, err = o.decodeAndValidateUnsignedCorim(data)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported media type: %s", mediaType)
	}

	if uc.Profile == nil {
		return errors.New("profile not set in CoRIM")
	}

	ucProfile := uc.Profile.String()
	o.logger.Debugw("CoRIM profile", "profile", ucProfile)

	if ucProfile != profile {
		err := fmt.Errorf(
			"CoRIM profile (%s) does not match media type profile (%s)",
			ucProfile,
			profile,
		)
		return err
	}

	handlerPlugin, err := o.SchemeManager.LookupByCorimProfile(profile)
	if err != nil {
		return err
	}

	resp, err := handlerPlugin.ValidateCorim(uc)
	if err != nil {
		return err
	} else if !resp.IsValid {
		return resp.Error()
	}

	label := fmt.Sprintf("%s/%s", tenantID, handlerPlugin.GetAttestationScheme())
	if err := o.EndorsementStore.AddCorimBytes(data, label, true); err != nil {
		return err
	}

	o.logger.Info("CoRIM added to store")
	return nil
}

func (o *TrustedService) GetAttestation(
	ctx context.Context,
	evidence *scheme.Evidence,
) ([]byte, error) {
	tenantID := vts.GetTenantID(ctx)
	o.logger.Infow("get attestation", "media-type", evidence.MediaType, "tenant-id", tenantID)

	// This result will be returned if we encounter an error before
	// obtaining the real result during AppraiseClaims() below.
	result := ear.NewAttestationResult("ERROR", config.Version, config.Developer)

	encodedNonce := base64.RawURLEncoding.EncodeToString(evidence.Nonce)
	result.Nonce = &encodedNonce

	handler, err := o.SchemeManager.LookupByEvidenceMediaType(evidence.MediaType)
	if err != nil {
		setAllClaims(result, ear.UnexpectedEvidenceClaim)
		addPolicyClaim(result, "problem", "could not resolve media type")
		return o.finalize(result, err)
	}

	schemeName := handler.GetAttestationScheme()
	if err := setScheme(result, schemeName); err != nil {
		return o.finalize(result, err)
	}

	storeLabel := fmt.Sprintf("%s/%s", tenantID, schemeName)

	trustAnchorIDs, err := handler.GetTrustAnchorIDs(evidence)
	if err != nil {
		if errors.Is(err, fault.ErrBadEvidence) {
			setAllClaims(result, ear.CryptoValidationFailedClaim)
			addPolicyClaim(result, "problem", "could not establish identity from evidence")
		}

		return o.finalize(result, err)
	}

	// TODO(setrofim): in principle, we should be matching exactly
	// here, as we are trying to identify triples for a specific
	// entity described by the taID environment. However, CoRIM does
	// not provide a good way to endorse an environment description. So
	// in practice, environments inside CoRIMs may serve
	// the double duty of providing a way to match measurements
	// _and_ an additional description of the environments that is
	// not present in the evidence.
	//
	// Specifically, our parsec-tpm scheme currently relies on the
	// fact that the trust anchors are provisioned with environment
	// that contains a class-id as well as an instance-id; the
	// evidence only contains an instance-id, and the class-id in
	// the trust anchor is then used to retrieve reference values;
	// there is no way to directly link reference values to the
	// evidence.
	//
	// Because provisioned environments may contain "additional"
	// descriptive elements, as well as elements used for matching,
	// we are forced to do inexact matching here for now, and leave
	// it to the attestation schemes to resolve this.
	matchExactly := false
	var trustAnchors []*comid.KeyTriple
	trustAnchors, err = o.getKeyTriples(trustAnchorIDs, storeLabel, matchExactly)

	if err != nil {
		if errors.Is(err, fault.ErrNotFound) {
			// finalize will be called in the next block
			o.logger.Warn("no trust anchor in store")
		} else {
			return o.finalize(result, err)
		}
	}
	if trustAnchors == nil {
		err = fault.BadEvidence("no trust anchor for %s", toJSON(trustAnchorIDs))
		setAllClaims(result, ear.CryptoValidationFailedClaim)
		addPolicyClaim(result, "problem", "no trust anchor for evidence")
		return o.finalize(result, err)
	}

	claims, err := handler.ExtractClaims(evidence, trustAnchors)
	if err != nil {
		if errors.Is(err, fault.ErrBadEvidence) {
			addPolicyClaim(result, "problem", err.Error())
		}
		return o.finalize(result, err)
	}

	referenceValueIDs, err := handler.GetReferenceValueIDs(trustAnchors, claims)
	if err != nil {
		return o.finalize(result, err)
	}

	o.logger.Debugw("constructed evidence context",
		"software-id", referenceValueIDs,
		"trust-anchor-id", trustAnchorIDs)

	var endorsements []*comid.ValueTriple
	o.logger.Debug("obtaining endorsements...")
	endorsements, err = o.getValueTriples(referenceValueIDs, storeLabel, true)
	if err != nil {
		if errors.Is(err, fault.ErrNotFound) {
			o.logger.Warn("no reference values in store")
		} else {
			return o.finalize(result, err)
		}
	}

	o.logger.Debug("validating evidence...")
	if err = handler.ValidateEvidenceIntegrity(evidence, trustAnchors, endorsements); err != nil {
		if errors.Is(err, fault.ErrBadEvidence) {
			setAllClaims(result, ear.CryptoValidationFailedClaim)
			addPolicyClaim(result, "problem",
				fmt.Sprintf("integrity validation failed: %s", err.Error()))
		}

		return o.finalize(result, err)
	}

	o.logger.Debug("appraising claims...")
	appraisedResult, err := handler.AppraiseClaims(claims, endorsements)
	if err != nil {
		return o.finalize(result, err)
	}
	appraisedResult.Nonce = result.Nonce
	result = appraisedResult
	initPolicyID(result, schemeName)

	o.logger.Debug("evaluating policy...")
	err = o.PolicyManager.Evaluate(ctx, tenantID, schemeName, claims, result, endorsements)
	if err != nil {
		return o.finalize(result, err)
	}

	o.logger.Infow("evaluated attestation result", "attestation-result", result)
	return o.finalize(result, nil)
}

func (o *TrustedService) GetEndorsements(
	ctx context.Context,
	mediaType string,
	queryText string,
) ([]byte, error) {
	o.logger.Infow("GetEndorsements", "media-type", mediaType)

	var query coserv.Coserv
	if err := query.FromBase64Url(queryText); err != nil {
		return nil, err
	}

	rootMediaType, mediaTypeParams, err := mime.ParseMediaType(mediaType)
	if err != nil {
		o.logger.Warnf("could not parse media type: %s", err.Error())
		return nil, err
	}
	profile := mediaTypeParams["profile"]

	shouldSign := false
	switch rootMediaType {
	case "application/coserv+cose":
		shouldSign = true
	case "application/coserv+cbor":
		// no-op
	default:
		message := fmt.Sprintf("unxpected media type: %s", mediaType)
		o.logger.Warn(message)
		return nil, errors.New(message)
	}

	result, err := o.EndorsementStore.ExecuteCoservQuery(profile, queryText, shouldSign)
	if err != nil {
		return nil, err
	}

	o.logger.Infow("endorsements obtained", "result", result)
	return result, nil
}

func (o *TrustedService) SetEndorsementsState(
	ctx context.Context,
	data []byte,
	isActive bool,
) error {
	tenantID := vts.GetTenantID(ctx)
	o.logger.Infow("SetEndorsementState", "tenantID", tenantID, "data", data, "isActive", isActive)
	if err := o.EndorsementStore.SetEndorsementsState(tenantID, data, isActive); err != nil {
		return err
	}
	o.logger.Info("endorsement state set")
	return nil
}

func (o *TrustedService) decodeAndValidateSignedCorim(data []byte) (*corim.UnsignedCorim, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty corim data")
	}

	o.logger.Debug("unmarshaling signed CoRIM and extrating x5chain")
	sc, err := corim.UnmarshalAndValidateSignedCorimFromCBOR(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse signed CoRIM: %w", err)
	}

	if sc.SigningCert == nil {
		return nil, fmt.Errorf("no signing certificate found in the CoRIM")
	}

	intermediateCertPool := x509.NewCertPool()
	for _, cert := range sc.IntermediateCerts {
		intermediateCertPool.AddCert(cert)
	}

	verifyOpts := x509.VerifyOptions{
		Roots:         o.RootCerts,
		Intermediates: intermediateCertPool,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	o.logger.Debug("verifying key usage")
	hasCodeSigningEKU := false
	for _, eku := range sc.SigningCert.ExtKeyUsage {
		if eku == x509.ExtKeyUsageCodeSigning {
			hasCodeSigningEKU = true
			break
		}
	}

	if !hasCodeSigningEKU {
		return nil, errors.New("certificate does not have codeSigning in its Extended Key Usage")
	}

	o.logger.Debug("verifying certificate chain")
	_, err = sc.SigningCert.Verify(verifyOpts)
	if err != nil {
		return nil, fmt.Errorf("certificate chain verification failed: %w", err)
	}

	o.logger.Debug("verifying signature")
	if err := sc.Verify(sc.SigningCert.PublicKey); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	o.logger.Debug("CoRIM validated")
	return &sc.UnsignedCorim, nil
}

func (o *TrustedService) decodeAndValidateUnsignedCorim(data []byte) (*corim.UnsignedCorim, error) {
	o.logger.Debug("unmarshaling unsigned CoRIM")
	ret, err := corim.UnmarshalAndValidateUnsignedCorimFromCBOR(data)
	if err != nil {
		return nil, err
	}

	o.logger.Debug("CoRIM validated")
	return ret, nil
}

func (o *TrustedService) getKeyTriples(
	trustAnchorIDs []*comid.Environment,
	label string,
	exact bool,
) ([]*comid.KeyTriple, error) {

	var keyTriples []*comid.KeyTriple //nolint

	for _, taID := range trustAnchorIDs {
		triples, err := o.EndorsementStore.GetKeyTriples(taID, label, exact)
		if err != nil {
			o.logger.Warnw("could not find in store", "taID", taID, "error", err)
			return nil, err
		}
		keyTriples = append(keyTriples, triples...)
	}

	if len(keyTriples) == 0 {
		return nil, fault.ErrNotFound
	}

	return keyTriples, nil
}

func (o *TrustedService) getValueTriples(
	referenceValueIDs []*comid.Environment,
	label string,
	exact bool,
) ([]*comid.ValueTriple, error) {

	var valueTriples []*comid.ValueTriple //nolint

	for _, valID := range referenceValueIDs {
		triples, err := o.EndorsementStore.GetValueTriples(valID, label, exact)
		if err != nil && !errors.Is(err, fault.ErrNotFound) {
			o.logger.Warnw("could not find in store", "valID", valID, "error", err)
			return nil, err
		}
		valueTriples = append(valueTriples, triples...)
	}

	if len(valueTriples) == 0 {
		return nil, fault.ErrNotFound
	}

	return valueTriples, nil
}

func (o *TrustedService) getSupportedVerificationMediaTypes() ([]string, error) {
	o.logger.Debug("GetSupportedProvisioningMediaTypes")
	ret := o.SchemeManager.GetRegisteredEvidenceMediaTypes()
	o.logger.Debugw("sending media types", "media-types", ret)
	return ret, nil
}

func (o *TrustedService) getSupportedProvisioningMediaTypes() ([]string, error) {
	o.logger.Debug("GetSupportedProvisioningMediaTypes")
	corimProfiles := o.SchemeManager.GetRegisteredCorimProfiles()
	ret := make([]string, 0, len(corimProfiles)*2)

	for _, profile := range corimProfiles {
		ret = append(ret,
			fmt.Sprintf(`application/rim+cbor; profile=%q`, profile),
			fmt.Sprintf(`application/rim+cose; profile=%q`, profile),
		)
	}

	o.logger.Debugw("sending media types", "media-types", ret)
	return ret, nil
}

func (o *TrustedService) getSupportedCoserMediaTypes() ([]string, error) {
	o.logger.Debug("GetSupportedCoservMediaTypes")
	corimProfiles := o.SchemeManager.GetRegisteredCorimProfiles()
	ret := make([]string, 0, len(corimProfiles)*2)

	for _, profile := range corimProfiles {
		ret = append(ret,
			fmt.Sprintf(`application/coserv+cbor; profile=%q`, profile),
			fmt.Sprintf(`application/coserv+cose; profile=%q`, profile),
		)
	}

	o.logger.Debugw("sending media types", "media-types", ret)
	return ret, nil
}

func (o *TrustedService) getEARSigningPublicKey() (string, error) {
	o.logger.Debug("GetEARSigningPublicKey")
	alg, key, err := o.EarSigner.GetEARSigningPublicKey()
	if err != nil {
		return "", err
	}

	err = key.Set("alg", alg.String())
	if err != nil {
		return "", err
	}

	b, err := json.Marshal(key)
	if err != nil {
		return "", err
	}

	o.logger.Debugw("sending key", "key", string(b))
	return string(b), nil
}

func (o *TrustedService) getCoservSigningPublicKey() (string, error) {
	o.logger.Debug("GetCoservSigningPublicKey")
	alg, key, err := o.EndorsementStore.CoservSigner().GetCoservSigningPublicKey()
	if err != nil {
		return "", err
	}

	err = key.Set("alg", alg.String())
	if err != nil {
		return "", err
	}

	b, err := json.Marshal(key)
	if err != nil {
		return "", err
	}

	o.logger.Debugw("sending key", "key", string(b))
	return string(b), nil
}

// finalize prepares the final appraisal context to be returned to the client.
// The EAR is signed using the verifier private key.
//
// The error parameter indicates whether there was an error during the
// attestation process. If a non-nil error is supplied, it is classified as a
// verifier malfunction - unless it's of type "bad evidence", in which case it
// is logged and the error is cleared because we assume the relevant claim has
// been already set in the attestation result.
func (o *TrustedService) finalize(
	result *ear.AttestationResult,
	err error,
) ([]byte, error) {
	if err != nil {
		if errors.Is(err, fault.ErrBadEvidence) {
			o.logger.Warn(err)
			// Clear the error as we've "handled" by setting the
			// claim in the result.
			err = nil
		} else {
			o.logger.Error(err)
			setAllClaims(result, ear.VerifierMalfunctionClaim)
		}

	}

	result.UpdateStatusFromTrustVector()

	signedEAR, signErr := o.EarSigner.Sign(*result)
	if signErr != nil {
		o.logger.Error(fmt.Errorf("failed to sign EAR: %w", signErr))
		// Signing error overrides whatever the problem that got us
		// here was, as it indicates a serious issue with the service.
		// TODO(setrofim): signing an EAR should be an error-free
		// operation. It should either succeed or panic if there is a
		// serious issue with the underlying platform (such as OOM).
		// Any other problems (e.g. bad/missing key) should be
		// identified and handled during service initialisation.
		err = signErr
	}

	return signedEAR, err
}

func setAllClaims(result *ear.AttestationResult, claim ear.TrustClaim) {
	for _, submod := range result.Submods {
		submod.TrustVector.SetAll(claim)
	}
}

func addPolicyClaim(result *ear.AttestationResult, name, claim string) {
	for _, submod := range result.Submods {
		if submod.VeraisonPolicyClaims == nil {
			claimsMap := make(map[string]any)
			submod.VeraisonPolicyClaims = &claimsMap
		}
		(*submod.VeraisonPolicyClaims)[name] = claim
	}
}

func setScheme(result *ear.AttestationResult, schemeName string) error {
	var ok bool
	if _, ok = result.Submods[schemeName]; ok {
		return fmt.Errorf("submod %q already exists in result", schemeName)
	}

	// now that the scheme is known, move the default submod (which is
	// currently under "ERROR") to under the scheme's name.
	result.Submods[schemeName], ok = result.Submods["ERROR"]
	if !ok {
		return errors.New("submod \"ERROR\" not in result; has the scheme already been set?")
	}
	delete(result.Submods, "ERROR")

	initPolicyID(result, schemeName)

	return nil
}

func initPolicyID(result *ear.AttestationResult, schemeName string) {
	policyID := fmt.Sprintf("policy:%s", schemeName)
	for _, submod := range result.Submods {
		submod.AppraisalPolicyID = &policyID
	}
}

func toJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}

	return string(encoded)
}
