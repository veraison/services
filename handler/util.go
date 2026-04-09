// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"fmt"
	"strings"
)

// PluginNameFromScheme generates a plugin name from an attestations scheme
// name.
func PluginNameFromScheme(schemeName string) string {
	name := strings.ToLower(strings.ReplaceAll(schemeName, " ", "-"))
	return fmt.Sprintf("%s-scheme-plugin", name)
}
