// Copyright 2023 Contributors to the Veraison project.
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
	DisplayVersion *bool   = flag.BoolP("version", "v", false, "print version and exit")
	DisplayHelp    *bool   = flag.BoolP("help", "h", false, "print help and exit")
	File           *string = flag.StringP("config", "c", "config.yaml", "configuration file")
)

// CmdLine parses the command line an
func CmdLine() {
	flag.Parse()

	if *DisplayVersion {
		fmt.Printf("%v\n", Version)
		os.Exit(0)
	}

	if *DisplayHelp {
		flag.Usage()
		os.Exit(0)
	}
}
