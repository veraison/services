// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package nvidia

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include "nvat_bindings.h"
#include <string.h>

static void* nvatHandle = NULL;
static char nvatLastError[256];

static typeof(&nvat_sdk_opts_create) nvat_sdk_opts_create_sym = NULL;
static typeof(&nvat_sdk_opts_free) nvat_sdk_opts_free_sym = NULL;
static typeof(&nvat_sdk_init) nvat_sdk_init_sym = NULL;
static typeof(&nvat_rim_store_create_remote) nvat_rim_store_create_remote_sym = NULL;
static typeof(&nvat_ocsp_client_create_default) nvat_ocsp_client_create_default_sym = NULL;
static typeof(&nvat_gpu_local_verifier_create) nvat_gpu_local_verifier_create_sym = NULL;
static typeof(&nvat_gpu_nras_verifier_create) nvat_gpu_nras_verifier_create_sym = NULL;
static typeof(&nvat_gpu_nras_verifier_upcast) nvat_gpu_nras_verifier_upcast_sym = NULL;
static typeof(&nvat_gpu_local_verifier_upcast) nvat_gpu_local_verifier_upcast_sym = NULL;
static typeof(&nvat_nonce_from_hex) nvat_nonce_from_hex_sym = NULL;
static typeof(&nvat_nonce_free) nvat_nonce_free_sym = NULL;
static typeof(&nvat_gpu_evidence_source_from_json_string) nvat_gpu_evidence_source_from_json_string_sym = NULL;
static typeof(&nvat_gpu_evidence_source_free) nvat_gpu_evidence_source_free_sym = NULL;
static typeof(&nvat_gpu_evidence_collect) nvat_gpu_evidence_collect_sym = NULL;
static typeof(&nvat_gpu_evidence_array_free) nvat_gpu_evidence_array_free_sym = NULL;
static typeof(&nvat_evidence_policy_create_default) nvat_evidence_policy_create_default_sym = NULL;
static typeof(&nvat_evidence_policy_free) nvat_evidence_policy_free_sym = NULL;
static typeof(&nvat_verify_gpu_evidence) nvat_verify_gpu_evidence_sym = NULL;
static typeof(&nvat_claims_collection_free) nvat_claims_collection_free_sym = NULL;
typedef nvat_rc_t (*nvat_http_options_set_tls_ca_cert_fn)(nvat_http_options_t*, const char*);
static nvat_http_options_set_tls_ca_cert_fn nvat_http_options_set_tls_ca_cert_sym = NULL;
typedef nvat_rc_t (*nvat_http_options_create_default_fn)(nvat_http_options_t*);
static nvat_http_options_create_default_fn nvat_http_options_create_default_sym = NULL;
typedef void (*nvat_http_options_free_fn)(nvat_http_options_t*);
static nvat_http_options_free_fn nvat_http_options_free_sym = NULL;

static void setLastError(const char* err)
{
	if (err == NULL) {
		nvatLastError[0] = '\0';
		return;
	}
	strncpy(nvatLastError, err, sizeof(nvatLastError) - 1);
	nvatLastError[sizeof(nvatLastError) - 1] = '\0';
}

static int load_symbol(void** target, const char* name)
{
	dlerror();
	void* sym = dlsym(nvatHandle, name);
	const char* err = dlerror();
	if (err != NULL) {
		setLastError(err);
		return -1;
	}
	*target = sym;
	return 0;
}

static void load_optional_symbol(void** target, const char* name)
{
	dlerror();
	void* sym = dlsym(nvatHandle, name);
	const char* err = dlerror();
	if (err != NULL) {
		*target = NULL;
		return;
	}
	*target = sym;
}

