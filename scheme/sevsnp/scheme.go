// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sevsnp

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"

	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/kds"
	"github.com/google/go-sev-guest/proto/sevsnp"
	sevsnpParser "github.com/jraman567/go-gen-ref/cmd/sevsnp"
	"github.com/veraison/cmw"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/ear"
	"github.com/veraison/ratsd/tokens"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/scheme/common"
	"github.com/veraison/services/vts/appraisal"
	"go.uber.org/zap"
)

var (
	ErrCertificateReadFailure = errors.New("failed to read certificate")
	ErrMissingCMW             = errors.New("CMW not found in evidence token")
	ErrMissingCertChain       = errors.New("evidence missing certificate chain")
	ErrNoARK                  = errors.New("missing ARK certificate in evidence")
	ErrNoASK                  = errors.New("missing ASK certificate in evidence")
	ErrNoVEK                  = errors.New("evidence must supply VLEK or VCEK")
	ErrTAMismatch             = errors.New("evidence Trust Anchor (ARK) doesn't match the provisioned one")

	EndorsementMediaTypeRV   = `application/corim-unsigned+cbor; profile="tag:amd.com,2024:snp-corim-profile"`
	EvidenceMediaTypeRATSd   = `application/eat+cwt; eat_profile="tag:github.com,2025:veraison/ratsd/cmw"`
	EvidenceMediaTypeTSMCbor = "application/vnd.veraison.tsm-report+cbor"
	EvidenceMediaTypeTSMJson = "application/vnd.veraison.configfs-tsm+json"
)

const (
	mKeyPolicy           = 2
	mKeyCurrentTcb       = 6
	mKeyPlatformInfo     = 7
	mKeyReportData       = 640
	mKeyMeasurement      = 641
	mKeyReportID         = 645
	mKeyReportIDMA       = 646
	mKeyReportedTcb      = 647
	mKeyChipID           = 3328
	mKeyCommittedTcb     = 3329
	mKeyCurrentVersion   = 3330
	mKeyCommittedVersion = 3936
	mKeyLaunchTcb        = 3968
)

