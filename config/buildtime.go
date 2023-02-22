// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package config

import "fmt"

var Version = "N/A"

var Developer = "Veraison Project"

// Valid values: "plugins", "builtin"
var SchemeLoader = "plugins"

func init() {
	if SchemeLoader != "plugins" && SchemeLoader != "builtin" {
		panic(fmt.Sprintf(
			`invalid scheme loader value: %q; must be either "plugins" or "builtin"`,
			SchemeLoader,
		))
	}
}
