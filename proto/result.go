// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package proto

import (
	"fmt"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

// Claim values as defined by
// https://datatracker.ietf.org/doc/draft-ietf-rats-ar4si/
const (
	// common to all claims

	// The Evidence received is insufficient to make a conclusion.  Note:
	// this should always be always treated equivalently by the Relying
	// Party as no claim being made.  I.e., the RP's Appraisal Policy for
	// Attestation Results SHOULD NOT make any distinction between a
	// Trustworthiness Claim with enumeration '0', and no Trustworthiness
	// Claim being provided.
	ARStatus_NO_CLAIM ARStatus = 0
	// The Evidence received contains unexpected elements which the
	// Verifier is unable to parse.  An example might be that the wrong
	// type of Evidence has been delivered.
	ARStatus_CANNOT_PARSE ARStatus = 1
	// A verifier malfunction occurred during the Verifier's appraisal
	// processing.
	ARStatus_VERIFIER_ERROR ARStatus = -1
	// Cryptographic validation of the Evidence has failed.
	ARStatus_INVALID ARStatus = 99

	// NOTE(setrofim): While AFFIRMING and CONTRAINDICATED values "2" and
	// "96" are shared in common between all claims, their specific
	// interpretation is claim-dependent. As such, I decided to preserve
	// this by defining separate consts for each claim, along with
	// associated explanatory comments.
	// Contrast this with the INVALID value "99", which, though
	// separately for each claim, has identical interpretations for all of
	// them.

	// configuration:  A Verifier has appraised an Attester's configuration,
	// and is able to make conclusions regarding the exposure of known
	// vulnerabilities.

	// The configuration is a known and approved config.
	ARStatus_CONF_AFFIRMING ARStatus = 2
	// The configuration includes or exposes no known vulnerabilities.
	ARStatus_CONF_NOVULN ARStatus = 3
	// The configuration includes or exposes known vulnerabilities.
	ARStatus_CONF_UNSAFE ARStatus = 32
	// The configuration is unsupportable as it exposes unacceptable
	// security vulnerabilities.
	ARStatus_CONF_UNSUPPORTABLE ARStatus = 96

	// executables:  A Verifier has appraised and evaluated relevant
	// runtime files, scripts, and/or other objects which have been
	// loaded into the Target environment's memory.

	// Only a recognized genuine set of approved executables, scripts,
	// files, and/or objects have been loaded during and after the boot
	// process.
	ARStatus_EXE_AFFIRMING ARStatus = 2
	// Only a recognized genuine set of approved executables have been
	// loaded during the boot process.
	ARStatus_EXE_BOOT_AFFIRMING ARStatus = 3
	// Only a recognized genuine set of executables, scripts, files, and/or
	// objects have been loaded.  However the Verifier cannot vouch for a
	// subset of these due to known bugs or other known vulnerabilities.
	ARStatus_EXE_UNSAFE ARStatus = 32
	// Runtime memory includes executables, scripts, files, and/or objects
	// which are not recognized.
	ARStatus_EXE_UNRECOGNIZED ARStatus = 33
	// Runtime memory includes executables, scripts, files, and/or object
	// which are contraindicated.
	ARStatus_EXE_CONTRAINDICATED ARStatus = 96

	// file-system:  A Verifier has evaluated a specific set of
	// directories within the Attester's file system.  (Note: the
	// Verifier may or may not indicate what these directory and
	// expected files are via an unspecified management interface.)

	// Only a recognized set of approved files are found.
	ARStatus_FS_AFFIRMING ARStatus = 2
	// The file system includes unrecognized executables, scripts, or
	// files.
	ARStatus_FS_UNSAFE ARStatus = 32
	// The file system includes contraindicated executables, scripts, or
	// files.
	ARStatus_FS_COUNTRAINDICATED ARStatus = 96

	// hardware:  A Verifier has appraised any Attester hardware and
	// firmware which are able to expose fingerprints of their identity and
	// running code.

	// An Attester has passed its hardware and/or firmware verifications
	// needed to demonstrate that these are genuine/ supported.
	ARStatus_HW_AFFIRMING ARStatus = 2
	//An Attester contains only genuine/supported hardware and/or firmware,
	//but there are known security vulnerabilities.
	ARStatus_HW_UNSAFE ARStatus = 32
	// Attester hardware and/or firmware is recognized, but its
	// trustworthiness is contraindicated.
	ARStatus_HW_CONTRAINDICATED ARStatus = 96
	// A Verifier does not recognize an Attester's hardware or firmware,
	// but it should be recognized.
	ARStatus_HW_UNRECOGNIZED ARStatus = 97

	// instance-identity:  A Verifier has appraised an Attesting
	// Environment's unique identity based upon private key signed Evidence
	// which can be correlated to a unique instantiated instance of the
	// Attester.  (Note: this Trustworthiness Claim should only be
	// generated if the Verifier actually expects to recognize the unique
	// identity of the Attester.)

	// The Attesting Environment is recognized, and the associated instance
	// of the Attester is not known to be compromised.
	ARStatus_IDENT_AFFIRMING ARStatus = 2
	// The Attesting Environment is recognized, and but its unique private
	// key indicates a device which is not trustworthy.
	ARStatus_IDENT_CONTRAINDICATED ARStatus = 96
	// The Attesting Environment is not recognized; however the Verifier
	// believes it should be.
	ARStatus_IDENT_UNRECOGNIZED ARStatus = 97

	// runtime-opaque:  A Verifier has appraised the visibility of Attester
	// objects in memory from perspectives outside the Attester.

	// the Attester's executing Target Environment and Attesting
	// Environments are encrypted and within Trusted Execution
	// Environment(s) opaque to the operating system, virtual machine
	// manager, and peer applications.  (Note: This value corresponds to
	// the protections asserted by O.RUNTIME_CONFIDENTIALITY from
	// [GP-TEE-PP])
	ARStatus_RT_AFFIRMING ARStatus = 2
	// the Attester's executing Target Environment and Attesting
	// Environments inaccessible from any other parallel application or
	// Guest VM running on the Attester's physical device.  (Note that
	// unlike ARStatus_RT_AFFIRMINGthese environments are not encrypted in
	// a way which restricts the Attester's root operator visibility.  See
	// O.TA_ISOLATION from [GP-TEE-PP].)
	ARStatus_RT_ISOLATED ARStatus = 32
	// The Verifier has concluded that in memory objects are unacceptably
	// visible within the physical host that supports the Attester.
	ARStatus_RT_EXPOSED ARStatus = 96

	// sourced-data:  A Verifier has evaluated of the integrity of data
	// objects from external systems used by the Attester.

	// All essential Attester source data objects have been provided by
	// other Attester(s) whose most recent appraisal(s) had both no
	// Trustworthiness Claims of "0" where the current Trustworthiness
	// Claim is "Affirming", as well as no "Warning" or "Contraindicated"
	// Trustworthiness Claims.
	ARStatus_SOURCED_AFFIRMING ARStatus = 2
	// Attester source data objects come from unattested sources, or
	// attested sources with "Warning" type Trustworthiness Claims.
	ARStatus_SOURCED_UNSAFE ARStatus = 32
	// Attester source data objects come from contraindicated sources.
	ARStatus_SOURCED_CONTRAINDICATED ARStatus = 96

	// storage-opaque:  A Verifier has appraised that an Attester is
	// capable of encrypting persistent storage.  (Note: Protections must
	// meet the capabilities of [OMTP-ATE] Section 5, but need not be
	// hardware tamper resistant.)

	// the Attester encrypts all secrets in persistent storage via using
	// keys which are never visible outside an HSM or the Trusted Execution
	// Environment hardware.
	ARStatus_SECRETS_AFFIRMING ARStatus = 2
	// the Attester encrypts all persistently stored secrets, but without
	// using hardware backed keys
	ARStatus_SECRETS_NOHWKEYS ARStatus = 32
	// There are persistent secrets which are stored unencrypted in an
	// Attester.
	ARStatus_SECRETS_EXPOSED ARStatus = 96
)

type ARStatus int8

func Int32ToStatus(i int32) (ARStatus, error) {
	if i > 127 || i < -128 {
		return 0, fmt.Errorf("out of range: %d", i)
	}

	return ARStatus(i), nil
}

func Int64ToStatus(i int64) (ARStatus, error) {
	if i > 127 || i < -128 {
		return 0, fmt.Errorf("out of range: %d", i)
	}

	return ARStatus(i), nil
}

func (o ARStatus) GetTier() TrustTier {
	if (o >= 2 && o <= 31) || (o <= -2 && o >= -32) {
		return TrustTier_AFFIRMING
	} else if (o >= 32 && o <= 95) || (o <= -33 && o >= -96) {
		return TrustTier_WARNING
	} else if (o >= 96 && o <= 127) || (o <= -97 && o >= -128) {
		return TrustTier_CONTRAINDICATED
	}

	return TrustTier_NONE
}

func (o ARStatus) Int32() int32 {
	return int32(o)
}

func GetInt32TrustTier(i int32) TrustTier {
	status, err := Int32ToStatus(i)
	if err != nil {
		return TrustTier_NONE
	}

	return status.GetTier()
}

func (o TrustTier) Int32() int32 {
	return int32(o)
}

func NewAttestationResult(ec *EvidenceContext) *AttestationResult {
	return &AttestationResult{
		Status: TrustTier_NONE,
		TrustVector: &TrustVector{
			InstanceIdentity: int32(ARStatus_NO_CLAIM),
			Configuration:    int32(ARStatus_NO_CLAIM),
			Executables:      int32(ARStatus_NO_CLAIM),
			FileSystem:       int32(ARStatus_NO_CLAIM),
			Hardware:         int32(ARStatus_NO_CLAIM),
			RuntimeOpaque:    int32(ARStatus_NO_CLAIM),
			StorageOpaque:    int32(ARStatus_NO_CLAIM),
			SourcedData:      int32(ARStatus_NO_CLAIM),
		},
		Timestamp:         timestamppb.Now(),
		ProcessedEvidence: ec.Evidence,
	}
}

func (o *AttestationResult) GetInstanceIdentityStatus() ARStatus {
	status, err := Int32ToStatus(o.TrustVector.InstanceIdentity)
	if err != nil {
		return ARStatus_VERIFIER_ERROR
	}

	return status
}

func (o *AttestationResult) SetInstanceIdentityStatus(status ARStatus) {
	o.TrustVector.InstanceIdentity = int32(status)
}

func (o *AttestationResult) GetConfigurationStatus() ARStatus {
	status, err := Int32ToStatus(o.TrustVector.Configuration)
	if err != nil {
		return ARStatus_VERIFIER_ERROR
	}

	return status
}

func (o *AttestationResult) SetConfigurationStatus(status ARStatus) {
	o.TrustVector.Configuration = int32(status)
}

func (o *AttestationResult) GetExecutablesStatus() ARStatus {
	status, err := Int32ToStatus(o.TrustVector.Executables)
	if err != nil {
		return ARStatus_VERIFIER_ERROR
	}

	return status
}

func (o *AttestationResult) SetExecutablesStatus(status ARStatus) {
	o.TrustVector.Executables = int32(status)
}

func (o *AttestationResult) GetFileSystemStatus() ARStatus {
	status, err := Int32ToStatus(o.TrustVector.FileSystem)
	if err != nil {
		return ARStatus_VERIFIER_ERROR
	}

	return status
}

func (o *AttestationResult) SetFileSystemStatus(status ARStatus) {
	o.TrustVector.FileSystem = int32(status)
}

func (o *AttestationResult) GetHardwareStatus() ARStatus {
	status, err := Int32ToStatus(o.TrustVector.Hardware)
	if err != nil {
		return ARStatus_VERIFIER_ERROR
	}

	return status
}

func (o *AttestationResult) SetHardwareStatus(status ARStatus) {
	o.TrustVector.Hardware = int32(status)
}

func (o *AttestationResult) GetRuntimeOpaqueStatus() ARStatus {
	status, err := Int32ToStatus(o.TrustVector.RuntimeOpaque)
	if err != nil {
		return ARStatus_VERIFIER_ERROR
	}

	return status
}

func (o *AttestationResult) SetRuntimeOpaqueStatus(status ARStatus) {
	o.TrustVector.RuntimeOpaque = int32(status)
}

func (o *AttestationResult) GetStorageOpaqueStatus() ARStatus {
	status, err := Int32ToStatus(o.TrustVector.StorageOpaque)
	if err != nil {
		return ARStatus_VERIFIER_ERROR
	}

	return status
}

func (o *AttestationResult) SetStorageOpaqueStatus(status ARStatus) {
	o.TrustVector.StorageOpaque = int32(status)
}

func (o *AttestationResult) GetSourcedDataStatus() ARStatus {
	status, err := Int32ToStatus(o.TrustVector.SourcedData)
	if err != nil {
		return ARStatus_VERIFIER_ERROR
	}

	return status
}

func (o *AttestationResult) SetSourcedDataStatus(status ARStatus) {
	o.TrustVector.SourcedData = int32(status)
}

func (o *AttestationResult) SetVerifierError() {
	o.Status = TrustTier_NONE
	o.SetInstanceIdentityStatus(ARStatus_VERIFIER_ERROR)
	o.SetConfigurationStatus(ARStatus_VERIFIER_ERROR)
	o.SetExecutablesStatus(ARStatus_VERIFIER_ERROR)
	o.SetFileSystemStatus(ARStatus_VERIFIER_ERROR)
	o.SetHardwareStatus(ARStatus_VERIFIER_ERROR)
	o.SetRuntimeOpaqueStatus(ARStatus_VERIFIER_ERROR)
	o.SetStorageOpaqueStatus(ARStatus_VERIFIER_ERROR)
	o.SetSourcedDataStatus(ARStatus_VERIFIER_ERROR)
}

func (o *AttestationResult) GeteTrustVectorTiers() (map[string]TrustTier, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}

	tvTiers := map[string]TrustTier{
		"instance-identity": o.GetInstanceIdentityStatus().GetTier(),
		"configuration":     o.GetConfigurationStatus().GetTier(),
		"executables":       o.GetExecutablesStatus().GetTier(),
		"file-system":       o.GetFileSystemStatus().GetTier(),
		"hardware":          o.GetHardwareStatus().GetTier(),
		"runtime-opaque":    o.GetRuntimeOpaqueStatus().GetTier(),
		"storage-opaque":    o.GetStorageOpaqueStatus().GetTier(),
		"sourced-data":      o.GetSourcedDataStatus().GetTier(),
	}

	return tvTiers, nil
}

