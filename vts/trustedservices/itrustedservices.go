// Copyright 2022-2024 Contributors to the Veraison project.
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
		evm plugin.IManager[handler.IEvidenceHandler],
		endm plugin.IManager[handler.IEndorsementHandler],
		stm plugin.IManager[handler.IStoreHandler],
	) error
	Close() error
	Run() error

	proto.VTSServer
}
