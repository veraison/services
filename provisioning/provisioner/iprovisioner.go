// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package provisioner

import "github.com/veraison/services/proto"

type IProvisioner interface {
	GetVTSState() (*proto.ServiceState, error)
	IsSupportedMediaType(mt string) (bool, error)
	SupportedMediaTypes() ([]string, error)
	SubmitEndorsements(tenantID string, data []byte, mt string) error
}
