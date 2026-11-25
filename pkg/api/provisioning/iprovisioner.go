// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package provisioning

//go:generate mockgen -source iprovisioner.go -destination mock_provisioner_test.go -package provisioning
type IProvisioner interface {
	ServerVersion() (string, error)
	IsSupportedMediaType(mt string) (bool, error)
	SupportedMediaTypes() ([]string, error)
	SubmitEndorsements(tenantID string, data []byte, mt string) error
	SetEndorsementsState(tenantID string, data []byte, setActive bool) error
}
