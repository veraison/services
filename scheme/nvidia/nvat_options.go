// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package nvidia

import (
	"fmt"
	"os"

	"github.com/veraison/services/log"
)

type NvatOptions struct {
	VerifierMode string
	ServiceToken string
	CaTrustFile  string
	RemoteHost   string
}

var defaultNvatOptions NvatOptions
var logger = log.Named("NVIDIA-options")

func init() {
	defaultNvatOptions.VerifierMode = "remote"

	if _, err := os.Stat("/etc/ssl/certs/ca-certificates.crt"); err == nil {
		defaultNvatOptions.CaTrustFile = "/etc/ssl/certs/ca-certificates.crt"
		return
	}

	if _, err := os.Stat("/etc/pki/tls/certs/ca-bundle.crt"); err == nil {
		defaultNvatOptions.CaTrustFile = "/etc/pki/tls/certs/ca-bundle.crt"
		return
	}

	defaultNvatOptions.CaTrustFile = ""
}

func DefaultNvatOptions() NvatOptions {
	return defaultNvatOptions
}

func (o NvatOptions) Valid() error {
	if o.VerifierMode == "" {
		return fmt.Errorf("verifier mode is required")
	}

	if o.VerifierMode != "remote" && o.VerifierMode != "local" {
		return fmt.Errorf("unsupported verifier mode: %s", o.VerifierMode)
	}

	if o.VerifierMode == "remote" && o.ServiceToken == "" {
		return fmt.Errorf("service token is required")
	}

	logger.Infof("NVAT option VerifierMode=%q\n", o.VerifierMode)

	if o.ServiceToken != "" {
		logger.Infof("NVAT option ServiceToken is set\n")
	}

	if o.CaTrustFile != "" {
		logger.Infof("NVAT option CaTrustFile=%q\n", o.CaTrustFile)
	} else {
		logger.Infof("NVAT option CaTrustFile not set; using NVAT defaults\n")
	}

	if o.RemoteHost != "" {
		logger.Infof("NVAT option RemoteHost=%q\n", o.RemoteHost)
	} else {
		logger.Infof("NVAT option RemoteHost not set; using NVAT defaults\n")
	}

	return nil
}
