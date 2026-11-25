// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package endorsement

import (
	"errors"
	"fmt"
	"iter"
	"slices"

	"github.com/veraison/corim/comid"
	"github.com/veraison/services/pkg/fault"
	"github.com/veraison/services/pkg/plugin"
	"go.uber.org/zap"
)

type Store struct {
	config         *StoreConfig
	logger         *zap.SugaredLogger
	handlersByName map[string]IHandler
}

func NewStore(config *StoreConfig, loader plugin.ILoader, logger *zap.SugaredLogger) (*Store, error) {
	if config == nil || loader == nil || logger == nil {
		panic("nil params")
	}

	ret := &Store{
		config:         config,
		logger:         logger,
		handlersByName: make(map[string]IHandler),
	}

	if err := ret.updateFromLoader(loader); err != nil {
		return nil, err
	}

	return ret, nil
}

func (o *Store) GetKeyTriples( // nolint:dupl
	env *comid.Environment,
	scheme string,
	exact bool,
) ([]*comid.KeyTriple, error) {
	o.logger.Debugf("searching key triples for environment %+v", env)
	for name, handler := range o.IterHandlers() {
		t, err := handler.GetKeyTriples(env, scheme, exact)
		if err == nil {
			o.logger.Debugf("found key triples in store %q", name)
			return t, nil
		}

		if filterError(err, o.logger, name) != nil {
			return nil, err
		}
	}

	return nil, fault.ErrNotFound
}

func (o *Store) GetValueTriples( // nolint:dupl
	env *comid.Environment,
	scheme string,
	exact bool,
) ([]*comid.ValueTriple, error) {
	o.logger.Debugw("searching value triples for", "env", env)
	for name, handler := range o.IterHandlers() {
		t, err := handler.GetValueTriples(env, scheme, exact)
		if err == nil {
			o.logger.Debugf("found value triples in store %q", name)
			return t, nil
		}

		if filterError(err, o.logger, name) != nil {
			return nil, err
		}
	}
	return nil, fault.ErrNotFound
}

func (o *Store) ExecuteCoservQuery(profile, query string, shouldSign bool) ([]byte, error) {
	o.logger.Debugw("coserv query", "profile", profile, "query", query)
	for name, handler := range o.IterHandlers() {
		result, err := handler.ExecuteCoservQuery(profile, query)
		if err != nil {
			if filterError(err, o.logger, name) != nil {
				return nil, err
			}
		} else {
			o.logger.Infof("CoSERV results found in store %q", name)
			if shouldSign {
				return o.config.Signer.Sign(result)
			}
			return result.ToCBOR()
		}
	}
	return nil, fault.ErrNotFound
}

func (o *Store) AddCorimBytes(data []byte, scheme string, activate bool) error {
	if err := o.Primary().AddCorimBytes(data, scheme, activate); err != nil {
		o.logger.Errorf(
			"failed to add endorsements to store %q: %s",
			o.Primary().GetName(),
			err.Error(),
		)
		return err
	}
	return nil
}

func (o *Store) SetEndorsementsState(label string, request []byte, setActive bool) error {
	if err := o.Primary().SetEndorsementsState(label, request, setActive); err != nil {
		o.logger.Errorf(
			"failed to change endorsement state (setActive=%v) in store %q: %s",
			setActive,
			o.Primary().GetName(),
			err.Error(),
		)
		return err
	}
	return nil
}

// IterHandlers iterates over active store handlers in the same order as
// specified by active-plugins setting.
func (o *Store) IterHandlers() iter.Seq2[string, IHandler] {
	return func(yield func(string, IHandler) bool) {
		for _, handlerName := range o.config.ActivePlugins {
			if !yield(handlerName, o.handlersByName[handlerName]) {
				break
			}
		}
	}
}

// Primary returns the first ative store handler. This handler supports writes.
func (o *Store) Primary() IHandler {
	return o.handlersByName[o.config.ActivePlugins[0]]
}

// CoservSigner returns the CoSERV signer used by this store.
func (o *Store) CoservSigner() ISigner {
	return o.config.Signer
}

func (o *Store) Close() {
	for name, handler := range o.handlersByName {
		if err := handler.Fini(); err != nil {
			o.logger.Errorf("closing %q: %s", name, err.Error())
		}
	}
}

func (o *Store) updateFromLoader(loader plugin.ILoader) error {
	if len(o.config.ActivePlugins) == 0 {
		return errors.New("must specify at least one active store plugin")
	}

	activated := make(map[string]bool, len(o.config.ActivePlugins))
	for _, loadedPlugin := range loader.Plugins() {
		pluggable, ok := loadedPlugin.Pluggables()[StorePluginType]
		if !ok {
			continue
		}

		handlerName := pluggable.GetName()
		if !slices.Contains(o.config.ActivePlugins, handlerName) {
			o.logger.Debugf("unloading inactive %s", handlerName)
			if err := loader.Unload(loadedPlugin, StorePluginType); err != nil {
				o.logger.Error(err)
			}
			continue
		}

		handler, ok := pluggable.(IHandler)
		if !ok {
			panic(fmt.Sprintf(
				"%s in %s does not implement scheme.IHander",
				StorePluginType,
				loadedPlugin.Source(),
			))
		}

		if handlerName == o.config.ActivePlugins[0] {
			if !handler.Can(GetTriples, UpdateStore) {
				return fmt.Errorf(
					"primary store %q must support updating and getting triples",
					handlerName,
				)
			}
		}

		if _, ok := o.handlersByName[handlerName]; ok {
			return fmt.Errorf(
				"multiple plugins provide endorsement store %s",
				handlerName,
			)
		}

		params, ok := o.config.PluginParameters[handlerName]
		if !ok {
			params = plugin.NewParameters()
		}

		if err := handler.Init(params); err != nil {
			return fmt.Errorf(
				"could not initialize %s: %w",
				handlerName,
				err,
			)
		}

		o.logger.Debugf("adding %s from %s", handlerName, loadedPlugin.Source())
		o.handlersByName[handlerName] = handler
		activated[handlerName] = true
	}

	for _, toActivate := range o.config.ActivePlugins {
		if _, ok := activated[toActivate]; !ok {
			return fmt.Errorf("no implementation found for active-plugins entry %q", toActivate)
		}
	}

	return nil
}

func filterError(err error, logger *zap.SugaredLogger, name string) error { // nolint: gocritic
	if err == nil {
		logger.Debugf("found in store %q", name)
		return nil
	}

	if errors.Is(err, fault.ErrNotFound) {
		logger.Debugf("not found in store `%s'", name)
		return nil
	}

	if errors.Is(err, fault.ErrNotSupported) {
		logger.Debugf("store %q does not support operation", name)
		return nil
	}

	logger.Errorf("Failed to fetch from store %q: %v", name, err)
	return err
}
