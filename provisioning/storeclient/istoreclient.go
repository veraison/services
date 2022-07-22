// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package storeclient

import (
	"github.com/veraison/services/proto"
)

type IStoreClient interface {
	proto.VTSClient
}
