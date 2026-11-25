// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package session

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

//go:generate mockgen -source imanager.go -destination ../mock_session_manager_test.go -package verification
type IManager interface {
	Init(v *viper.Viper) error
	SetSession(id uuid.UUID, tenant string, session json.RawMessage, ttl time.Duration) error
	GetSession(id uuid.UUID, tenant string) (json.RawMessage, error)
	DelSession(id uuid.UUID, tenant string) error
	Close() error
}
