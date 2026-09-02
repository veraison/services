// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package builtin

import (
	"github.com/veraison/services/handler"
	"github.com/veraison/services/plugin"

	scheme3 "github.com/veraison/services/scheme/arm-cca"
	scheme10 "github.com/veraison/services/scheme/da-spdm"
	scheme1 "github.com/veraison/services/scheme/parsec-cca"
	scheme5 "github.com/veraison/services/scheme/parsec-tpm"
	scheme6 "github.com/veraison/services/scheme/psa-iot"
	scheme7 "github.com/veraison/services/scheme/sevsnp"
	scheme4 "github.com/veraison/services/scheme/tpm-enacttrust"
	store3 "github.com/veraison/services/store-plugin/amd-kds-coserv"
	store1 "github.com/veraison/services/store-plugin/corim-store"
	store2 "github.com/veraison/services/store-plugin/nvidia-coserv"
)

type PluginClass uint8

const (
	SchemePlugin PluginClass = iota
	StorePlugin
)

var plugins = map[PluginClass][]plugin.IPluggable{
	SchemePlugin: schemePlugins,
	StorePlugin:  storePlugins,
}

var schemePlugins = []plugin.IPluggable{
	handler.MustNewSchemeImplementationWrapper(scheme1.Descriptor, scheme1.NewImplementation()),
	handler.MustNewSchemeImplementationWrapper(scheme3.Descriptor, scheme3.NewImplementation()),
	handler.MustNewSchemeImplementationWrapper(scheme4.Descriptor, scheme4.NewImplementation()),
	handler.MustNewSchemeImplementationWrapper(scheme5.Descriptor, scheme5.NewImplementation()),
	handler.MustNewSchemeImplementationWrapper(scheme6.Descriptor, scheme6.NewImplementation()),
	handler.MustNewSchemeImplementationWrapper(scheme7.Descriptor, scheme7.NewImplementation()),
	handler.MustNewSchemeImplementationWrapper(scheme10.Descriptor, scheme10.NewImplementation()),
}

var storePlugins = []plugin.IPluggable{
	&store2.CoservProxyHandler{},
	&store3.CoservProxyHandler{},
	store1.NewStore(),
}