var Descriptor = handler.SchemeDescriptor{
	Name:         "SEVSNP",
	VersionMajor: 1,
	VersionMinor: 0,
	CorimProfiles: []string{
		ProfileString,
		ArkProfileString,
	},
	EvidenceMediaTypes: []string{
		EvidenceMediaTypeTSMCbor,
		EvidenceMediaTypeTSMJson,
		EvidenceMediaTypeRATSd,
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
	tsm, err := parseEvidence(evidence)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	certChain, err := parseCertChainFromTSMReport(tsm)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	ark, err := readCert(certChain.GetArkCert())
	if err != nil {
		return nil, handler.BadEvidence("can't read ARK to compose TA ID: %w", err)
	}

	cert, err := x509.ParseCertificate(ark)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	return []*comid.Environment{
		{
			Class: &comid.Class{
				Vendor: &cert.Subject.Organization[0],
				Model:  &cert.Subject.CommonName,
			},
		},
	}, nil
}

func (o *Implementation) GetReferenceValueIDs(
	trustAnchors []*comid.KeyTriple,
	claims map[string]any,
) ([]*comid.Environment, error) {
	evCorim, err := transformClaimsToCorim(claims)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	var ret []*comid.Environment // nolint:prealloc
	rvIter, iterErr := evCorim.IterRefVals()
	for refVal := range rvIter {
		ret = append(ret, &refVal.Environment)
	}
	if err := iterErr(); err != nil {
		return nil, handler.BadEvidence(err)
	}

	return ret, nil
}

func (o *Implementation) ExtractClaims(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
) (map[string]any, error) {
	evCoRIM, err := transformEvidenceToCorim(evidence)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	return common.ToMapViaJSON(evCoRIM)
}

func (o *Implementation) ValidateEvidenceIntegrity(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
	endorsements []*comid.ValueTriple,
) error {
	taCert, err := getCertFromTrustAnchors(trustAnchors)
	if err != nil {
		return err
	}

	tsm, err := parseEvidence(evidence)
	if err != nil {
		return handler.BadEvidence(err)
	}

	certChain, err := parseCertChainFromTSMReport(tsm)
	if err != nil {
		return handler.BadEvidence(err)
	}

	if err := validateCertChain(certChain); err != nil {
		return handler.BadEvidence(err)
	}

	if !bytes.Equal(certChain.GetArkCert(), taCert) {
		return handler.BadEvidence(ErrTAMismatch)
	}

	return nil
}

func (o *Implementation) AppraiseClaims(
	claims map[string]any,
	endorsements []*comid.ValueTriple,
) (*ear.AttestationResult, error) {
	result := handler.CreateAttestationResult(Descriptor.Name)
	appraisal := result.Submods[Descriptor.Name]

	appraisal.TrustVector.Hardware = ear.UnsafeHardwareClaim
	appraisal.TrustVector.RuntimeOpaque = ear.VisibleMemoryRuntimeClaim

	evMeasurements, err := transformClaimsToMeasurementsMap(claims)
	if err != nil {
		return result, handler.BadEvidence(err)
	}

	matched := false
	for i, endorsement := range endorsements {
		o.logger.Debugf("attempting to match endorsement %d...", i)
		refMeasurements, err := transformValueTripleToMeasurementsMap(endorsement)
		if err != nil {
			return result, err
		}

		if tryMatchEvidence(o.logger, evMeasurements, refMeasurements) {
			matched = true
			break
		}
	}

	if matched {
		o.logger.Debug("success!")
		appraisal.TrustVector.Hardware = ear.GenuineHardwareClaim
		appraisal.TrustVector.RuntimeOpaque = ear.EncryptedMemoryRuntimeClaim
	} else {
		o.logger.Debug("failed to match evidence to reference values!")
	}

	appraisal.UpdateStatusFromTrustVector()
	appraisal.VeraisonAnnotatedEvidence = &claims

	return result, nil
}

func parseEvidence(evidence *appraisal.Evidence) (*tokens.TSMReport, error) {
	var (
		err           error
		tsm           = new(tokens.TSMReport)
		cmwCollection cmw.CMW
	)

	switch evidence.MediaType {
	case EvidenceMediaTypeTSMCbor:
		err = tsm.FromCBOR(evidence.Data)
		if err != nil {
			return nil, err
		}
	case EvidenceMediaTypeTSMJson:
		err = tsm.FromJSON(evidence.Data)
		if err != nil {
			return nil, err
		}
	case EvidenceMediaTypeRATSd:
		eat := make(map[string]any)

		err = json.Unmarshal(evidence.Data, &eat)
		if err != nil {
			return nil, err
		}

		cmwBase64, ok := eat["cmw"].(string)
		if !ok {
			return nil, handler.BadEvidence(ErrMissingCMW)
		}

		cmwJson, err := base64.StdEncoding.DecodeString(cmwBase64)
		if err != nil {
			return nil, err
		}

		err = cmwCollection.UnmarshalJSON(cmwJson)
		if err != nil {
			return nil, err
		}

		cmwMonad, err := cmwCollection.GetCollectionItem("tsm-report")
		if err != nil {
			return nil, err
		}

		cmwType, err := cmwMonad.GetMonadType()
		if err != nil {
			return nil, err
		}
		if cmwType != EvidenceMediaTypeTSMJson {
			return nil, fmt.Errorf("unexpected CMW type: %s", cmwType)
		}
		cmwValue, err := cmwMonad.GetMonadValue()
		if err != nil {
			return nil, err
		}

		err = tsm.FromJSON(cmwValue)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unexpected media type: %s", evidence.MediaType)
	}

	return tsm, nil
}

func readCert(cert []byte) ([]byte, error) {
	if len(cert) == 0 {
		return nil, errors.New("empty certificate")
	}

	block, _ := pem.Decode(cert)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, ErrCertificateReadFailure
	}
	return block.Bytes, nil
}

func parseCertChainFromTSMReport(tsm *tokens.TSMReport) (*sevsnp.CertificateChain, error) {
	var certTable abi.CertTable

	if len(tsm.AuxBlob) == 0 {
		return nil, ErrMissingCertChain
	}

	if err := certTable.Unmarshal(tsm.AuxBlob); err != nil {
		return nil, err
	}

	return certTable.Proto(), nil
}

func transformClaimsToCorim(claims map[string]any) (*corim.UnsignedCorim, error) {
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return nil, err
	}

	var ret corim.UnsignedCorim
	if err := ret.FromJSON(claimsJSON); err != nil {
		return nil, err
	}

	return &ret, nil
}

func transformEvidenceToComid(evidence *appraisal.Evidence) (*comid.Comid, error) {
	tsm, err := parseEvidence(evidence)
	if err != nil {
		return nil, err
	}

	reportProto, err := abi.ReportToProto(tsm.OutBlob)
	if err != nil {
		return nil, err
	}

	evComid, err := sevsnpParser.ReportToComid(reportProto, 0)
	if err != nil {
		return nil, err
	}

	err = evComid.Valid()
	if err != nil {
		return nil, err
	}

	return evComid, err
}

