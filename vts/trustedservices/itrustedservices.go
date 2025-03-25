// Copyright 2022-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package trustedservices

import (
	"github.com/spf13/viper"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/proto"
)

type ITrustedServices interface {
	Init(
		cfg *viper.Viper,
		evidenceManager plugin.IManager[handler.IEvidenceHandler],
		endorsementManager plugin.IManager[handler.IEndorsementHandler],
		storeManager plugin.IManager[handler.IStoreHandler],
		coservProxyManager plugin.IManager[handler.ICoservProxyHandler],
	) error
	Close() error
	Run() error

	proto.VTSServer
}
