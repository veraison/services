// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package vts

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const DefaultTenantID = "0"

func SetTenantID(ctx context.Context, value string) context.Context {
	md := metadata.Pairs("tenant-id", value)
	return metadata.NewOutgoingContext(ctx, md)
}

func GetTenantID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return DefaultTenantID
	}

	values := md.Get("tenant-id")
	if len(values) == 0 {
		return DefaultTenantID
	}

	return values[0]
}