func (o *AttestationResult) UpdateOverallStatus() error {
	tvTiers, err := o.GeteTrustVectorTiers()
	if err != nil {
		return err
	}

	newOverallStatus := TrustTier_NONE

	for _, tier := range tvTiers {
		if tier == TrustTier_NONE {
			continue
		}

		if newOverallStatus == TrustTier_NONE {
			newOverallStatus = tier
			continue
		}

		if newOverallStatus < tier {
			newOverallStatus = tier
			continue
		}
	}

	// If the status has been manually set to be higher than what is
	// warranted by the trust vector, we do not want to "downgrade the
	// severity" here...
	if o.Status < newOverallStatus {
		o.Status = newOverallStatus
	}

	return nil
}

func (o *AttestationResult) Validate() error {

	claimVals := map[string]int32{
		"instance-identity": o.TrustVector.InstanceIdentity,
		"configuration":     o.TrustVector.Configuration,
		"executables":       o.TrustVector.Executables,
		"file-system":       o.TrustVector.FileSystem,
		"hardware":          o.TrustVector.Hardware,
		"runtime-opaque":    o.TrustVector.RuntimeOpaque,
		"storage-opaque":    o.TrustVector.StorageOpaque,
		"sourced-data":      o.TrustVector.SourcedData,
	}

	for claim, val := range claimVals {
		_, err := Int32ToStatus(val)
		if err != nil {
			return fmt.Errorf("%s: %w", claim, err)
		}
	}

	return nil
}
