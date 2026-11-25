// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package plugin

import "net/rpc"

type RPCChannel[I IPluggable] struct {
	GetClient func(c *rpc.Client) any
	GetServer func(i I) any
}
