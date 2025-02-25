// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sessionmanager

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/veraison/services/config"
)

const DefaultMemcachedServer = "localhost:11211"

type memcachedConfig struct {
	Servers []string `mapstructure:"servers"`
}

type Memcached struct {
	client *memcache.Client
}

func NewMemcached() *Memcached {
	return &Memcached{}
}

func (o *Memcached) Init(v *viper.Viper) error {
	cfg := memcachedConfig{
		Servers: []string{DefaultMemcachedServer},
	}

	if v != nil {
		loader := config.NewLoader(&cfg)
		if err := loader.LoadFromViper(v); err != nil {
			return fmt.Errorf("memcached: %w", err)
		}
	}

	client := memcache.New(cfg.Servers...)
	if err := client.Ping(); err != nil {
		return fmt.Errorf("memcached: %w", err)
	}
	o.client = client

	return nil
}

func (o *Memcached) SetSession(
	id uuid.UUID,
	tenant string,
	session json.RawMessage,
	ttl time.Duration,
) error {
	item := &memcache.Item{
		Key: makeKey(id, tenant),
		Value: session,
		Expiration: int32(ttl.Seconds()),
	}
	return o.client.Set(item)
}

func (o *Memcached) DelSession(id uuid.UUID, tenant string) error {
	return o.client.Delete(makeKey(id, tenant))
}

func (o *Memcached) GetSession(id uuid.UUID, tenant string) (json.RawMessage, error) {
	item, err := o.client.Get(makeKey(id, tenant))
	if err != nil {
		if err.Error() == "memcache: cache miss" {
			return nil, fmt.Errorf(
				"session not found for (id, tenant)=(%s, %s)", id, tenant)
		}
		return nil, err
	}

	return item.Value, nil
}

func (o *Memcached) Close() error {
	return o.client.Close()
}
