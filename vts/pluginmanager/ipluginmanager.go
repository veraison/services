// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package pluginmanager

import (
	"github.com/spf13/viper"
	"github.com/veraison/services/scheme"
)

type ISchemePluginManager interface {
	Init(*viper.Viper) error
	Close() error

	LookupByMediaType(mediaType string) (scheme.IScheme, error)
	LookupBySchemeName(format string) (scheme.IScheme, error)
	SupportedVerificationMediaTypes() ([]string, error)
}
