// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package corim_store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/veraison/corim-store/pkg/model"
	corimstore "github.com/veraison/corim-store/pkg/store"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	handler "github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/provisioning/lifecycle"
	vtscoserv "github.com/veraison/services/vts/coserv"
	"go.uber.org/zap"
)

const (
	PluginName = "corim-store"
)

// implement the IEndorsementStore interface for corimstore.Store
type Store struct {
	Store     *corimstore.Store
	logger    *zap.SugaredLogger
	CoservCfg *vtscoserv.StoreConfig
}

func NewStore() *Store {
	logger := log.Named(PluginName)
	return &Store{nil, logger, nil}
}

func (s *Store) GetKeyTriples(env *comid.Environment, label string, exact bool) ([]*comid.KeyTriple, error) {
	res, err := s.Store.GetActiveKeyTriples(env, label, exact)
	if errors.Is(err, corimstore.ErrNoMatch) {
		return nil, handler.ErrNotFound
	}
	return res, err
}

func (s *Store) GetValueTriples(env *comid.Environment, label string, exact bool) ([]*comid.ValueTriple, error) {
	res, err := s.Store.GetActiveValueTriples(env, label, exact)
	if errors.Is(err, corimstore.ErrNoMatch) {
		return nil, handler.ErrNotFound
	}
	return res, err
}

