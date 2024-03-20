// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"github.com/veraison/ear"
	"github.com/veraison/services/config"
)

func CreateAttestationResult(submodName string) *ear.AttestationResult {
	return ear.NewAttestationResult(submodName, config.Version, config.Developer)
}
