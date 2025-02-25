// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sessionmanager

import (
	"net/url"

	"github.com/google/uuid"
)

func makeKey(id uuid.UUID, tenant string) string {
	// session://{tenant}/{uuid}
	u := url.URL{
		Scheme: "session",
		Host:   tenant,
		Path:   id.String(),
	}

	return u.String()
}
