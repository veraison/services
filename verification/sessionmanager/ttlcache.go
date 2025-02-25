// Copyright 2022-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sessionmanager

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jellydator/ttlcache/v3"
	"github.com/spf13/viper"
)

type TTLCache struct {
	cache *ttlcache.Cache[string, json.RawMessage]
}

func NewTTLCache() *TTLCache {
	return &TTLCache{
		cache: ttlcache.New[string, json.RawMessage](),
	}
}

func (o *TTLCache) Init(v *viper.Viper) error {
	o.cache = ttlcache.New[string, json.RawMessage]()

	go o.cache.Start()

	return nil
}

func (o *TTLCache) Close() error {
	o.cache.Stop()

	return nil
}

func (o *TTLCache) SetSession(
	id uuid.UUID,
	tenant string,
	session json.RawMessage,
	ttl time.Duration,
) error {
	_ = o.cache.Set(makeKey(id, tenant), session, ttl)

	return nil
}

func (o *TTLCache) DelSession(id uuid.UUID, tenant string) error {
	o.cache.Delete(makeKey(id, tenant))

	return nil
}

func (o *TTLCache) GetSession(id uuid.UUID, tenant string) (json.RawMessage, error) {
	if item := o.cache.Get(makeKey(id, tenant)); item != nil {
		return item.Value(), nil
	}

	return nil, fmt.Errorf("session not found for (id, tenant)=(%s, %s)", id, tenant)
}
