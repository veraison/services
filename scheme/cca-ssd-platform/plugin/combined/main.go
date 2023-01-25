// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/veraison/services/decoder"
	"github.com/veraison/services/plugin"
	scheme "github.com/veraison/services/scheme/cca-ssd-platform"
)

func main() {
	decoder.RegisterEndorsementDecoder(&scheme.EndorsementDecoder{})
	decoder.RegisterEvidenceDecoder(&scheme.EvidenceDecoder{})
	plugin.Serve()
}