func transformEvidenceToCorim(evidence *appraisal.Evidence) (*corim.UnsignedCorim, error) {
	evComid, err := transformEvidenceToComid(evidence)
	if err != nil {
		return nil, err
	}

	evCorim := corim.UnsignedCorim{}
	evCorim.SetProfile(EndorsementMediaTypeRV)
	evCorim.AddComid(evComid)

	return &evCorim, nil
}

func getCertFromTrustAnchors(trustAnchors []*comid.KeyTriple) ([]byte, error) {
	vk, err := common.ExtractOneVerifKey(trustAnchors)
	if err != nil {
		return nil, err
	}

	if vk.Type() != comid.PKIXBase64CertType {
		return nil, fmt.Errorf("wrong trust anchor: expected %s, found %s",
			comid.PKIXBase64CertType, vk.Type())
	}

	return []byte(vk.String()), nil
}

func validateCertChain(certChain *sevsnp.CertificateChain) error {
	if len(certChain.GetArkCert()) == 0 {
		return handler.BadEvidence(ErrNoARK)
	}

	if len(certChain.GetAskCert()) == 0 {
		return handler.BadEvidence(ErrNoASK)
	}

	if len(certChain.GetVcekCert()) == 0 && len(certChain.GetVlekCert()) == 0 {
		return handler.BadEvidence(ErrNoVEK)
	}

	return nil
}

func transformClaimsToMeasurementsMap(claims map[string]any) (map[uint64]comid.Measurement, error) {
	evCorim, err := transformClaimsToCorim(claims)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	vtIter, iterErr := evCorim.IterRefVals()
	valueTriples := slices.Collect(vtIter)
	if err := iterErr(); err != nil {
		return nil, err
	}

	if numValueTriples := len(valueTriples); numValueTriples != 1 {
		return nil, fmt.Errorf("expected exactly one triple in evidence; found %d", numValueTriples)
	}

	return transformValueTripleToMeasurementsMap(valueTriples[0])
}

func transformValueTripleToMeasurementsMap(vt *comid.ValueTriple) (map[uint64]comid.Measurement, error) {
	ret := make(map[uint64]comid.Measurement)

	for _, measurement := range vt.Measurements.Values {
		key, err := measurement.Key.GetKeyUint()
		if err != nil {
			return nil, err
		}

		ret[key] = measurement
	}

	return ret, nil
}

func tryMatchEvidence(
	logger *zap.SugaredLogger,
	evMeasurements, refMeasurements map[uint64]comid.Measurement,
) bool {
	for key, refMeasurement := range refMeasurements {
		// We can skip validating certain claims for the following reasons:
		// - POLICY ToDo: Do we need to test individual policy features?
		// - CURRENT_TCB is informational only. It's best handled by policy
		// - PLATFORM_INFO ToDO: Do we need to test individual platform features?
		// - REPORT_DATA is a nonce supplied by user for freshness. It's used
		//       for freshness verification, and verified as part of
		//       evidence integrity check (session nonce check).
		// - REPORT_ID is ephemeral, so we can't use it for verification.
		// - REPORT_ID_MA is also ephemeral, used for migration
		// - CHIP_ID is unique to an specific attester, but reference values could be used more generally
		// - Current Version (CURRENT_MAJOR/MINOR/BUILD) should already be part of REPORTED_TCB.
		//     ToDo: It is a good idea to test it anyway, but the Version type only tests for
		//     equality, and this would trigger spurious failures
		// - COMMITTED_TCB is informational, used by the host to advance REPORTED_TCB
		if key == mKeyPolicy ||
			key == mKeyCurrentTcb ||
			key == mKeyPlatformInfo ||
			key == mKeyReportData ||
			key == mKeyReportID ||
			key == mKeyReportIDMA ||
			key == mKeyChipID ||
			key == mKeyCommittedTcb ||
			key == mKeyCurrentVersion ||
			key == mKeyCommittedVersion {
			continue
		}

		evMeasurement, ok := evMeasurements[key]
		if !ok {
			logger.Debugf("key %d not in evidence", key)
			return false
		}

		switch key {
		case mKeyReportedTcb:
			if !compareTcb(logger, refMeasurement, evMeasurement) {
				logger.Debugf("reported TCB (key %d) failed to match", mKeyReportedTcb)
				return false
			}
		case mKeyLaunchTcb:
			evReportedTcb, ok := evMeasurements[mKeyReportedTcb]
			if !ok {
				logger.Debugf("key %d not in evidence", mKeyReportedTcb)
				return false
			}

			if !compareTcb(logger, refMeasurement, evReportedTcb) {
				// TODO: Is this a failure condition?
				log.Debug("TEE launched with older TCB version")
			}
		default:
			if !compareMeasurements(logger, refMeasurement, evMeasurement) {
				logger.Debugf("MKey %d does not match reference", key)
				return false
			}
		}

	}

	return true
}

