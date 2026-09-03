// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package endorsementstore

import (
	"errors"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/plugin"
	"go.uber.org/zap"
)

type StoreManager = plugin.IManager[handler.IEndorsementStorePlugin]

var ErrNoStores = errors.New("no store plugins active")

type VtsEndorsementStore struct {
	logger *zap.SugaredLogger
	stores []handler.IEndorsementStorePlugin
}

func (s *VtsEndorsementStore) addStore(store handler.IEndorsementStorePlugin) {
	if s.stores == nil {
		s.stores = []handler.IEndorsementStorePlugin{store}
	} else {
		s.stores = append(s.stores, store)
	}
}

func (s *VtsEndorsementStore) storeList() ([]handler.IEndorsementStorePlugin, error) {
	if len(s.stores) == 0 {
		return nil, ErrNoStores
	}
	return s.stores, nil
}

func CreateEndorsementStore(plugins []string, manager StoreManager, logger *zap.SugaredLogger) (handler.IEndorsementStore, error) {
	if len(plugins) == 0 {
		err := errors.New("no plugins in `active-stores` list")
		logger.Errorf("could not create endorsementstore: %v", err)
		return nil, err
	}
	store := VtsEndorsementStore{
		logger: logger,
	}
	for _, pl := range plugins {
		st, err := manager.LookupByName(pl)
		if err != nil {
			logger.Errorf("failed to load store `%s': %v", pl, err)
			return nil, err
		}
		store.addStore(st)
	}
	return &store, nil
}

func (s *VtsEndorsementStore) GetKeyTriples(env *comid.Environment, scheme string, exact bool) ([]*comid.KeyTriple, error) {
	s.logger.Debugw("searching key triples for", "env", env)
	errs := make(map[error]bool, 2) // keep track of kind of errors encountered
	stores, err := s.storeList()
	if err != nil {
		s.logger.Errorf("failed to fetch store list: %v", err)
		return nil, err
	}
	for _, store := range stores {
		name := store.GetName()
		t, err := store.GetKeyTriples(env, scheme, exact)
		if err != nil {
			if logOrErr(err, s.logger, name, &errs) != nil {
				return nil, err
			}
		} else {
			s.logger.Infof("found key triples in store `%s'", name)
			return t, nil
		}
	}
	return nil, computeStoreErr(errs)
}

func (s *VtsEndorsementStore) GetValueTriples(env *comid.Environment, scheme string, exact bool) ([]*comid.ValueTriple, error) {
	s.logger.Debugw("searching value triples for", "env", env)
	errs := make(map[error]bool, 2) // keep track of kind of errors encountered
	stores, err := s.storeList()
	if err != nil {
		s.logger.Errorf("failed to fetch store list: %v", err)
		return nil, err
	}
	for _, store := range stores {
		name := store.GetName()
		t, err := store.GetValueTriples(env, scheme, exact)
		if err != nil {
			if logOrErr(err, s.logger, name, &errs) != nil {
				return nil, err
			}
		} else {
			s.logger.Infof("found value triples in store `%s'", name)
			return t, nil
		}
	}
	return nil, computeStoreErr(errs)
}

func (s *VtsEndorsementStore) ExecuteCoservQuery(profile, query string) (*coserv.Coserv, error) {
	s.logger.Debugw("coserv query", "profile", profile, "query", query)
	errs := make(map[error]bool, 2) // keep track of kind of errors encountered
	stores, err := s.storeList()
	if err != nil {
		s.logger.Errorf("failed to fetch store list: %v", err)
		return nil, err
	}
	for _, store := range stores {
		name := store.GetName()
		res, err := store.ExecuteCoservQuery(profile, query)
		if err != nil {
			if logOrErr(err, s.logger, name, &errs) != nil {
				return nil, err
			}
		} else {
			s.logger.Infof("CoSERV results found in store `%s'", name)
			return res, nil
		}
	}
	return nil, computeStoreErr(errs)
}

func (s *VtsEndorsementStore) AddCorimBytes(data []byte, scheme string, activate bool) error {
	// only submit endorsements to the first store in list
	stores, err := s.storeList()
	if err != nil {
		s.logger.Errorf("failed to fetch store list: %v", err)
		return err
	}
	store := stores[0] // stores is guaranteed to contain at least one entry
	name := store.GetName()
	if err := store.AddCorimBytes(data, scheme, activate); err != nil {
		s.logger.Errorf("failed to add endorsements to store `%s': %v", name, err)
		return err
	}
	return nil
}

func logOrErr(err error, logger *zap.SugaredLogger, name string, errs *map[error]bool) error {
	if err != nil {
		// errors.Is can be used because the plugin rpc client parses the error
		// before returning.
		if errors.Is(err, handler.ErrNotFound) {
			(*errs)[handler.ErrNotFound] = true
			logger.Debugf("not found in store `%s'", name)
			return nil
		}
		if errors.Is(err, handler.ErrUnsupported) {
			(*errs)[handler.ErrUnsupported] = true
			logger.Debugf("store `%s' does not support operation", name)
			return nil
		}
		logger.Errorf("Failed to fetch from store `%s': %v", name, err)
		return err
	} // else
	return nil
}

func computeStoreErr(errs map[error]bool) error {
	// if ErrNotFound is encountered at least once, it means there
	// is at least one store that supports the operation.
	if _, ok := errs[handler.ErrNotFound]; ok {
		return handler.ErrNotFound
	}
	// got ErrUnsupported from all stores
	return handler.ErrUnsupported
}
