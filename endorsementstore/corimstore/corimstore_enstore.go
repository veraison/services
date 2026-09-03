// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package corimstore

import (
	"context"
	"errors"
	"strings"

	corimstore "github.com/veraison/corim-store/pkg/store"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	handler "github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/plugin"
	vtsstore "github.com/veraison/services/vts/endorsementstore"
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
	CoservCfg *vtsstore.StoreCommonParams
}

func NewStore() *DefaultStore {
	logger := log.Named(PluginName)
	return &DefaultStore{nil, logger, nil}
}

func (s *DefaultStore) GetKeyTriples(env *comid.Environment, label string, exact bool) ([]*comid.KeyTriple, error) {
	res, err := s.Store.GetActiveKeyTriples(env, label, exact)
	if errors.Is(err, corimstore.ErrNoMatch) {
		return nil, handler.ErrNotFound
	}
	return res, err
}

func (s *DefaultStore) GetValueTriples(env *comid.Environment, label string, exact bool) ([]*comid.ValueTriple, error) {
	res, err := s.Store.GetActiveValueTriples(env, label, exact)
	if errors.Is(err, corimstore.ErrNoMatch) {
		return nil, handler.ErrNotFound
	}
	return res, err
}

func (s *DefaultStore) ExecuteCoservQuery(mediaType, query string) (*coserv.Coserv, error) {
	// If reading CoSERV config failed during initialization,
	// CoSERV interface would be disabled.
	if s.CoservCfg == nil {
		s.logger.Errorf("store is not configured for CoSERV")
		return nil, errors.New("missing configurations for CoSERV service")
	}
	s.logger.Infof("got coserv query: %v", query)
	coservService := corimstore.NewCoSERVService(s.Store, s.CoservCfg.FallbackAuthority, s.CoservCfg.MaxExpiry)
	var q coserv.Coserv
	if err := q.FromBase64Url(query); err != nil {
		s.logger.Errorf("could not decode string to coserv: %v", err)
		return nil, err
	}
	if err := coservService.UpdateCoSERV(&q); err != nil {
		s.logger.Errorf("could not update coserv: %v", err)
		return nil, err
	}
	if q.Results == nil {
		return nil, errors.New("internal error: bad CoSERV result: result-set is nil")
	}
	// return ErrNotFound instead of empty results
	if q.Results.AKQ == nil && q.Results.RVQ == nil {
		return nil, handler.ErrNotFound
	}
	s.logger.Debugf("got coserv response: %v", q)
	return &q, nil
}

func (s *DefaultStore) AddCorimBytes(data []byte, label string, activate bool) error {
	s.logger.Debugf("adding CoRIM")
	return s.Store.AddBytes(data, label, activate)
}

func (s *DefaultStore) Fini() error {
	s.logger.Info("closing corimstore")
	if err := s.Store.Close(); err != nil {
		s.logger.Errorf("Failed to close corim-store: %v", err)
		return err
	}
	return nil
}

func (s *DefaultStore) Init(params *plugin.Parameters) error {
	s.logger.Debug("initializing default store")
	if params == nil {
		return errors.New("parameters are required for corimstore")
	}
	cfg, err := ConfigFromParameters(params, s.logger)
	if err != nil {
		s.logger.Errorf("Failed to load corim-store parameters: %v", err)
		return err
	}
	s.CoservCfg = cfg.StoreCommonParams

	s.logger.Debugf("connecting to %s store %s", cfg.DBMS, cfg.DSN)

	store, err := corimstore.Open(context.Background(), cfg.StoreConfig())
	if err != nil {
		return err
	}

	// The store must be initialized before it may be used. In general, we
	// rely on the store being pointed to by DSN to be initialized prior
	// to starting the VTS. For in-memory store this can never be the case, so we
	// initialize it here.
	if strings.Contains(cfg.DSN, ":memory:") {
		if err := store.Init(); err != nil {
			return err
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