static int loadNvatLibraryWithPath(const char* path)
{
	if (nvatHandle != NULL) {
		return 0;
	}

	dlerror();
	nvatHandle = dlopen(path, RTLD_LAZY | RTLD_LOCAL);
	if (nvatHandle == NULL) {
		setLastError(dlerror());
		return -1;
	}

#define LOAD_SYMBOL(name)                                                         \
	if (load_symbol((void**)&name##_sym, #name) != 0) {                           \
		dlclose(nvatHandle);                                                      \
		nvatHandle = NULL;                                                        \
		return -1;                                                                \
	}

	LOAD_SYMBOL(nvat_sdk_opts_create);
	LOAD_SYMBOL(nvat_sdk_opts_free);
	LOAD_SYMBOL(nvat_sdk_init);
	LOAD_SYMBOL(nvat_rim_store_create_remote);
	LOAD_SYMBOL(nvat_ocsp_client_create_default);
	LOAD_SYMBOL(nvat_gpu_local_verifier_create);
	LOAD_SYMBOL(nvat_gpu_nras_verifier_create);
	LOAD_SYMBOL(nvat_gpu_nras_verifier_upcast);
	LOAD_SYMBOL(nvat_gpu_local_verifier_upcast);
	LOAD_SYMBOL(nvat_nonce_from_hex);
	LOAD_SYMBOL(nvat_nonce_free);
	LOAD_SYMBOL(nvat_gpu_evidence_source_from_json_string);
	LOAD_SYMBOL(nvat_gpu_evidence_source_free);
	LOAD_SYMBOL(nvat_gpu_evidence_collect);
	LOAD_SYMBOL(nvat_gpu_evidence_array_free);
	LOAD_SYMBOL(nvat_evidence_policy_create_default);
	LOAD_SYMBOL(nvat_evidence_policy_free);
	LOAD_SYMBOL(nvat_verify_gpu_evidence);
	LOAD_SYMBOL(nvat_claims_collection_free);
	LOAD_SYMBOL(nvat_http_options_create_default);
	LOAD_SYMBOL(nvat_http_options_free);

	load_optional_symbol((void**)&nvat_http_options_set_tls_ca_cert_sym, "nvat_http_options_set_tls_ca_cert");

#undef LOAD_SYMBOL

	setLastError(NULL);
	return 0;
}

int nvat_load_library()
{
	return loadNvatLibraryWithPath("libnvat.so.1");
}

const char* nvat_last_error()
{
	return nvatLastError[0] != '\0' ? nvatLastError : "unknown error";
}

nvat_rc_t nvat_sdk_opts_create_dyn(nvat_sdk_opts_t* opts)
{
	return nvat_sdk_opts_create_sym(opts);
}

void nvat_sdk_opts_free_dyn(nvat_sdk_opts_t* opts)
{
	nvat_sdk_opts_free_sym(opts);
}

nvat_rc_t nvat_sdk_init_dyn(nvat_sdk_opts_t opts)
{
	return nvat_sdk_init_sym(opts);
}

nvat_rc_t nvat_rim_store_create_remote_dyn(nvat_rim_store_t* store, const char* host, const char* api_key, nvat_http_options_t http_opts)
{
	return nvat_rim_store_create_remote_sym(store, host, api_key, http_opts);
}

nvat_rc_t nvat_ocsp_client_create_default_dyn(nvat_ocsp_client_t* client, const char* host, const char* api_key, nvat_http_options_t http_opts)
{
	return nvat_ocsp_client_create_default_sym(client, host, api_key, http_opts);
}

nvat_rc_t nvat_gpu_local_verifier_create_dyn(nvat_gpu_local_verifier_t* verifier, nvat_rim_store_t store, nvat_ocsp_client_t client, nvat_detached_eat_options_t options)
{
	return nvat_gpu_local_verifier_create_sym(verifier, store, client, options);
}

nvat_rc_t nvat_gpu_nras_verifier_create_dyn(nvat_gpu_nras_verifier_t* verifier, const char* host, const char* api_key, nvat_http_options_t http_opts)
{
	return nvat_gpu_nras_verifier_create_sym(verifier, host, api_key, http_opts);
}

nvat_gpu_verifier_t nvat_gpu_nras_verifier_upcast_dyn(nvat_gpu_nras_verifier_t verifier)
{
	return nvat_gpu_nras_verifier_upcast_sym(verifier);
}

nvat_gpu_verifier_t nvat_gpu_local_verifier_upcast_dyn(nvat_gpu_local_verifier_t verifier)
{
	return nvat_gpu_local_verifier_upcast_sym(verifier);
}

nvat_rc_t nvat_nonce_from_hex_dyn(nvat_nonce_t* nonce, const char* hex)
{
	return nvat_nonce_from_hex_sym(nonce, hex);
}

void nvat_nonce_free_dyn(nvat_nonce_t* nonce)
{
	nvat_nonce_free_sym(nonce);
}

nvat_rc_t nvat_gpu_evidence_source_from_json_string_dyn(nvat_gpu_evidence_source_t* source, const char* json)
{
	return nvat_gpu_evidence_source_from_json_string_sym(source, json);
}

void nvat_gpu_evidence_source_free_dyn(nvat_gpu_evidence_source_t* source)
{
	nvat_gpu_evidence_source_free_sym(source);
}

nvat_rc_t nvat_gpu_evidence_collect_dyn(nvat_gpu_evidence_source_t source, nvat_nonce_t nonce, nvat_gpu_evidence_t** evidence_array, size_t* num_evidences)
{
	return nvat_gpu_evidence_collect_sym(source, nonce, evidence_array, num_evidences);
}

void nvat_gpu_evidence_array_free_dyn(nvat_gpu_evidence_t** evidence_array, size_t num_evidences)
{
	nvat_gpu_evidence_array_free_sym(evidence_array, num_evidences);
}

nvat_rc_t nvat_evidence_policy_create_default_dyn(nvat_evidence_policy_t* policy)
{
	return nvat_evidence_policy_create_default_sym(policy);
}

void nvat_evidence_policy_free_dyn(nvat_evidence_policy_t* policy)
{
	nvat_evidence_policy_free_sym(policy);
}

nvat_rc_t nvat_verify_gpu_evidence_dyn(const nvat_gpu_verifier_t verifier, const nvat_gpu_evidence_t* evidence_array, size_t num_evidences, const nvat_evidence_policy_t policy, nvat_str_t* out_detached_eat, nvat_claims_collection_t* out_claims)
{
	return nvat_verify_gpu_evidence_sym(verifier, evidence_array, num_evidences, policy, out_detached_eat, out_claims);
}

void nvat_claims_collection_free_dyn(nvat_claims_collection_t* claims)
{
	nvat_claims_collection_free_sym(claims);
}

nvat_rc_t nvat_http_options_set_tls_ca_cert_dyn(nvat_http_options_t* opts, const char* path)
{
	if (nvat_http_options_set_tls_ca_cert_sym == NULL) {
		return NVAT_RC_FEATURE_NOT_ENABLED;
	}
	return nvat_http_options_set_tls_ca_cert_sym(opts, path);
}

nvat_rc_t nvat_http_options_create_default_dyn(nvat_http_options_t* opts)
{
	return nvat_http_options_create_default_sym(opts);
}

void nvat_http_options_free_dyn(nvat_http_options_t* opts)
{
	nvat_http_options_free_sym(opts);
}
*/
import "C"
