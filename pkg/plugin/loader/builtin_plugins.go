//go:build builtin

// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package loader

import (
	"github.com/veraison/services/pkg/endorsement"
	"github.com/veraison/services/pkg/plugin/builtin"
	"github.com/veraison/services/pkg/scheme"

	scheme01 "github.com/veraison/services/plugins/scheme/arm-cca"
	scheme02 "github.com/veraison/services/plugins/scheme/da-spdm"
	scheme03 "github.com/veraison/services/plugins/scheme/psa-iot"
	scheme04 "github.com/veraison/services/plugins/scheme/sevsnp"
	scheme05 "github.com/veraison/services/plugins/scheme/tpm-enacttrust"
	store01 "github.com/veraison/services/plugins/store/amd-kds-coserv"
	store02 "github.com/veraison/services/plugins/store/corim-store"
	store03 "github.com/veraison/services/plugins/store/coserv-broker"
	store04 "github.com/veraison/services/plugins/store/nvidia-coserv"
)

func init() {
	builtin.MustRegister(scheme.PluginType,
		scheme.MustNewSchemeImplementationWrapper(scheme01.Descriptor, scheme01.NewImplementation()))
	builtin.MustRegister(scheme.PluginType,
		scheme.MustNewSchemeImplementationWrapper(scheme02.Descriptor, scheme02.NewImplementation()))
	builtin.MustRegister(scheme.PluginType,
		scheme.MustNewSchemeImplementationWrapper(scheme03.Descriptor, scheme03.NewImplementation()))
	builtin.MustRegister(scheme.PluginType,
		scheme.MustNewSchemeImplementationWrapper(scheme04.Descriptor, scheme04.NewImplementation()))
	builtin.MustRegister(scheme.PluginType,
		scheme.MustNewSchemeImplementationWrapper(scheme05.Descriptor, scheme05.NewImplementation()))

	builtin.MustRegister(endorsement.StorePluginType, endorsement.MustWrapHandler(
		store01.NewStore(),
		endorsement.ExecuteCoservQuery,
	))
	builtin.MustRegister(endorsement.StorePluginType, endorsement.MustWrapHandler(
		store02.NewStore(),
		endorsement.GetTriples,
		endorsement.UpdateStore,
		endorsement.ExecuteCoservQuery,
	))
	builtin.MustRegister(endorsement.StorePluginType, endorsement.MustWrapHandler(
		store03.NewStore(),
		endorsement.ExecuteCoservQuery,
	))
	builtin.MustRegister(endorsement.StorePluginType, endorsement.MustWrapHandler(
		store04.NewStore(),
		endorsement.ExecuteCoservQuery,
	))
}
