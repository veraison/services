//go:build !builtin

// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package loader

import (
	"github.com/veraison/services/pkg/plugin"
	"github.com/veraison/services/pkg/plugin/goplugin"
	"go.uber.org/zap"
)

func NewPluginLoader(logger *zap.SugaredLogger) plugin.ILoader {
	return goplugin.NewPluginLoader(logger)
}
