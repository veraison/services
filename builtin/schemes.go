// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package builtin

import (
	"github.com/veraison/services/handler"
	"github.com/veraison/services/plugin"

	scheme9 "github.com/veraison/services/scheme/amd-kds-coserv"
	scheme3 "github.com/veraison/services/scheme/arm-cca"
	scheme8 "github.com/veraison/services/scheme/nvidia-coserv"
	scheme1 "github.com/veraison/services/scheme/parsec-cca"
	scheme5 "github.com/veraison/services/scheme/parsec-tpm"
	scheme6 "github.com/veraison/services/scheme/psa-iot"
	scheme2 "github.com/veraison/services/scheme/riot"
	scheme7 "github.com/veraison/services/scheme/sevsnp"
	scheme4 "github.com/veraison/services/scheme/tpm-enacttrust"
)

var plugins = []plugin.IPluggable{
	handler.MustNewSchemeImplementationWrapper(scheme1.Descriptor, scheme1.NewImplementation()),
	handler.MustNewSchemeImplementationWrapper(scheme2.Descriptor, scheme2.NewImplementation()),
	handler.MustNewSchemeImplementationWrapper(scheme3.Descriptor, scheme3.NewImplementation()),
	handler.MustNewSchemeImplementationWrapper(scheme4.Descriptor, scheme4.NewImplementation()),
	handler.MustNewSchemeImplementationWrapper(scheme5.Descriptor, scheme5.NewImplementation()),
	handler.MustNewSchemeImplementationWrapper(scheme6.Descriptor, scheme6.NewImplementation()),
	handler.MustNewSchemeImplementationWrapper(scheme7.Descriptor, scheme7.NewImplementation()),
	&scheme8.CoservProxyHandler{},
	&scheme9.CoservProxyHandler{},
}
