// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package nvidia

/*
#include "nvat_bindings.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type NvatGpuInterface struct {
	rimStore         C.nvat_rim_store_t
	ocspClient       C.nvat_ocsp_client_t
	gpuLocalVerifier C.nvat_gpu_local_verifier_t
	gpuNrasVerifier  C.nvat_gpu_nras_verifier_t
	gpuVerifier      C.nvat_gpu_verifier_t
	httpOptions      C.nvat_http_options_t
	options          NvatOptions
}

func NewNvatGpuInterface(opts NvatOptions) (*NvatGpuInterface, error) {
	n := &NvatGpuInterface{
		options: opts,
	}
	if err := n.initLibrary(); err != nil {
		return nil, err
	}
	return n, nil
}

func (n *NvatGpuInterface) initLibrary() error {
	if status := C.nvat_load_library(); status != 0 {
		return fmt.Errorf("failed to load NVAT library: %s", C.GoString(C.nvat_last_error()))
	}

	var opts C.nvat_sdk_opts_t

	//nolint:gocritic // CGo call; &opts is an output parameter, not a comparison.
	status := C.nvat_sdk_opts_create_dyn(&opts)
	if status != 0 {
		return fmt.Errorf("nvat_sdk_opts_create failed with code: %d", int(status))
	}
	//nolint:gocritic // CGo call; &opts is an input/output parameter, not a comparison.
	defer C.nvat_sdk_opts_free_dyn(&opts)

	status = C.nvat_sdk_init_dyn(opts)
	if status != 0 {
		return fmt.Errorf("nvat_sdk_init failed: %d", int(status))
	}

	//nolint:gocritic // CGo call; &n.httpOptions is an output parameter, not a comparison.
	status = C.nvat_http_options_create_default_dyn(&n.httpOptions)
	if status != C.NVAT_RC_OK {
		return fmt.Errorf("nvat_http_options_create_default failed with code: %d", int(status))
	}

	if n.options.CaTrustFile != "" {
		caPath := C.CString(n.options.CaTrustFile)
		defer C.free(unsafe.Pointer(caPath))
		//nolint:gocritic // CGo call; &n.httpOptions is an input/output parameter, not a comparison.
		status = C.nvat_http_options_set_tls_ca_cert_dyn(&n.httpOptions, caPath)
		if status == C.NVAT_RC_FEATURE_NOT_ENABLED {
			logger.Infof("NVAT http options TLS CA override not supported by current NVAT build; continuing with defaults\n")
		} else if status != C.NVAT_RC_OK {
			//nolint:gocritic // CGo call; &n.httpOptions is an input/output parameter, not a comparison.
			C.nvat_http_options_free_dyn(&n.httpOptions)
			return fmt.Errorf("nvat_http_options_set_tls_ca_cert failed with code: %d", int(status))
		}
	}

	switch n.options.VerifierMode {
	case "remote":
		if err := n.createGpuNrasVerifier(); err != nil {
			return err
		}
		n.gpuVerifier = C.nvat_gpu_nras_verifier_upcast_dyn(n.gpuNrasVerifier)
	case "local":
		if err := n.createRemoteRIMStore(); err != nil {
			return err
		}

		if err := n.createOcspClient(); err != nil {
			return err
		}

		if err := n.createGpuLocalVerifier(); err != nil {
			return err
		}
		n.gpuVerifier = C.nvat_gpu_local_verifier_upcast_dyn(n.gpuLocalVerifier)
	default:
		return fmt.Errorf("unknown verifier mode: %s", n.options.VerifierMode)
	}

	return nil
}

func (n *NvatGpuInterface) VerifyEvidence(nonceHex string, evidenceJSON []byte) error {
	cNonceHex := C.CString(nonceHex)
	defer C.free(unsafe.Pointer(cNonceHex))

	var nonce C.nvat_nonce_t
	//nolint:gocritic // CGo call; &nonce is an output parameter, not a comparison.
	status := C.nvat_nonce_from_hex_dyn(&nonce, cNonceHex)
	if status != C.NVAT_RC_OK {
		return fmt.Errorf("nvat_nonce_from_hex failed with code: %d", int(status))
	}
	//nolint:gocritic // CGo call; &nonce is an input/output parameter, not a comparison.
	defer C.nvat_nonce_free_dyn(&nonce)

	cEvidenceJSON := C.CString(string(evidenceJSON))
	defer C.free(unsafe.Pointer(cEvidenceJSON))

	var evidenceSource C.nvat_gpu_evidence_source_t
	//nolint:gocritic // CGo call; &evidenceSource is an output parameter, not a comparison.
	status = C.nvat_gpu_evidence_source_from_json_string_dyn(&evidenceSource, cEvidenceJSON)
	if status != C.NVAT_RC_OK {
		return fmt.Errorf("nvat_gpu_evidence_source_from_json_string failed with code: %d", int(status))
	}
	//nolint:gocritic // CGo call; &evidenceSource is an input/output parameter, not a comparison.
	defer C.nvat_gpu_evidence_source_free_dyn(&evidenceSource)

	var gpuEvidenceArray *C.nvat_gpu_evidence_t
	var numEvidences C.size_t
	//nolint:gocritic // CGo call; pointer arguments are output parameters, not comparisons.
	status = C.nvat_gpu_evidence_collect_dyn(evidenceSource, nonce, &gpuEvidenceArray, &numEvidences)
	if status != C.NVAT_RC_OK {
		return fmt.Errorf("nvat_gpu_evidence_collect failed with code: %d", int(status))
	}
	defer func() {
		if gpuEvidenceArray != nil {
			//nolint:gocritic // CGo call; &gpuEvidenceArray is an input/output parameter, not a comparison.
			C.nvat_gpu_evidence_array_free_dyn(&gpuEvidenceArray, numEvidences)
		}
	}()

	var policy C.nvat_evidence_policy_t
	//nolint:gocritic // CGo call; &policy is an output parameter, not a comparison.
	status = C.nvat_evidence_policy_create_default_dyn(&policy)
	if status != C.NVAT_RC_OK {
		return fmt.Errorf("nvat_evidence_policy_create_default failed with code: %d", int(status))
	}
	//nolint:gocritic // CGo call; &policy is an input/output parameter, not a comparison.
	defer C.nvat_evidence_policy_free_dyn(&policy)

	var claims C.nvat_claims_collection_t = nil
	//nolint:gocritic // CGo call; &claims is an output parameter, not a comparison.
	status = C.nvat_verify_gpu_evidence_dyn(
		n.gpuVerifier,
		gpuEvidenceArray,
		numEvidences,
		policy,
		nil,
		&claims,
	)
	if status != C.NVAT_RC_OK {
		return fmt.Errorf("evidence appraisal failed with NVAT status: %d", int(status))
	}

	if claims != nil {
		//nolint:gocritic // CGo call; &claims is an input/output parameter, not a comparison.
		C.nvat_claims_collection_free_dyn(&claims)
	}

	return nil
}

func (n *NvatGpuInterface) createRemoteRIMStore() error {
	var remoteHost *C.char
	if n.options.RemoteHost != "" {
		remoteHost = C.CString(n.options.RemoteHost)
		defer C.free(unsafe.Pointer(remoteHost))
	}
	apiKey := C.CString(n.options.ServiceToken)
	defer C.free(unsafe.Pointer(apiKey))

	//nolint:gocritic // CGo call; &n.rimStore is an output parameter, not a comparison.
	status := C.nvat_rim_store_create_remote_dyn(&n.rimStore, remoteHost, apiKey, n.httpOptions)
	if status != 0 {
		return fmt.Errorf("nvat_rim_store_create_remote failed with code: %d", int(status))
	}
	return nil
}

func (n *NvatGpuInterface) createOcspClient() error {
	var remoteHost *C.char
	if n.options.RemoteHost != "" {
		remoteHost = C.CString(n.options.RemoteHost)
		defer C.free(unsafe.Pointer(remoteHost))
	}
	apiKey := C.CString(n.options.ServiceToken)
	defer C.free(unsafe.Pointer(apiKey))

	//nolint:gocritic // CGo call; &n.ocspClient is an output parameter, not a comparison.
	status := C.nvat_ocsp_client_create_default_dyn(&n.ocspClient, remoteHost, apiKey, n.httpOptions)
	if status != 0 {
		return fmt.Errorf("nvat_ocsp_client_create_default failed with code: %d", int(status))
	}
	return nil
}

func (n *NvatGpuInterface) createGpuLocalVerifier() error {
	var detachedEatOptions C.nvat_detached_eat_options_t

	//nolint:gocritic // CGo call; &n.gpuLocalVerifier is an output parameter, not a comparison.
	status := C.nvat_gpu_local_verifier_create_dyn(&n.gpuLocalVerifier, n.rimStore, n.ocspClient, detachedEatOptions)
	if status != 0 {
		return fmt.Errorf("nvat_gpu_local_verifier failed with code: %d", int(status))
	}
	return nil
}

func (n *NvatGpuInterface) createGpuNrasVerifier() error {
	var remoteHost *C.char
	if n.options.RemoteHost != "" {
		remoteHost = C.CString(n.options.RemoteHost)
		defer C.free(unsafe.Pointer(remoteHost))
	}
	apiKey := C.CString(n.options.ServiceToken)
	defer C.free(unsafe.Pointer(apiKey))

	//nolint:gocritic // CGo call; &n.gpuNrasVerifier is an output parameter, not a comparison.
	status := C.nvat_gpu_nras_verifier_create_dyn(&n.gpuNrasVerifier, remoteHost, apiKey, n.httpOptions)
	if status != 0 {
		return fmt.Errorf("nvat_gpu_nras_verifier failed with code: %d", int(status))
	}

	return nil
}
