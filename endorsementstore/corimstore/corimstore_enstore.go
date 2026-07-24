// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package corimstore

import (
	"context"
	"strings"

	corimstore "github.com/veraison/corim-store/pkg/store"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/services/config"
	handler "github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/plugin"
	"go.uber.org/zap"
)

const (
	PluginName = "corim-store"
	SchemeName = "ANY"
)

// implement the IEndorsementStore interface for corimstore.Store
type DefaultStore struct {
	Store  *corimstore.Store
	logger *zap.SugaredLogger
}

func NewStore() *DefaultStore {
	logger := log.Named(PluginName)
	logger.Debug("initializing default store")
	return &DefaultStore{nil, logger}
}

func (s *DefaultStore) GetKeyTriples(env *comid.Environment, label string, exact bool) ([]*comid.KeyTriple, error) {
	res, err := s.Store.GetActiveKeyTriples(env, label, exact)
	if err == corimstore.ErrNoMatch {
		return nil, handler.ErrNotFound
	}
	return res, err
}

func (s *DefaultStore) GetValueTriples(env *comid.Environment, label string, exact bool) ([]*comid.ValueTriple, error) {
	res, err := s.Store.GetActiveValueTriples(env, label, exact)
	if err == corimstore.ErrNoMatch {
		return nil, handler.ErrNotFound
	}
	return res, err
}

func (s *DefaultStore) ExecuteCoservQuery(mediaType, query string) (*coserv.Coserv, error) {
	s.logger.Infof("got coserv query: %v", query)
	fallbackAuthority, err := comid.NewCryptoKeyTaggedBytes([]byte("dummyauth"))
	if err != nil {
		s.logger.Errorf("could not create authority: %v", err)
		return nil, err
	}
	// TODO: read max-validity from config. VTS parses the max expiry
	// in the coserv context structure. But if executing coserv query
	// is made part of store interface, this parameter should be passed
	// to each store. Right now just using 100s. This must be changed to
	// an input parameter once we finalize the interface and its responsibilities
	coservService := corimstore.NewCoSERVService(s.Store, fallbackAuthority, 100)
	var q coserv.Coserv
	if err := q.FromBase64Url(query); err != nil {
		s.logger.Errorf("could not decode string to coserv: %v", err)
		return nil, err
	}
	if err := coservService.UpdateCoSERV(&q); err != nil {
		s.logger.Errorf("could not update coserv: %v", err)
		return nil, err
	}
	// return ErrNotFound instead of empty results
	if q.Results == nil {
		return nil, handler.ErrNotFound
	}
	s.logger.Debugf("got coserv response: %v", q)
	return &q, nil
}

func (s *DefaultStore) AddCorimBytes(data []byte, label string, activate bool) error {
	s.logger.Debugf("adding CoRIM")
	return s.Store.AddBytes(data, label, activate)
}

func (s *DefaultStore) Close() error {
	s.logger.Info("closing corimstore")
	return s.Store.Close()
}

func (s *DefaultStore) Init(params *plugin.Parameters) error {
	if params == nil {
		panic("parameters are required for corimstore")
	}
	var cfg Config

	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromMap(params.Map()); err != nil {
		return err
	}

	s.logger.Debugf("connecting to %s store %s", cfg.DBMS, cfg.DSN)

	store, err := corimstore.Open(context.Background(), cfg.StoreConfig())
	if err != nil {
		panic(err)
	}

	// The store must be initialized before it may be used. In general, we
	// rely on the store being pointed to by DSN to be initialized prior
	// to starting the VTS. For in-memory store this can never be the case, so we
	// initialize it here.
	if strings.Contains(cfg.DSN, ":memory:") {
		if err := store.Init(); err != nil {
			panic(err)
		}
	}

	s.Store = store
	return nil
}

func (s *DefaultStore) GetName() string {
	return PluginName
}

func (s *DefaultStore) GetAttestationScheme() string {
	return SchemeName
}

func (s *DefaultStore) GetSupportedMediaTypes() map[string][]string {
	// the store supports all corim media types, but adding them
	// in the map causes issues while loading plugins, since only
	// one plugin can be available for a media type. This plugin
	// is fetched using name, so this is ok.
	return map[string][]string{
		"any": []string{"*"},
	}
}