func (s *Store) ExecuteCoservQuery(profile, query string) (*coserv.Coserv, error) {
	// If reading CoSERV config failed during initialization,
	// CoSERV interface would be disabled.
	if s.CoservCfg == nil {
		panic("received CoSERV request when CoSERV API is disabled")
	}
	s.logger.Infof("got coserv query: %v", query)
	coservService := corimstore.NewCoSERVService(s.Store, s.CoservCfg.Authority, s.CoservCfg.MaxExpiry)
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

func (s *Store) AddCorimBytes(data []byte, label string, activate bool) error {
	s.logger.Debugf("adding CoRIM")
	return s.Store.AddBytes(data, label, activate)
}

func (s *Store) SetEndorsementsState(label string, req []byte, state bool) error {
	s.logger.Debugw("SetEndorsementsState", "set-active", state, "tenant-id", label)

	var elmq lifecycle.Query
	if err := elmq.FromCBOR(req); err != nil {
		return handler.ErrBadRequest
	}

	if err := elmq.Valid(); err != nil {
		return handler.ErrBadRequest
	}

	switch {
	case elmq.EnvironmentSelector != nil:
		if err := s.setEndorsementsStateUsingEnv(&elmq, state); err != nil {
			return err
		}
	case elmq.RimSelector != nil:
		if err := s.setEndorsementsStateUsingIDs(&elmq, state); err != nil {
			return err
		}
	default:
		// should not get here -- query already validated
		panic("invalid ELM query")
	}

	return nil

}

func (o *Store) setEndorsementsStateUsingEnv(query *lifecycle.Query, setActive bool) error {
	if err := query.Valid(); err != nil {
		return err
	}

	var (
		err                   error
		valueTripleQueryGroup *corimstore.ValueTripleQueryGroup
		keyTripleQueryGroup   *corimstore.KeyTripleQueryGroup
	)

	switch *query.ArtifactType {
	case coserv.ArtifactTypeReferenceValues:
		valueTripleQueryGroup, err = valueTripleQueryGroupFromEnvironmentSelector(query.EnvironmentSelector)
		if err != nil {
			return err
		}

		if _, err := query.Profile.Get(); err == nil {
			valueTripleQueryGroup.ForEach(func(v *corimstore.ValueTripleQuery) {
				v.ProfileFromEAT(query.Profile)
			})
		}
	case coserv.ArtifactTypeTrustAnchors:
		keyTripleQueryGroup, err = keyTripleQueryGroupFromEnvironmentSelector(query.EnvironmentSelector)
		if err != nil {
			return err
		}

		if _, err := query.Profile.Get(); err == nil {
			keyTripleQueryGroup.ForEach(func(k *corimstore.KeyTripleQuery) {
				k.ProfileFromEAT(query.Profile)
			})
		}
	default:
		return errors.New("only reference values and trust anchors are supported at present")
	}

	return o.setTripleQueryGroupsActive(valueTripleQueryGroup, keyTripleQueryGroup, setActive)
}

func (o *Store) setEndorsementsStateUsingIDs(query *lifecycle.Query, setActive bool) error {
	if err := query.Valid(); err != nil {
		return err
	}

	valueTripleQueryGroup, err := valueTripleQueryGroupFromRimSelector(query.RimSelector)
	if err != nil {
		return err
	}

	keyTripleQueryGroup, err := keyTripleQueryGroupFromRimSelector(query.RimSelector)
	if err != nil {
		return err
	}

	return o.setTripleQueryGroupsActive(valueTripleQueryGroup, keyTripleQueryGroup, setActive)
}

func valueTripleQueryGroupFromEnvironmentSelector(selector *coserv.EnvironmentSelector) (*corimstore.ValueTripleQueryGroup, error) { //nolint:dupl
	valueTripleQueryGroup := corimstore.NewValueTripleQueryGroup()

	if selector.Classes != nil {
		for i, statefulClass := range *selector.Classes {
			q, err := corimstore.ValueTripleQueryFromStatefulClass(&statefulClass)
			if err != nil {
				return nil, fmt.Errorf("stateful class %d: %w", i, err)
			}

			q.TripleType(model.ReferenceValueTriple).ValidOn(time.Now())

			valueTripleQueryGroup.Add(q)
		}
	}

	if selector.Instances != nil {
		for i, statefulInstance := range *selector.Instances {
			q, err := corimstore.ValueTripleQueryFromStatefulInstance(&statefulInstance)
			if err != nil {
				return nil, fmt.Errorf("stateful instance: %d: %w", i, err)
			}

			q.TripleType(model.ReferenceValueTriple).ValidOn(time.Now())

			valueTripleQueryGroup.Add(q)
		}
	}

	if selector.Groups != nil {
		for i, statefulGroup := range *selector.Groups {
			q, err := corimstore.ValueTripleQueryFromStatefulGroup(&statefulGroup)
			if err != nil {
				return nil, fmt.Errorf("stateful group: %d: %w", i, err)
			}

			q.TripleType(model.ReferenceValueTriple).ValidOn(time.Now())

			valueTripleQueryGroup.Add(q)
		}
	}

	return valueTripleQueryGroup, nil
}

func keyTripleQueryGroupFromEnvironmentSelector(selector *coserv.EnvironmentSelector) (*corimstore.KeyTripleQueryGroup, error) { // nolint:dupl
	keyTripleQueryGroup := corimstore.NewKeyTripleQueryGroup()

	if selector.Classes != nil {
		for i, statefulClass := range *selector.Classes {
			q, err := corimstore.KeyTripleQueryFromStatefulClass(&statefulClass)
			if err != nil {
				return nil, fmt.Errorf("stateful class %d: %w", i, err)
			}

			q.TripleType(model.AttestKeyTriple).ValidOn(time.Now())

			keyTripleQueryGroup.Add(q)
		}
	}

	if selector.Instances != nil {
		for i, statefulInstance := range *selector.Instances {
			q, err := corimstore.KeyTripleQueryFromStatefulInstance(&statefulInstance)
			if err != nil {
				return nil, fmt.Errorf("stateful instance: %d: %w", i, err)
			}

			q.TripleType(model.AttestKeyTriple).ValidOn(time.Now())

			keyTripleQueryGroup.Add(q)
		}
	}

	if selector.Groups != nil {
		for i, statefulGroup := range *selector.Groups {
			q, err := corimstore.KeyTripleQueryFromStatefulGroup(&statefulGroup)
			if err != nil {
				return nil, fmt.Errorf("stateful group: %d: %w", i, err)
			}

			q.TripleType(model.AttestKeyTriple).ValidOn(time.Now())

			keyTripleQueryGroup.Add(q)
		}
	}

	return keyTripleQueryGroup, nil
}

func valueTripleQueryGroupFromRimSelector(selector *coserv.RimSelectorIDs) (*corimstore.ValueTripleQueryGroup, error) {
	if selector == nil {
		return nil, fmt.Errorf("empty RIM selector")
	}

	queryGroup := corimstore.NewValueTripleQueryGroup()

	for _, id := range *selector {
		q := corimstore.NewValueTripleQuery()

		tagId := id.TagID.String()
		// swid.TagID does not expose the underlying type in any way, but we
		// need it to correctly reconstruct it inside corim-store, so we guess
		// by seeing if the ID parses as a valid UUID.
		// see: https://github.com/veraison/corim-store/blob/d1c9b858ead512ff45ba54d0e16781fafcff242c/pkg/model/manifest.go#L94
		var tagIdType model.TagIDType
		if _, err := uuid.Parse(tagId); err == nil {
			tagIdType = model.UUIDTagID
		} else {
			tagIdType = model.StringTagID
		}

		log.Debugf("tag id: %v, tag id type: %v", tagId, tagIdType)

		switch id.Type {
		case coserv.RimSelectorTypeComid:
			q.ModuleTagID(tagIdType, tagId)
		case coserv.RimSelectorTypeCoswid:
			// not supported by corim-store
			// see: https://github.com/veraison/corim-store/blob/d1c9b858ead512ff45ba54d0e16781fafcff242c/pkg/store/coserv.go#L333
			return nil, errors.New("CoSWID selectors not supported")
		case coserv.RimSelectorTypeCorim:
			q.ManifestID(tagIdType, tagId)
		}

		q.TripleType(model.ReferenceValueTriple).ValidOn(time.Now())

		queryGroup.Add(q)
	}

	return queryGroup, nil
}

func keyTripleQueryGroupFromRimSelector(selector *coserv.RimSelectorIDs) (*corimstore.KeyTripleQueryGroup, error) {
	if selector == nil {
		return nil, fmt.Errorf("empty RIM ID selector")
	}

	queryGroup := corimstore.NewKeyTripleQueryGroup()

	for _, id := range *selector {
		q := corimstore.NewKeyTripleQuery()

		switch id.Type {
		case coserv.RimSelectorTypeComid:
			q.ModuleTagIDValue(id.TagID.String())
		case coserv.RimSelectorTypeCoswid:
			// not supported by corim-store
			// see: https://github.com/veraison/corim-store/blob/d1c9b858ead512ff45ba54d0e16781fafcff242c/pkg/store/coserv.go#L333
			return nil, errors.New("CoSWID selectors not supported")
		case coserv.RimSelectorTypeCorim:
			q.ManifestIDValue(id.TagID.String())
		}

		q.TripleType(model.AttestKeyTriple).ValidOn(time.Now())

		queryGroup.Add(q)
	}

	return queryGroup, nil
}

func (o *Store) setTripleQueryGroupsActive(
	valueTripleQueryGroup *corimstore.ValueTripleQueryGroup,
	keyTripleQueryGroup *corimstore.KeyTripleQueryGroup,
	setActive bool,
) error {
	txStore, err := o.Store.BeginTx(nil)
	if err != nil {
		return err
	}

	if valueTripleQueryGroup != nil {
		if _, err := txStore.SetValueTriplesActive(valueTripleQueryGroup, setActive); err != nil {
			_ = txStore.Tx().Rollback()
			return err
		}
	}

	if keyTripleQueryGroup != nil {
		if _, err := txStore.SetKeyTriplesActive(keyTripleQueryGroup, setActive); err != nil {
			_ = txStore.Tx().Rollback()
			return err
		}
	}

	return txStore.Tx().Commit()
}

func (s *Store) Fini() error {
	s.logger.Info("closing corimstore")
	if s.Store == nil {
		panic("attempted to close an uninitialized store")
	}
	if err := s.Store.Close(); err != nil {
		s.logger.Errorf("Failed to close corim-store: %v", err)
		return err
	}
	return nil
}

func (s *Store) Init(params *plugin.Parameters) error {
	s.logger.Debug("initializing default store")
	if params == nil {
		return errors.New("parameters are required for corimstore")
	}
	cfg, err := ConfigFromParameters(params, s.logger)
	if err != nil {
		s.logger.Errorf("Failed to load corim-store parameters: %v", err)
		return err
	}
	s.CoservCfg = cfg.CoservConfig()

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

func (s *Store) GetName() string {
	return PluginName
}

func (s *Store) GetAttestationScheme() string {
	// FIXME(dhanus): This store will in principle work for any schemes that
	// support CoRIM endorsements. So returning a single scheme name does
	// not make sense. The plugin interface must be updated to support plugins
	// that are associated with multiple schemes.
	return ""
}

func (s *Store) GetSupportedMediaTypes() map[string][]string {
	// FIXME(dhanus): This method does not make much sense for store plugin,
	// but something similar is required to let VTS know the CoSERV profiles
	// that are supported. The plugin interface must be updated to somehow
	// incorporate this.
	return nil
}
