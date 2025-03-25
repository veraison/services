// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package nvidiacoserv

import (
	"errors"
)

type CoservProxyHandler struct{}

func (s CoservProxyHandler) GetName() string {
	return "nvidia-coserv-proxy-handler"
}

func (s CoservProxyHandler) GetAttestationScheme() string {
	return SchemeName
}

func (s CoservProxyHandler) GetSupportedMediaTypes() []string {
	return CoservMediaTypes
}

func (s CoservProxyHandler) GetEndorsements(tenantID string, query string) ([]byte, error) {
	return nil, errors.New("TODO NVIDIA proxy plugin")
}
