// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

// Veraison services common command line flags:
// -c, --config <configuration-file> (default "config.yaml")
// -h, --help
// -v, --version
var (
	displayVersion *bool   = flag.BoolP("version", "v", false, "print version and exit")
	displayHelp    *bool   = flag.BoolP("help", "h", false, "print help and exit")
	file           *string = flag.StringP("config", "c", "config.yaml", "configuration file")
)

// CmdLine parses the command line an
func CmdLine() {
	flag.Parse()

	if *displayVersion {
		fmt.Printf("%v\n", Version)
		os.Exit(0)
	}

	if *displayHelp {
		flag.Usage()
		os.Exit(0)
	}
}

// File returns path to the file from which configuration will be read by
// default. The value defaults to "config.yaml" in the current directory but
// may be overridden with --config CLI flag.
func File() string {
	if file == nil {
		return ""
	}

	return *file
}