func compareTcb(logger *zap.SugaredLogger, refM comid.Measurement, evM comid.Measurement) bool {
	if refM.Val.SVN == nil {
		logger.Debug("reference doesn't have SVN")
		return false
	}

	if evM.Val.SVN == nil {
		logger.Debug("evidence doesn't have SVN")
		return false
	}

	refTcbParts, err := transformSVNtoTCB(*refM.Val.SVN)
	if err != nil {
		logger.Debugf("could not transform reference SVN to TCB parts: %v", err)
		return false
	}

	evTcbParts, err := transformSVNtoTCB(*evM.Val.SVN)
	if err != nil {
		logger.Debugf("could not transform evidence SVN to TCB parts: %v", err)
		return false
	}

	if evTcbParts.BlSpl < refTcbParts.BlSpl ||
		evTcbParts.SnpSpl < refTcbParts.SnpSpl ||
		evTcbParts.TeeSpl < refTcbParts.TeeSpl ||
		evTcbParts.UcodeSpl < refTcbParts.UcodeSpl {
		return false
	}

	return true
}

// transformSVNtoTCB extracts TCB from the supplied SVN. SEV-SNP's TCB_VERSION
// is a composite version; it's bitfield consisting of SVNs from various firmware components
func transformSVNtoTCB(svn comid.SVN) (*kds.TCBParts, error) {
	var (
		tcbVersion uint64
		err        error
		tcbParts   kds.TCBParts
	)

	// ToDo: following is a circuitous way to obtain the 64-bit TCB integer value
	// from SVN. Consider updating the SVN type to return a 64-bit value
	switch v := svn.Value.(type) {
	case *comid.TaggedSVN:
		tcbString := v.String()
		tcbVersion, err = strconv.ParseUint(tcbString, 10, 64)
	case *comid.TaggedMinSVN:
		tcbString := v.String()
		tcbVersion, err = strconv.ParseUint(tcbString, 10, 64)
	default:
		err = fmt.Errorf("unsupported SVN type: %v", reflect.TypeOf(svn.Value))
	}

	if err != nil {
		return nil, err
	}

	tcbParts = kds.DecomposeTCBVersion(kds.TCBVersion(tcbVersion))

	return &tcbParts, nil
}

// compareMeasurements checks if two given comid.Measurement variables are equal.
func compareMeasurements(logger *zap.SugaredLogger, refM comid.Measurement, evM comid.Measurement) bool {
	// RawValue comparison
	if refM.Val.RawValue != nil {
		if evM.Val.RawValue == nil {
			return false
		}

		refDigest, _ := refM.Val.RawValue.GetBytes()
		return evM.Val.RawValue.CompareAgainstReference(refDigest, nil)
	}

	// Digests comparison
	if refM.Val.Digests != nil {
		if evM.Val.Digests == nil {
			return false
		}

		return evM.Val.Digests.CompareAgainstReference(*refM.Val.Digests)
	}

	// SVN comparison
	if refM.Val.SVN != nil {
		if evM.Val.SVN == nil {
			logger.Debug("evidence doesn't have SVN")
			return false
		}

		if c, ok := evM.Val.SVN.Value.(*comid.TaggedSVN); ok {
			if r, ok := refM.Val.SVN.Value.(*comid.TaggedSVN); ok {
				return c.CompareAgainstRefSVN(*r)
			} else if r, ok := refM.Val.SVN.Value.(*comid.TaggedMinSVN); ok {
				return c.CompareAgainstRefMinSVN(*r)
			} else {
				logger.Debug("unknown refVal SVN type")
				return false
			}
		} else if c, ok := evM.Val.SVN.Value.(*comid.TaggedMinSVN); ok {
			if r, ok := refM.Val.SVN.Value.(*comid.TaggedMinSVN); ok {
				return c.Equal(*r)
			}
			logger.Debug("can't compare TaggedMinSVN against TaggedSVN")
			return false
		} else {
			logger.Debug("unknown evidence SVN type")
			return false
		}
	}

	return true
}
