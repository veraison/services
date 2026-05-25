// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package trustedservices

import (
	"errors"
	"strings"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/plugin"
	"go.uber.org/zap"
)

type StoreManager = plugin.IManager[handler.IEndorsementStorePlugin]

type Store struct {
	logger   *zap.SugaredLogger
	primary  handler.IEndorsementStorePlugin
	fallback []handler.IEndorsementStorePlugin
}

func (s *Store) storeList() []handler.IEndorsementStorePlugin {
	stores := []handler.IEndorsementStorePlugin{s.primary}
	stores = append(stores, s.fallback...)
	return stores
}

func loadPrimaryStore(manager StoreManager) (handler.IEndorsementStorePlugin, error) {
	primaryStore, err := manager.LookupByName("corim-store")
	if err != nil {
		return nil, err
	}
	return primaryStore, nil
}

func CreateStore(
	label string,
	manager StoreManager,
	logger *zap.SugaredLogger) (handler.IEndorsementStore, error) {
	ps, err := loadPrimaryStore(manager)
	if err != nil {
		logger.Warnf("could not load primary store: %s", err)
		return nil, errors.New("could not load corimstore")
	}
	logger.Debug("loaded primary store")

	st := Store{
		logger:   logger,
		primary:  ps,
		fallback: nil,
	}
	b, a, found := strings.Cut(label, "/")
	if found {
		label = a
	} else {
		label = b
	}
	pl, err := manager.LookupByAttestationScheme(label)
	if err != nil {
		logger.Infof("no fallback stores found with label: %s", label)
	} else {
		st.fallback = []handler.IEndorsementStorePlugin{pl}
	}
	return &st, nil
}

func CreateStoreFromMediaType(
	mt string,
	manager StoreManager,
	logger *zap.SugaredLogger) (handler.IEndorsementStore, error) {
	ps, err := loadPrimaryStore(manager)
	if err != nil {
		logger.Warnf("could not load primary store: %s", err)
		return nil, errors.New("could not load corimstore")
	}
	logger.Debug("loaded primary store")
	st := Store{
		logger:   logger,
		primary:  ps,
		fallback: nil,
	}
	pl, err := manager.LookupByMediaType(mt)
	if err != nil {
		logger.Infof("no fallback stores found with media-type: %s", mt)
	} else {
		st.fallback = []handler.IEndorsementStorePlugin{pl}
	}
	return &st, nil

}

func (s *Store) GetKeyTriples(env *comid.Environment, scheme string, exact bool) ([]*comid.KeyTriple, error) {
	s.logger.Debugw("searching key triples for", "env", env)
	stores := s.storeList()
	for _, store := range stores {
		name := store.GetName()
		t, err := store.GetKeyTriples(env, scheme, exact)
		if err != nil {
			if !errors.Is(err, handler.ErrNotFound) {
				s.logger.Warnf("error from store `%s' while fetching key triples: %v", name, err)
			} else {
				s.logger.Debugf("key triples not found in store `%s'", name)
			}
		} else {
			s.logger.Infof("found key triples in store %s", name)
			return t, nil
		}
	}
	return nil, handler.ErrNotFound
}

func (s *Store) GetValueTriples(env *comid.Environment, scheme string, exact bool) ([]*comid.ValueTriple, error) {
	s.logger.Debugw("searching value triples for", "env", env)
	stores := s.storeList()
	for _, store := range stores {
		name := store.GetName()
		t, err := store.GetValueTriples(env, scheme, exact)
		if err != nil {
			if !errors.Is(err, handler.ErrNotFound) {
				s.logger.Warnf("error from store `%s' while fetching value triples: %v", name, err)
			} else {
				s.logger.Debugf("value triples not found in store `%s'", name)
			}
		} else {
			s.logger.Infof("found value triples in store `%s'", name)
			return t, nil
		}
	}
	return nil, handler.ErrNotFound
}

func (s *Store) ExecuteCoservQuery(mediaType, query string) (*coserv.Coserv, error) {
	s.logger.Debugw("coserv query", "media-type", mediaType, "query", query)
	stores := s.storeList()
	for _, store := range stores {
		name := store.GetName()
		res, err := store.ExecuteCoservQuery(mediaType, query)
		if err != nil {
			if !errors.Is(err, handler.ErrNotFound) {
				s.logger.Warnf("error from store `%s' while executing CoSERV query: %v", name, err)
			} else {
				s.logger.Debugf("CoSERV results not found in store `%s'", name)
			}
		} else {
			s.logger.Info("CoSERV results found in store `%s'", name)
			return res, nil
		}
	}
	return nil, handler.ErrNotFound
}

func (s *Store) AddCorimBytes(data []byte, scheme string, activate bool) error {
	return s.primary.AddCorimBytes(data, scheme, activate)
}

func (s *Store) Close() error {
	// the composite store should not be closed using
	// the Close method. Instead the primary store and
	// secondary store managers must be closed separately.
	// This is done during service shutdown.
	return nil
}

func CloseCorimStore(manager StoreManager) error {
	// Note: the Close method should ideally be called by the plugin
	// manager. Since there is no `Finalize' method in IPluggable,
	// this method is implemented to close the stores (at present
	// only corimstore has non-trivial close mehtod).
	st, err := loadPrimaryStore(manager)
	if err != nil {
		return err
	}
	return st.Close()
}
