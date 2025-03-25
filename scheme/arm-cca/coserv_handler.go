// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package arm_cca

import (
	"errors"
)

type CoservProxyHandler struct{}

func (s CoservProxyHandler) GetName() string {
	return "cca-coserv-handler"
}

func (s CoservProxyHandler) GetAttestationScheme() string {
	return SchemeName
}

func (s CoservProxyHandler) GetSupportedMediaTypes() []string {
	return CoservMediaTypes
}

func (s CoservProxyHandler) GetEndorsements(tenantID string, query string) ([]byte, error) {
	// TODO move to a new NVIDIA / AMD plugin
	return nil, errors.New("TODO CCA plugin")
}
