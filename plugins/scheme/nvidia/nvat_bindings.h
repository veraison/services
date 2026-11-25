// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
#ifndef NVAT_BINDINGS_H
#define NVAT_BINDINGS_H

#include <nvat.h>
#include <stdlib.h>

int nvat_load_library(void);
const char* nvat_last_error(void);
nvat_rc_t nvat_sdk_opts_create_dyn(nvat_sdk_opts_t* opts);
void nvat_sdk_opts_free_dyn(nvat_sdk_opts_t* opts);
nvat_rc_t nvat_sdk_init_dyn(nvat_sdk_opts_t opts);
nvat_rc_t nvat_rim_store_create_remote_dyn(nvat_rim_store_t* store, const char* host, const char* api_key, nvat_http_options_t http_opts);
nvat_rc_t nvat_ocsp_client_create_default_dyn(nvat_ocsp_client_t* client, const char* host, const char* api_key, nvat_http_options_t http_opts);
nvat_rc_t nvat_gpu_local_verifier_create_dyn(nvat_gpu_local_verifier_t* verifier, nvat_rim_store_t store, nvat_ocsp_client_t client, nvat_detached_eat_options_t options);
nvat_rc_t nvat_gpu_nras_verifier_create_dyn(nvat_gpu_nras_verifier_t* verifier, const char* host, const char* api_key, nvat_http_options_t http_opts);
nvat_gpu_verifier_t nvat_gpu_nras_verifier_upcast_dyn(nvat_gpu_nras_verifier_t verifier);
nvat_gpu_verifier_t nvat_gpu_local_verifier_upcast_dyn(nvat_gpu_local_verifier_t verifier);
nvat_rc_t nvat_nonce_from_hex_dyn(nvat_nonce_t* nonce, const char* hex);
void nvat_nonce_free_dyn(nvat_nonce_t* nonce);
nvat_rc_t nvat_gpu_evidence_source_from_json_string_dyn(nvat_gpu_evidence_source_t* source, const char* json);
void nvat_gpu_evidence_source_free_dyn(nvat_gpu_evidence_source_t* source);
nvat_rc_t nvat_gpu_evidence_collect_dyn(nvat_gpu_evidence_source_t source, nvat_nonce_t nonce, nvat_gpu_evidence_t** evidence_array, size_t* num_evidences);
void nvat_gpu_evidence_array_free_dyn(nvat_gpu_evidence_t** evidence_array, size_t num_evidences);
nvat_rc_t nvat_evidence_policy_create_default_dyn(nvat_evidence_policy_t* policy);
void nvat_evidence_policy_free_dyn(nvat_evidence_policy_t* policy);
nvat_rc_t nvat_verify_gpu_evidence_dyn(const nvat_gpu_verifier_t verifier, const nvat_gpu_evidence_t* evidence_array, size_t num_evidences, const nvat_evidence_policy_t policy, nvat_str_t* out_detached_eat, nvat_claims_collection_t* out_claims);
void nvat_claims_collection_free_dyn(nvat_claims_collection_t* claims);
nvat_rc_t nvat_http_options_set_tls_ca_cert_dyn(nvat_http_options_t* opts, const char* path);
nvat_rc_t nvat_http_options_create_default_dyn(nvat_http_options_t* opts);
void nvat_http_options_free_dyn(nvat_http_options_t* opts);

#endif /* NVAT_BINDINGS_H */
