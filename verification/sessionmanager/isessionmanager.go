// Copyright 2022-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sessionmanager

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type ISessionManager interface {
	Init(v *viper.Viper) error
	SetSession(id uuid.UUID, tenant string, session json.RawMessage, ttl time.Duration) error
	GetSession(id uuid.UUID, tenant string) (json.RawMessage, error)
	DelSession(id uuid.UUID, tenant string) error
	Close() error
}
